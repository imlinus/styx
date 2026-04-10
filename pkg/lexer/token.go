package lexer

type TokenType string

type Token struct {
	Type    TokenType
	Literal string
}

const (
	ILLEGAL = "ILLEGAL"
	EOF     = "EOF"

	// Identifiers + literals
	IDENT    = "IDENT"    // let, function, if, ...
	INT      = "INT"      // 1343456
	STRING   = "STRING"   // "foobar"
	TEMPLATE = "TEMPLATE" // `hello ${name}`

	// Operators
	ASSIGN   = "="
	PLUS     = "+"
	MINUS    = "-"
	ASTERISK = "*"
	SLASH    = "/"
	EQ       = "=="
	NOT_EQ   = "!="
	CONCAT   = ".."
	GT       = ">"
	GE       = ">="
	LT       = "<"
	LE       = "<="

	// Delimiters
	COMMA     = ","
	COLON     = ":"
	LPAREN    = "("
	RPAREN    = ")"
	LBRACE    = "{"
	RBRACE    = "}"
	SEMICOLON = ";"
	DOT       = "."
	LBRACKET  = "["
	RBRACKET  = "]"

	// Keywords
	FUNCTION = "FUNCTION"
	LET      = "LET"
	IF       = "IF"
	ELSE     = "ELSE"
	ELSEIF   = "ELSEIF"
	END      = "END"
	RETURN   = "RETURN"
	TRUE     = "TRUE"
	FALSE    = "FALSE"
	THEN     = "THEN"
	IMPORT   = "IMPORT"
	FROM     = "FROM"
	NULL     = "NULL"
	AND      = "AND"
	OR       = "OR"
	LOOP     = "LOOP"
	IN       = "IN"
	BREAK    = "BREAK"
	CONTINUE = "CONTINUE"
)

var keywords = map[string]TokenType{
	"function": FUNCTION,
	"let":      LET,
	"if":       IF,
	"else":     ELSE,
	"elseif":   ELSEIF,
	"end":      END,
	"return":   RETURN,
	"true":     TRUE,
	"false":    FALSE,
	"then":     THEN,
	"import":   IMPORT,
	"from":     FROM,
	"null":     NULL,
	"and":      AND,
	"or":       OR,
	"loop":     LOOP,
	"in":       IN,
	"break":    BREAK,
	"continue": CONTINUE,
}

func LookupIdent(ident string) TokenType {
	if tok, ok := keywords[ident]; ok {
		return tok
	}
	return IDENT
}
