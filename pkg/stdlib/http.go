package stdlib

import (
	"encoding/json"
	"io"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"styx/pkg/object"
)

func GetHttpModule(r *http.Request, w http.ResponseWriter, basePath string) *object.Hash {
	// GET parameters
	getPairs := make(map[string]object.HashPair)
	for key, values := range r.URL.Query() {
		getPairs[(&object.String{Value: key}).HashKey()] = object.HashPair{
			Key:   &object.String{Value: key},
			Value: &object.String{Value: values[0]}, // Just the first value for now
		}
	}
	getTable := &object.Hash{Pairs: getPairs}

	// POST parameters (JSON body)
	postPairs := make(map[string]object.HashPair)
	if r.Method == http.MethodPost {
		body, err := io.ReadAll(r.Body)
		if err == nil {
			var data map[string]interface{}
			if err := json.Unmarshal(body, &data); err == nil {
				for key, val := range data {
					objVal := object.NativeToStyxObject(val)
					postPairs[(&object.String{Value: key}).HashKey()] = object.HashPair{
						Key:   &object.String{Value: key},
						Value: objVal,
					}
				}
			}
		}
	}
	postTable := &object.Hash{Pairs: postPairs}

	// HEADERS
	headerPairs := make(map[string]object.HashPair)
	for key, values := range r.Header {
		sKey := &object.String{Value: key}
		headerPairs[sKey.HashKey()] = object.HashPair{
			Key:   sKey,
			Value: &object.String{Value: values[0]},
		}
	}
	// Add Host manually as Go removes it from Header map
	hostKey := &object.String{Value: "Host"}
	headerPairs[hostKey.HashKey()] = object.HashPair{
		Key:   hostKey,
		Value: &object.String{Value: r.Host},
	}
	headersTable := &object.Hash{Pairs: headerPairs}

	// Module Hash
	modulePairs := make(map[string]object.HashPair)

	modulePairs[(&object.String{Value: "get"}).HashKey()] = object.HashPair{
		Key:   &object.String{Value: "get"},
		Value: getTable,
	}

	modulePairs[(&object.String{Value: "post"}).HashKey()] = object.HashPair{
		Key:   &object.String{Value: "post"},
		Value: postTable,
	}

	modulePairs[(&object.String{Value: "headers"}).HashKey()] = object.HashPair{
		Key:   &object.String{Value: "headers"},
		Value: headersTable,
	}

	modulePairs[(&object.String{Value: "method"}).HashKey()] = object.HashPair{
		Key:   &object.String{Value: "method"},
		Value: &object.String{Value: r.Method},
	}

	modulePairs[(&object.String{Value: "url"}).HashKey()] = object.HashPair{
		Key:   &object.String{Value: "url"},
		Value: &object.String{Value: r.URL.String()},
	}

	ip, _, _ := net.SplitHostPort(r.RemoteAddr)
	if ip == "" {
		ip = r.RemoteAddr
	}

	modulePairs[(&object.String{Value: "ip"}).HashKey()] = object.HashPair{
		Key:   &object.String{Value: "ip"},
		Value: &object.String{Value: ip},
	}

	// proxy(targetUrl, headersOverride)
	modulePairs[(&object.String{Value: "proxy"}).HashKey()] = object.HashPair{
		Key: &object.String{Value: "proxy"},
		Value: &object.Builtin{
			Fn: func(args ...object.Object) object.Object {
				if len(args) < 1 || args[0].Type() != object.STRING_OBJ {
					return &object.Error{Message: "proxy expects at least 1 string argument"}
				}
				target := args[0].(*object.String).Value

				proxyReq, err := http.NewRequest(r.Method, target, r.Body)
				if err != nil {
					return &object.Error{Message: "proxy error creating request: " + err.Error()}
				}

				// Copy original headers
				for k, v := range r.Header {
					proxyReq.Header[k] = v
				}

				// Apply overrides if provided
				if len(args) == 2 && args[1].Type() == object.HASH_OBJ {
					override := args[1].(*object.Hash)
					for _, pair := range override.Pairs {
						if k, ok := pair.Key.(*object.String); ok {
							if v, ok := pair.Value.(*object.String); ok {
								proxyReq.Header.Set(k.Value, v.Value)
							}
						}
					}
				}

				client := &http.Client{}
				resp, err := client.Do(proxyReq)
				if err != nil {
					return &object.Error{Message: "proxy error executing request: " + err.Error()}
				}
				defer resp.Body.Close()

				// Copy response headers
				for k, v := range resp.Header {
					for _, vv := range v {
						w.Header().Add(k, vv)
					}
				}
				w.WriteHeader(resp.StatusCode)

				// Copy response body
				io.Copy(w, resp.Body)

				return &object.Null{}
			},
		},
	}

	// status(code)
	modulePairs[(&object.String{Value: "status"}).HashKey()] = object.HashPair{
		Key: &object.String{Value: "status"},
		Value: &object.Builtin{
			Fn: func(args ...object.Object) object.Object {
				if len(args) == 1 && args[0].Type() == object.INTEGER_OBJ {
					w.WriteHeader(int(args[0].(*object.Integer).Value))
				}
				return &object.Null{}
			},
		},
	}

	// header(key, value)
	modulePairs[(&object.String{Value: "header"}).HashKey()] = object.HashPair{
		Key: &object.String{Value: "header"},
		Value: &object.Builtin{
			Fn: func(args ...object.Object) object.Object {
				if len(args) == 2 && args[0].Type() == object.STRING_OBJ && args[1].Type() == object.STRING_OBJ {
					w.Header().Set(args[0].(*object.String).Value, args[1].(*object.String).Value)
				}
				return &object.Null{}
			},
		},
	}

	// file(path)
	modulePairs[(&object.String{Value: "file"}).HashKey()] = object.HashPair{
		Key: &object.String{Value: "file"},
		Value: &object.Builtin{
			Fn: func(args ...object.Object) object.Object {
				if len(args) < 1 || args[0].Type() != object.STRING_OBJ {
					return &object.Error{Message: "http.file expects 1 string argument"}
				}
				path := args[0].(*object.String).Value

				// Make relative paths absolute to the script
				if !filepath.IsAbs(path) {
					path = filepath.Join(basePath, path)
				}

				// Optional: Add basic content type detection
				ext := filepath.Ext(path)
				contentType := "text/plain"
				switch ext {
				case ".js":
					contentType = "application/javascript"
				case ".css":
					contentType = "text/css"
				case ".html":
					contentType = "text/html"
				case ".png":
					contentType = "image/png"
				case ".jpg", ".jpeg":
					contentType = "image/jpeg"
				case ".svg":
					contentType = "image/svg+xml"
				case ".json":
					contentType = "application/json"
				}

				w.Header().Set("Content-Type", contentType)

				f, err := os.Open(path)
				if err != nil {
					return &object.Error{Message: "http.file error: " + err.Error()}
				}
				defer f.Close()

				io.Copy(w, f)
				return &object.Null{}
			},
		},
	}

	return &object.Hash{Pairs: modulePairs}
}
