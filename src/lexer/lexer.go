package lexer

import (
	"fmt"
	"lizalang/token"
	"lizalang/utils"
	"slices"
	"strconv"
)

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

func (lexer *Lexer) Advance() {
	lexer.Pos++
	lexer.CurChar = lexer.NextChar
	if lexer.Pos+1 < len(lexer.Code) {
		lexer.NextChar = lexer.Code[lexer.Pos+1]
	}
}

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
	if len(code) > 2 {
		lexer.CurChar = []rune(code)[0]
		lexer.NextChar = []rune(code)[1]
	}

	return lexer
}

func (lexer *Lexer) Lex() {
	for {
		switch {
		case lexer.CurChar == 0:
			lexer.NewToken(token.EOF, 0)
			break
		case utils.IsLetter(lexer.CurChar):
			lexer.handleIdOrKeyword()
			break
		case utils.IsDigit(lexer.CurChar):
			err := lexer.handleNumber()
			if err != nil {
				lexer.Errors = append(lexer.Errors, err)
			}
			break
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
			err := lexer.handleString()
			if err != nil {
				lexer.Errors = append(lexer.Errors, err)
			}
			break
		case lexer.CurChar == '\n':
			lexer.NewToken(token.NewInstruction, lexer.CurChar)
			lexer.Line++
			break
		case lexer.CurChar == ' ' || lexer.CurChar == '\t':
			break
		case lexer.CurChar == '/' && lexer.NextChar == '/':
			for lexer.NextChar != '\n' {
				lexer.Advance()
			}
			break
		case lexer.CurChar == '/' && lexer.NextChar == '*':
			for lexer.CurChar == '*' && lexer.NextChar == '/' {
				lexer.Advance()
			}
			lexer.Advance()
			break
		default:
			if oneCharToken, ok := token.SymbolMap[lexer.CurChar]; ok {
				lexer.NewToken(oneCharToken, string(lexer.CurChar))
				break
			}
			lexer.NewToken(token.Invalid, lexer.CurChar)
			lexer.Errors = append(lexer.Errors, fmt.Errorf("Invalid character %s on a line %d", string(lexer.CurChar), lexer.Line))
			break
		}
		if lexer.Tokens[len(lexer.Tokens)-1].Type == token.EOF {
			break
		}
		lexer.Advance()

	}
}

func (lexer *Lexer) NewToken(tokentype token.TokenType, value any) {
	tok := token.Token{
		tokentype,
		value,
		lexer.Line,
		lexer.File,
	}
	lexer.Tokens = append(lexer.Tokens, tok)
}

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
	if slices.Contains(token.KeyWords, str) {
		lexer.NewToken(token.Keyword, str)
	} else {
		lexer.NewToken(token.Identifier, str)
	}
}

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
		if floatNum, err := strconv.ParseFloat(num, 64); err == nil {
			lexer.NewToken(token.Float, floatNum)
		} else {
			lexer.NewToken(token.Invalid, "NaN")
			return fmt.Errorf("Error at line %d: %s is not a number", lexer.Line, string(num))
		}
	} else {
		intNum, err := strconv.ParseInt(num, 10, 64)
		if err != nil {
			return fmt.Errorf("Error at line %d: %s is not a number", lexer.Line, string(num))
		}
		lexer.NewToken(token.Int, intNum)
	}
	return nil
}

func (lexer *Lexer) handleString() error {
	lexer.Advance()
	str := ""
	originalLine := lexer.Line
	for lexer.CurChar != '"' {
		if lexer.CurChar == '\\' {
			lexer.Advance()
			str = string(append([]rune(str), utils.EscapeSeq[lexer.CurChar]))
		} else {
			str = string(append([]rune(str), lexer.CurChar))
		}
		if lexer.CurChar == 0 {
			return fmt.Errorf("Error at line %d: missing \"", originalLine)
		}
		if lexer.CurChar == '\n' {
			lexer.Line++
		}
		lexer.Advance()
	}
	lexer.NewToken(token.String, str)
	return nil
}
