package lexer

import (
	"strings"

	genalphatypes "bobik.squidwock.com/root/gal/genalpha"
)

func Lex(contents string) []genalphatypes.Token {
	var tokens []genalphatypes.Token

	for i, line := range strings.Split(contents, "\n") {
		line = strings.ReplaceAll(line, "\r", "") // windows line endings

		tokens = append(tokens, lexLine(line, i)...)
		tokens = append(tokens, genalphatypes.Token{
			Type:  genalphatypes.TokenTypeNewline,
			Value: "\n",
			Line:  i,
		})
	}

	return tokens
}

// todo escaped characters in strings
func lexLine(line string, lineNum int) []genalphatypes.Token {
	var tokens []genalphatypes.Token

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

		if inComment {
			i++
			continue
		}

		if inString {
			if line[i-1] == '\\' && char == '"' {
				i++
				continue
			}
			if char == '"' {
				replacables := []string{"\\\""}
				replaces := []string{"\""}
				finalString := line[tokenStart+1 : i]
				for j, replacable := range replacables {
					finalString = strings.ReplaceAll(finalString, replacable, replaces[j])
				}
				tokens = append(tokens, genalphatypes.Token{
					Type:  genalphatypes.TokenTypeString,
					Value: finalString,
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

		if char == '`' {
			tokens = append(tokens, genalphatypes.Token{
				Type:  genalphatypes.TokenTypeComment,
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

			tokens = append(tokens, genalphatypes.Token{
				Type:  genalphatypes.TokenTypeWhitespace,
				Value: string(line[i]),
				Line:  lineNum,
			})

			i++
			continue
		}

		if char == '[' || char == ']' || char == '(' || char == ')' || char == '{' || char == '}' || char == ',' {
			tokens = append(tokens, genalphatypes.Token{
				Type:  genalphatypes.TokenTypePunctuation,
				Value: string(char),
				Line:  lineNum,
			})

			i++
			continue
		}

		if char == ';' {
			tokens = append(tokens, genalphatypes.Token{
				Type:  genalphatypes.TokenTypeNewline,
				Value: string(char),
				Line:  lineNum,
			})

			i++
			continue
		}

		// number
		if char >= '0' && char <= '9' {
			if inString {
				i++
				continue
			}

			var added = false
			for j := i + 1; j < len(line); j++ {
				if line[j] >= '0' && line[j] <= '9' || line[j] == '.' {
					continue
				}

				tokens = append(tokens, genalphatypes.Token{
					Type:  genalphatypes.TokenTypeNumber,
					Value: line[i:j],
					Line:  lineNum,
				})

				added = true
				i = j
				break
			}

			if !added {
				tokens = append(tokens, genalphatypes.Token{
					Type:  genalphatypes.TokenTypeNumber,
					Value: line[i:],
					Line:  lineNum,
				})

				i += len(line[i:])
			}

			continue
		}

		if char == '+' || char == '-' || char == '*' || char == '/' || char == '%' || char == '=' || char == '!' || char == '<' || char == '>' || char == '&' || char == '|' || char == '^' {
			tokens = append(tokens, genalphatypes.Token{
				Type:  genalphatypes.TokenTypeOperator,
				Value: string(char),
				Line:  lineNum,
			})

			i++
			continue
		}

		keyword := fromIndexContainsAny(line, i, genalphatypes.Keywords)
		if keyword != "" {
			tokens = append(tokens, genalphatypes.Token{
				Type:  genalphatypes.TokenTypeKeyword,
				Value: keyword,
				Line:  lineNum,
			})

			i += len(keyword)
			continue
		}

		// identifier
		if char >= 'a' && char <= 'z' || char >= 'A' && char <= 'Z' || char == '_' || char == '.' {
			if inString {
				i++
				continue
			}

			var added = false
			for j := i + 1; j < len(line); j++ {
				if line[j] >= 'a' && line[j] <= 'z' || line[j] >= 'A' && line[j] <= 'Z' || line[j] >= '0' && line[j] <= '9' || line[j] == '_' || line[j] == '.' {
					continue
				}

				tokens = append(tokens, genalphatypes.Token{
					Type:  genalphatypes.TokenTypeIdentifier,
					Value: line[i:j],
					Line:  lineNum,
				})

				added = true
				i = j
				break
			}

			if !added {
				tokens = append(tokens, genalphatypes.Token{
					Type:  genalphatypes.TokenTypeIdentifier,
					Value: line[i:],
					Line:  lineNum,
				})

				i += len(line[i:])
			}

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
