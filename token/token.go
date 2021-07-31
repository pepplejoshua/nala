package token

// TokenType is an alias for the string type
type TokenType string

// using bytes or int for TokenType would've been more efficient
type Token struct {
	Type    TokenType
	Literal string
}

const (
	ILLEGAL = "ILLEGAL"
	EOF     = "EOF"

	// identifiers and literals
	IDENT  = "IDENT"
	INT    = "INT"
	STRING = "STRING"

	// operators
	ASSIGN     = "="
	PLUS       = "+"
	MINUS      = "-"
	BANG       = "!"
	ASTERISK   = "*"
	SLASH      = "/"
	MODULO     = "%"
	APOSTROPHE = "'"

	LT = "<"
	GT = ">"

	EQ     = "=="
	NOT_EQ = "!="

	// delimiters
	COMMA     = ","
	SEMICOLON = ";"

	LPAREN   = "("
	RPAREN   = ")"
	LBRACE   = "{"
	RBRACE   = "}"
	LBRACKET = "["
	RBRACKET = "]"
	COLON    = ":"
	PIPE     = "|"

	// keywords
	FUNCTION = "FUNCTION"
	LET      = "LET"
	TRUE     = "TRUE"
	FALSE    = "FALSE"
	IF       = "IF"
	ELSE     = "ELSE"
	RETURN   = "RETURN"
	MACRO    = "MACRO"
	CONS     = "CONS"
	LIST     = "LIST"
)

var keywords = map[string]TokenType{
	"fn":     FUNCTION,
	"let":    LET,
	"true":   TRUE,
	"false":  FALSE,
	"if":     IF,
	"else":   ELSE,
	"return": RETURN,
	"macro":  MACRO,
	"cons":   CONS,
	"list":   LIST,
}

func LookupIdent(ident string) TokenType {
	// special if form allowing binding and then checking of boolean condition
	// after semi-colon
	if tok, ok := keywords[ident]; ok {
		return tok
	}
	return IDENT
}
