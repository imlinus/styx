# Styx

A lightweight programming language for the web.

Styx is a hobby project born from my love of the stateless, "save-and-refresh" execution model of PHP, but I prefer the clean syntax of Lua or Python. I built this to have a fun, isolated environment for my own back-ends.

## Why I Built This

- **The PHP Workflow:** I want to edit a file, save it, and refresh the browser to see the result instantly. No build steps, no server restarts.
- **Clean-ish Syntax:** I like `function/end` and `if/then` blocks. No more `$variable` everywhere or weird `->` arrows.
- **Modern Comforts:** I added `let` for declarations and backtick template literals `${interpolation}` because I use them too much in JS to live without them.
- **Web First:** `http`, `json`, `time`, and `env` are global primitives. It's built specifically to handle requests and return data.

## Getting It

```bash
curl -L https://github.com/imlinus/styx/releases/latest/download/styx -o styx
chmod +x styx
```

## Quick Example

I've provided a few examples in the `examples` folder.

```bash
./styx examples/basics.sx
```
