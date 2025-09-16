package scanner

import (
	"internal/token"
	"internal/util/log"
	"strconv"
	"strings"
	"unicode"
)

type Scanner struct {
	source []rune

	start   int
	current int
	line    int
	col     int

	tokens []token.Token
}

func NewScanner(source string) *Scanner {
	return &Scanner{
		source: []rune(source),
	}
}

func (s *Scanner) ScanTokens() ([]token.Token, []error) {
	errors := make([]error, 0)

	for !s.isAtEnd() {
		s.col += s.current - s.start
		s.start = s.current

		if err := s.scanToken(); err != nil {
			errors = append(errors, err)
		}
	}

	s.start = s.current
	s.addToken(token.EOF, nil)

	return s.tokens, errors
}

func (s *Scanner) scanToken() error {
	c := s.advance()

	switch c {
	case ' ', '\r', '\t':
		// Ignore whitespace.
	case '\n':
		s.line++
		s.col = 0
	case '(':
		s.addToken(token.LEFT_PAREN, nil)
	case ')':
		s.addToken(token.RIGHT_PAREN, nil)
	case '{':
		s.addToken(token.LEFT_BRACE, nil)
	case '}':
		s.addToken(token.RIGHT_BRACE, nil)
	case ',':
		s.addToken(token.COMMA, nil)
	case '-':
		s.addToken(token.MINUS, nil)
	case '+':
		s.addToken(token.PLUS, nil)
	case ':':
		s.addToken(token.COLON, nil)
	case ';':
		s.addToken(token.SEMICOLON, nil)
	case '?':
		s.addToken(token.QUESTION, nil)
	case '/':
		if s.advanceIfMatch('/') {
			for s.peek() != '\n' && !s.isAtEnd() {
				s.advance()
			}

			s.addToken(token.COMMENT, string(s.source[s.start+2:s.current]))
		} else if s.advanceIfMatch('*') {
			for {
				if s.peek() == '\n' {
					s.line++
				}

				if s.isAtEnd() {
					err := NewScanErrorWithLog("Unterminated multi-line comment", s.line, "")
					return err
				}

				if s.advanceIfMatch('*', '/') {
					s.addToken(token.MULTI_COMMENT, string(s.source[s.start+2:s.current-2]))

					break
				}

				s.advance()
			}
		} else {
			s.addToken(token.SLASH, nil)
		}
	case '*':
		s.addToken(token.STAR, nil)
	case '!':
		if s.advanceIfMatch('=') {
			s.addToken(token.BANG_EQUAL, nil)
		} else {
			s.addToken(token.BANG, nil)
		}
	case '=':
		if s.advanceIfMatch('=') {
			s.addToken(token.EQUAL_EQUAL, nil)
		} else {
			s.addToken(token.EQUAL, nil)
		}
	case '<':
		if s.advanceIfMatch('=') {
			s.addToken(token.LESS_EQUAL, nil)
		} else {
			s.addToken(token.LESS, nil)
		}
	case '>':
		if s.advanceIfMatch('=') {
			s.addToken(token.GREATER_EQUAL, nil)
		} else {
			s.addToken(token.GREATER, nil)
		}
	case '"':
		// String literal with escape support (\n, \r, \t, \", \\, \uXXXX)
		var builder strings.Builder

		for !s.isAtEnd() {
			ch := s.advance()
			if ch == '\n' { // raw newline inside string literal
				s.line++
			}
			if ch == '"' { // closing quote
				// finished (we already consumed closing quote; break)
				break
			}

			if ch == '\\' { // escape sequence
				if s.isAtEnd() {
					return NewScanErrorWithLog("Unterminated escape sequence", s.line, "")
				}
				esc := s.advance()
				switch esc {
				case 'n':
					builder.WriteRune('\n')
				case 'r':
					builder.WriteRune('\r')
				case 't':
					builder.WriteRune('\t')
				case '"':
					builder.WriteRune('"')
				case '\\':
					builder.WriteRune('\\')
				case 'u':
					// Expect exactly 4 hex digits.
					hexDigits := make([]rune, 0, 4)
					for i := 0; i < 4; i++ {
						if s.isAtEnd() {
							return NewScanErrorWithLog("Incomplete unicode escape (expect 4 hex digits)", s.line, "")
						}
						h := s.advance()
						if !(h >= '0' && h <= '9' || h >= 'a' && h <= 'f' || h >= 'A' && h <= 'F') {
							return NewScanErrorWithLog("Invalid unicode escape (non-hex digit)", s.line, "")
						}
						hexDigits = append(hexDigits, h)
					}
					code, err := strconv.ParseInt(string(hexDigits), 16, 32)
					if err != nil {
						return NewScanErrorWithLog("Invalid unicode escape: "+err.Error(), s.line, "")
					}
					builder.WriteRune(rune(code))
				default:
					return NewScanErrorWithLog("Unknown escape sequence: \\"+string(esc), s.line, "")
				}
				continue
			}

			builder.WriteRune(ch)
		}

		if s.isAtEnd() && (len(s.source) == 0 || s.source[s.current-1] != '"') {
			return NewScanErrorWithLog("Unterminated string", s.line, "")
		}

		// add token including original lexeme slice; literal is decoded content
		s.addToken(token.STRING, builder.String())
	case '.':
		if !s.isDigit(s.peek()) {
			s.addToken(token.DOT, nil)
			break
		}

		fallthrough // If it's a digit, handle it in the number section.
	default:
		if s.isDigit(c) || c == '.' {
			dotCount := 0

			if c == '.' {
				dotCount = 1
			}

			for s.isDigit(s.peek()) {
				s.advance()
			}

			for s.peek() == '.' && s.isDigit(s.peekNext()) {
				dotCount += 1

				s.advance()
				for s.isDigit(s.peek()) {
					s.advance()
				}
			}

			if dotCount > 1 {
				err := NewScanErrorWithLog("Invalid number format: multiple decimal points", s.line, "")
				return err
			}

			if dotCount == 0 {
				return s.addIntToken()
			} else {
				return s.addRealToken()
			}
		}

		if s.isLetter(c) {
			for s.isLetterDigitMark(s.peek()) {
				s.advance()
			}

			text := string(s.source[s.start:s.current])
			tokenType, ok := token.Keywords[text]
			if !ok {
				tokenType = token.IDENTIFIER
			}

			s.addToken(tokenType, nil)

			return nil
		}

		err := NewScanErrorWithLog("Unexpected character: "+string(c), s.line, "")
		return err
	}

	return nil
}

func (s *Scanner) addToken(t token.TokenType, literal any) {
	text := s.source[s.start:s.current]

	token := token.Token{
		TokenType: t,
		Lexeme:    string(text),
		Literal:   literal,
		Offset: token.Offset{
			Line:  s.line,
			Index: s.col - 1,
		},
	}

	log.Debug("Token", log.S("tokenType", t.String()), log.A("token", token))

	s.tokens = append(s.tokens, token)
}

func (s *Scanner) addIntToken() error {
	lexeme := string(s.source[s.start:s.current])
	intVal, err := strconv.ParseInt(lexeme, 10, 64)
	if err != nil {
		scanErr := NewScanErrorWithLog("Invalid integer literal: "+err.Error(), s.line, "")
		return scanErr
	}

	s.addToken(token.NUMBER_INT, intVal)

	return nil
}

func (s *Scanner) addRealToken() error {
	lexeme := string(s.source[s.start:s.current])
	realVal, err := strconv.ParseFloat(lexeme, 64)
	if err != nil {
		scanErr := NewScanErrorWithLog("Invalid float literal: "+err.Error(), s.line, "")
		return scanErr
	}

	s.addToken(token.NUMBER_REAL, realVal)

	return nil
}

func (s *Scanner) isAtEnd() bool {
	return s.current >= len(s.source)
}

func (s *Scanner) advance() (char rune) {
	char = s.source[s.current]
	s.current++

	return
}

func (s *Scanner) peek() rune {
	if s.isAtEnd() {
		return '\000'
	}

	return s.source[s.current]
}

func (s *Scanner) peekNext() rune {
	if s.current+1 >= len(s.source) {
		return '\000'
	}

	return s.source[s.current+1]
}

func (s *Scanner) advanceIfMatch(chars ...rune) bool {
	org := s.current

	for _, char := range chars {
		if s.isAtEnd() || s.source[s.current] != char {
			s.current = org

			return false
		}

		s.current++
	}

	return true
}

func (s *Scanner) isDigit(c rune) bool {
	return c >= '0' && c <= '9'
}

func (s *Scanner) isLetter(c rune) bool {
	return c == '_' || unicode.IsLetter(c)
}

func (s *Scanner) isLetterDigitMark(c rune) bool {
	return s.isLetter(c) || unicode.IsDigit(c) || unicode.IsMark(c)
}
