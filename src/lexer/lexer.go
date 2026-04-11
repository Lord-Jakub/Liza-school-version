package lexer

import (
	"fmt"
	"lizalang/token"
	"lizalang/utils"
	"slices"
	"strconv"
)

// Lexer represents a lexical analyzer that converts source code into tokens.
// It maintains current position, characters, and accumulates tokens and errors.
type Lexer struct {
	Code     []rune
	Pos      int
	Line     int
	CurChar  rune
	NextChar rune
	Tokens   []token.Token
	File     string
	Errors   []error
}

// New creates and initializes a new Lexer instance from source code.
func New(code string, file string) *Lexer {
	lexer := &Lexer{
		[]rune(code + string(0)),
		0,
		1,
		0,
		0,
		make([]token.Token, 0),
		file,
		make([]error, 0),
	}

	// Initialize current and next character if code is long enough.
	if len(code) > 2 {
		lexer.CurChar = []rune(code)[0]
		lexer.NextChar = []rune(code)[1]
	}

	return lexer
}

// Advance moves the lexer one character forward,
// updating both current and lookahead characters.
func (lexer *Lexer) Advance() {
	lexer.Pos++
	lexer.CurChar = lexer.NextChar
	if lexer.Pos+1 < len(lexer.Code) {
		lexer.NextChar = lexer.Code[lexer.Pos+1]
	}
}

// Lex performs the main lexing loop, producing tokens from input source.
func (lexer *Lexer) Lex() {
	for lexer.Pos <= len(lexer.Code) {
		switch {
		case lexer.CurChar == 0:
			// End of input
			lexer.NewToken(token.EOF, 0)
			break

		case utils.IsLetter(lexer.CurChar):
			// Identifier or keyword
			lexer.handleIdOrKeyword()
			break

		case utils.IsDigit(lexer.CurChar):
			// Numeric literal
			err := lexer.handleNumber()
			if err != nil {
				lexer.Errors = append(lexer.Errors, err)
			}
			break

		// Multi-character operators
		case lexer.CurChar == '<' && lexer.NextChar == '=':
			lexer.NewToken(token.Operator, "<=")
			lexer.Advance()
			break
		case lexer.CurChar == '>' && lexer.NextChar == '=':
			lexer.NewToken(token.Operator, ">=")
			lexer.Advance()
			break
		case lexer.CurChar == '!' && lexer.NextChar == '=':
			lexer.NewToken(token.Operator, "!=")
			lexer.Advance()
			break
		case lexer.CurChar == '=' && lexer.NextChar == '=':
			lexer.NewToken(token.Operator, "==")
			lexer.Advance()
			break
		case lexer.CurChar == '&' && lexer.NextChar == '&':
			lexer.NewToken(token.Operator, "&&")
			lexer.Advance()
			break
		case lexer.CurChar == '|' && lexer.NextChar == '|':
			lexer.NewToken(token.Operator, "||")
			lexer.Advance()
			break

		case lexer.CurChar == '"':
			// String literal
			err := lexer.handleString()
			if err != nil {
				lexer.Errors = append(lexer.Errors, err)
			}
			break

		case lexer.CurChar == '\n':
			// Newline treated as instruction separator
			lexer.NewToken(token.NewInstruction, lexer.CurChar)
			lexer.Line++
			break

		case lexer.CurChar == ' ' || lexer.CurChar == '\t':
			// Skip whitespace
			break

		case lexer.CurChar == '/' && lexer.NextChar == '/':
			// Single-line comment
			for lexer.NextChar != '\n' {
				lexer.Advance()
			}
			break

		case lexer.CurChar == '/' && lexer.NextChar == '*':
			// Multi-line comment
			for lexer.CurChar == '*' && lexer.NextChar == '/' {
				lexer.Advance()
			}
			lexer.Advance()
			break

		default:
			// Single-character tokens or invalid input
			if oneCharToken, ok := token.SymbolMap[lexer.CurChar]; ok {
				lexer.NewToken(oneCharToken, string(lexer.CurChar))
				break
			}

			// Unknown character
			lexer.NewToken(token.Invalid, lexer.CurChar)
			lexer.Errors = append(lexer.Errors,
				fmt.Errorf("Invalid character %s on a line %d", string(lexer.CurChar), lexer.Line))
			break
		}

		lexer.Advance()
	}

	// Ensure EOF token at the end
	lexer.NewToken(token.EOF, 0)
}

// NewToken creates a new token and appends it to the token stream.
func (lexer *Lexer) NewToken(tokentype token.TokenType, value any) {
	tok := token.Token{
		tokentype,
		value,
		lexer.Line,
		lexer.File,
	}
	lexer.Tokens = append(lexer.Tokens, tok)
}

// handleIdOrKeyword scans identifiers and checks for reserved keywords.
func (lexer *Lexer) handleIdOrKeyword() {
	str := ""
	for utils.IsLetter(lexer.CurChar) || utils.IsDigit(lexer.CurChar) {
		str = string(append([]rune(str), lexer.CurChar))

		if utils.IsLetter(lexer.NextChar) || utils.IsDigit(lexer.NextChar) {
			lexer.Advance()
		} else {
			break
		}
	}

	// Distinguish keyword vs identifier
	if slices.Contains(token.KeyWords, str) {
		lexer.NewToken(token.Keyword, str)
	} else {
		lexer.NewToken(token.Identifier, str)
	}
}

// handleNumber scans integer and floating-point literals.
// Returns an error if parsing fails.
func (lexer *Lexer) handleNumber() error {
	num := ""
	hasDot := false

	for utils.IsDigit(lexer.CurChar) || lexer.CurChar == '.' {
		num = string(append([]rune(num), lexer.CurChar))

		if lexer.CurChar == '.' {
			hasDot = true
		}

		if utils.IsDigit(lexer.NextChar) || lexer.NextChar == '.' {
			lexer.Advance()
		} else {
			break
		}
	}

	if hasDot {
		// Float parsing
		if floatNum, err := strconv.ParseFloat(num, 64); err == nil {
			lexer.NewToken(token.Float, floatNum)
		} else {
			lexer.NewToken(token.Invalid, "NaN")
			return fmt.Errorf("Error at line %d: %s is not a number", lexer.Line, string(num))
		}
	} else {
		// Integer parsing
		intNum, err := strconv.ParseInt(num, 10, 64)
		if err != nil {
			return fmt.Errorf("Error at line %d: %s is not a number", lexer.Line, string(num))
		}
		lexer.NewToken(token.Int, intNum)
	}

	return nil
}

// handleString scans string literals, supporting escape sequences.
// Returns an error if the string is not properly terminated.
func (lexer *Lexer) handleString() error {
	lexer.Advance() // skip opening quote

	str := ""
	originalLine := lexer.Line

	for lexer.CurChar != '"' {
		if lexer.CurChar == '\\' {
			// Handle escape sequence
			lexer.Advance()
			str = string(append([]rune(str), utils.EscapeSeq[lexer.CurChar]))
		} else {
			str = string(append([]rune(str), lexer.CurChar))
		}

		// Unterminated string
		if lexer.CurChar == 0 {
			return fmt.Errorf("Error at line %d: missing \"", originalLine)
		}

		// Track line numbers inside strings
		if lexer.CurChar == '\n' {
			lexer.Line++
		}

		lexer.Advance()
	}

	lexer.NewToken(token.String, str)
	return nil
}
