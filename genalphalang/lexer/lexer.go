package lexer

import (
	"strings"
)

type TokenType int

const (
	TokenTypeIdentifier TokenType = iota
	TokenTypeNumber
	TokenTypeString
	TokenTypeBoolean
	TokenTypeKeyword
	TokenTypeOperator
	TokenTypePunctuation
	TokenTypeComment
	TokenTypeWhitespace
	TokenTypeNewline
	TokenTypeUnknown
)

type Keyword string

const (
	KeywordVar    Keyword = "fax"
	KeywordIf     Keyword = "skibidi"
	KeywordIfYes  Keyword = "yeah"
	KeywordIfNo   Keyword = "nah"
	KeywordFunc   Keyword = "lowkey"
	KeywordEnd    Keyword = "end"
	KeywordCall   Keyword = "fire"
	KeywordWhile  Keyword = "yap"
	KeywordImport Keyword = "gyat"
	KeywordReturn Keyword = "rizzult"
	KeywordTrue   Keyword = "yay"
	KeywordFalse  Keyword = "nay"
)

var (
	Keywords = []string{
		string(KeywordVar),
		string(KeywordIf),
		string(KeywordIfYes),
		string(KeywordIfNo),
		string(KeywordFunc),
		string(KeywordEnd),
		string(KeywordCall),
		string(KeywordWhile),
		string(KeywordImport),
		string(KeywordReturn),
	}
)

type Token struct {
	Type  TokenType
	Value string
	Line  int
}

func Lex(contents string, filename string) []Token {
	var tokens []Token

	for i, line := range strings.Split(contents, "\n") {
		line = strings.ReplaceAll(line, "\r", "") // windows line endings

		// first line must contain the special license sentence
		if i == 0 {
			if line != "on gyat no rizz this project shall be blessed by the cringe of all us rizzlers and shall be licensed under the skidibi license." {
				panic(filename + " First line must contain the special license sentence")
			}

			continue
		}

		tokens = append(tokens, lexLine(line, i)...)
		tokens = append(tokens, Token{
			Type:  TokenTypeNewline,
			Value: "\n",
			Line:  i,
		})
	}

	return tokens
}

// todo escaped characters in strings
func lexLine(line string, lineNum int) []Token {
	var tokens []Token

	var tokenStart int
	var inString bool
	var inComment bool

	var i = 0
	var char byte

	for {

		if i >= len(line) {
			if inString {
				panic("Unclosed string at line " + string(lineNum))
			}

			break
		}

		char = line[i]
		// fmt.Println(string(char), i, line)

		if inComment {
			i++
			continue
		}

		if inString {
			if char == '"' {
				tokens = append(tokens, Token{
					Type:  TokenTypeString,
					Value: line[tokenStart : i+1],
					Line:  lineNum,
				})
				inString = false
			}

			i++
			continue
		}

		if char == '"' {
			inString = true
			tokenStart = i

			i++
			continue
		}

		if char == '/' {
			tokens = append(tokens, Token{
				Type:  TokenTypeComment,
				Value: line[i:],
				Line:  lineNum,
			})
			break
		}

		if char == ' ' || char == '\t' {
			if inString {
				i++
				continue
			}

			tokens = append(tokens, Token{
				Type:  TokenTypeWhitespace,
				Value: string(line[i]),
				Line:  lineNum,
			})

			i++
			continue
		}

		if char == '[' || char == ']' || char == '(' || char == ')' || char == '{' || char == '}' || char == ',' || char == ';' {
			tokens = append(tokens, Token{
				Type:  TokenTypePunctuation,
				Value: string(char),
				Line:  lineNum,
			})

			i++
			continue
		}

		if char == '+' || char == '-' || char == '*' || char == '/' || char == '%' || char == '=' || char == '!' || char == '<' || char == '>' {
			tokens = append(tokens, Token{
				Type:  TokenTypeOperator,
				Value: string(char),
				Line:  lineNum,
			})

			i++
			continue
		}

		keyword := fromIndexContainsAny(line, i, Keywords)
		if keyword != "" {
			tokens = append(tokens, Token{
				Type:  TokenTypeKeyword,
				Value: keyword,
				Line:  lineNum,
			})

			i += len(keyword)
			continue
		}

		// identifier
		if char >= 'a' && char <= 'z' || char >= 'A' && char <= 'Z' || char == '_' {
			if inString {
				i++
				continue
			}

			var added = false
			for j := i + 1; j < len(line); j++ {
				if line[j] >= 'a' && line[j] <= 'z' || line[j] >= 'A' && line[j] <= 'Z' || line[j] >= '0' && line[j] <= '9' || line[j] == '_' {
					continue
				}

				tokens = append(tokens, Token{
					Type:  TokenTypeIdentifier,
					Value: line[i:j],
					Line:  lineNum,
				})

				added = true
				i = j
				break
			}

			if !added {
				tokens = append(tokens, Token{
					Type:  TokenTypeIdentifier,
					Value: line[i:],
					Line:  lineNum,
				})
				break
			}
		}

		if char >= '0' && char <= '9' {
			if inString {
				i++
				continue
			}

			var added = false
			for j := i + 1; j < len(line); j++ {
				if line[j] >= '0' && line[j] <= '9' {
					continue
				}

				tokens = append(tokens, Token{
					Type:  TokenTypeNumber,
					Value: line[i:j],
					Line:  lineNum,
				})

				added = true
				i = j
				break
			}

			if !added {
				tokens = append(tokens, Token{
					Type:  TokenTypeNumber,
					Value: line[i:],
					Line:  lineNum,
				})
			}

			i++
			continue
		}
	}

	return tokens
}

func fromIndexContainsAny(line string, i int, words []string) string {
	for _, word := range words {
		if fromIndexContainsWord(line, i, word) {
			return word
		}
	}

	return ""
}

func fromIndexContainsWord(line string, i int, word string) bool {
	if i+len(word) > len(line) {
		return false
	}

	return line[i:i+len(word)] == word
}
