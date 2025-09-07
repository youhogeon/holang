package scanner

import (
	"internal/util/log"
	"strconv"
	"unicode"
)

type Scanner struct {
	source []rune

	start   int
	current int
	line    int

	tokens []Token
}

func NewScanner(source string) *Scanner {
	return &Scanner{
		source: []rune(source),
	}
}

func (s *Scanner) ScanTokens() ([]Token, []error) {
	errors := make([]error, 0)

	for !s.isAtEnd() {
		s.start = s.current

		if err := s.scanToken(); err != nil {
			errors = append(errors, err)
		}
	}

	s.start = s.current
	s.addToken(EOF, nil)

	return s.tokens, errors
}

func (s *Scanner) scanToken() error {
	c := s.advance()

	switch c {
	case ' ', '\r', '\t':
		// Ignore whitespace.
	case '\n':
		s.line++
	case '(':
		s.addToken(LEFT_PAREN, nil)
	case ')':
		s.addToken(RIGHT_PAREN, nil)
	case '{':
		s.addToken(LEFT_BRACE, nil)
	case '}':
		s.addToken(RIGHT_BRACE, nil)
	case ',':
		s.addToken(COMMA, nil)
	case '-':
		s.addToken(MINUS, nil)
	case '+':
		s.addToken(PLUS, nil)
	case ':':
		s.addToken(COLON, nil)
	case ';':
		s.addToken(SEMICOLON, nil)
	case '?':
		s.addToken(QUESTION, nil)
	case '/':
		if s.advanceIfMatch('/') {
			for s.peek() != '\n' && !s.isAtEnd() {
				s.advance()
			}

			s.addToken(COMMENT, string(s.source[s.start+2:s.current]))
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
					s.addToken(MULTI_COMMENT, string(s.source[s.start+2:s.current-2]))

					break
				}

				s.advance()
			}
		} else {
			s.addToken(SLASH, nil)
		}
	case '*':
		s.addToken(STAR, nil)
	case '!':
		if s.advanceIfMatch('=') {
			s.addToken(BANG_EQUAL, nil)
		} else {
			s.addToken(BANG, nil)
		}
	case '=':
		if s.advanceIfMatch('=') {
			s.addToken(EQUAL_EQUAL, nil)
		} else {
			s.addToken(EQUAL, nil)
		}
	case '<':
		if s.advanceIfMatch('=') {
			s.addToken(LESS_EQUAL, nil)
		} else {
			s.addToken(LESS, nil)
		}
	case '>':
		if s.advanceIfMatch('=') {
			s.addToken(GREATER_EQUAL, nil)
		} else {
			s.addToken(GREATER, nil)
		}
	case '"':
		for s.peek() != '"' && !s.isAtEnd() {
			if s.peek() == '\n' {
				s.line++
			}

			s.advance()
		}

		if s.isAtEnd() {
			err := NewScanErrorWithLog("Unterminated string", s.line, "")
			return err
		}

		s.advance()

		s.addToken(STRING, string(s.source[s.start+1:s.current-1]))
	case '.':
		if !s.isDigit(s.peek()) {
			s.addToken(DOT, nil)
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
			tokenType, ok := keywords[text]
			if !ok {
				tokenType = IDENTIFIER
			}

			s.addToken(tokenType, nil)

			return nil
		}

		err := NewScanErrorWithLog("Unexpected character: "+string(c), s.line, "")
		return err
	}

	return nil
}

func (s *Scanner) addToken(t TokenType, literal any) {
	text := s.source[s.start:s.current]

	token := Token{
		TokenType: t,
		Lexeme:    string(text),
		Literal:   literal,
		Offset: Offset{
			Line:  s.line,
			Index: s.start,
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

	s.addToken(NUMBER_INT, intVal)

	return nil
}

func (s *Scanner) addRealToken() error {
	lexeme := string(s.source[s.start:s.current])
	realVal, err := strconv.ParseFloat(lexeme, 64)
	if err != nil {
		scanErr := NewScanErrorWithLog("Invalid float literal: "+err.Error(), s.line, "")
		return scanErr
	}

	s.addToken(NUMBER_REAL, realVal)

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
