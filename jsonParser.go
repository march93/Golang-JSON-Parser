package main

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"strconv"
)

type Token int

const (
	WHITESPACE  Token = iota // 0
	LEFTCURLY                // 1
	RIGHTCURLY               // 2
	LEFTSQUARE               // 3
	RIGHTSQUARE              // 4
	COLON                    // 5
	COMMA                    // 6
	BOOL                     // 7
	STRING                   // 8
	INT                      // 9
)

type TokenList struct {
	token Token
	char  string
}

var indentLevel = 0
var squareOne = 0

func scanFile(input []byte) []TokenList {
	tokens := []TokenList{}
	var strBuffer bytes.Buffer
	var intBuffer bytes.Buffer
	var boolBuffer bytes.Buffer
	strBlocker := false
	intBlocker := false
	boolBlocker := false
	previousByte := ""

	for _, byte := range input {
		str := string(byte)

		// is int
		if _, err := strconv.Atoi(str); err == nil {
			if !strBlocker {
				if intBuffer.Len() > 0 {
					intBuffer.WriteString(str)
					previousByte = str
				} else {
					intBuffer.WriteString(str)
					previousByte = str
					intBlocker = true
				}
			}
		} else { // not int
			if intBuffer.Len() > 0 {
				// check if == . || e || E || + || -
				if str == "." || str == "e" || str == "E" || str == "+" || str == "-" {
					intBuffer.WriteString(str)
					previousByte = str
				} else {
					tokens = append(tokens, TokenList{INT, intBuffer.String()})
					intBuffer.Reset()
					intBlocker = false
				}
			} else {
				if !strBlocker && !intBlocker && !boolBlocker {
					if str == "." || str == "e" || str == "E" || str == "+" || str == "-" {
						intBuffer.WriteString(str)
						previousByte = str
					}
				}
			}
		}

		// check for bool
		if str == "t" || str == "f" || str == "n" {
			if !strBlocker && !intBlocker {
				boolBuffer.WriteString(str)
				previousByte = str
				boolBlocker = true
			}
		} else if boolBlocker {
			if boolBuffer.String() == "true" || boolBuffer.String() == "false" || boolBuffer.String() == "null" {
				tokens = append(tokens, TokenList{BOOL, boolBuffer.String()})
				boolBuffer.Reset()
				boolBlocker = false
			} else {
				boolBuffer.WriteString(str)
				previousByte = str
			}
		}

		if !intBlocker && !boolBlocker {
			switch str {
			case " ", "\t", "\n":
				if !strBlocker {
					tokens = append(tokens, TokenList{WHITESPACE, str})
				} else {
					strBuffer.WriteString(str)
					previousByte = str
				}
			case "{":
				if !strBlocker {
					tokens = append(tokens, TokenList{LEFTCURLY, str})
				} else {
					strBuffer.WriteString(str)
					previousByte = str
				}
			case "}":
				if !strBlocker {
					tokens = append(tokens, TokenList{RIGHTCURLY, str})
				} else {
					strBuffer.WriteString(str)
					previousByte = str
				}
			case "[":
				if !strBlocker {
					tokens = append(tokens, TokenList{LEFTSQUARE, str})
				} else {
					strBuffer.WriteString(str)
					previousByte = str
				}
			case "]":
				if !strBlocker {
					tokens = append(tokens, TokenList{RIGHTSQUARE, str})
				} else {
					strBuffer.WriteString(str)
					previousByte = str
				}
			case ":":
				if !strBlocker {
					tokens = append(tokens, TokenList{COLON, str})
				} else {
					strBuffer.WriteString(str)
					previousByte = str
				}
			case ",":
				if !strBlocker {
					tokens = append(tokens, TokenList{COMMA, str})
				} else {
					strBuffer.WriteString(str)
					previousByte = str
				}
			case "\"":
				if strBuffer.Len() > 0 && previousByte != "\\" { // opening double quote already in buffer
					strBuffer.WriteString(str)
					tokens = append(tokens, TokenList{STRING, strBuffer.String()})
					strBuffer.Reset()
					strBlocker = false
				} else {
					strBuffer.WriteString(str)
					previousByte = str
					strBlocker = true
				}
			default:
				if strBlocker && strBuffer.Len() > 0 {
					strBuffer.WriteString(str)
					previousByte = str
				}
			}
		}
	}
	return tokens
}

func formatBracket(indent int, toIndent bool, wasSquare bool, square int, str string) {
	var indentBuffer bytes.Buffer
	for i := 0; i < indent; i++ {
		indentBuffer.WriteString("\t")
	}
	if str == "{" {
		if toIndent && indent == 0 {
			fmt.Printf(indentBuffer.String() + "<span style='color:blue'>" + str + "</span>" + "\n")
			indentBuffer.Reset()
		} else if wasSquare {
			if square == 0 {
				fmt.Printf(indentBuffer.String() + "<span style='color:blue'>" + str + "</span>" + "\n")
				indentBuffer.Reset()
			} else {
				fmt.Printf("\n" + indentBuffer.String() + "<span style='color:blue'>" + str + "</span>" + "\n")
				indentBuffer.Reset()
			}
		} else {
			fmt.Printf("\n\t" + indentBuffer.String() + "<span style='color:blue'>" + str + "</span>" + "\n")
			indentLevel++
			indentBuffer.Reset()
		}
		indentBuffer.Reset()
	} else if str == "[" {
		fmt.Printf("<span style='color:limegreen'>" + str + "</span>" + "\n")
		indentBuffer.Reset()
	} else if str == "}" {
		fmt.Printf("\n" + indentBuffer.String() + "<span style='color:blue'>" + str + "</span>")
		indentLevel--
		indentBuffer.Reset()
	} else if str == "]" {
		fmt.Printf("\n" + indentBuffer.String() + "<span style='color:limegreen'>" + str + "</span>")
		indentBuffer.Reset()
	}
}

func formatString(indent int, toIndent bool, str string) {
	var indentBuffer bytes.Buffer
	var stringBuffer bytes.Buffer
	var escapeBuffer bytes.Buffer
	escapeBlocker := false
	unicodeBlocker := false
	unicodeCounter := 0

	for i := 0; i < indent; i++ {
		indentBuffer.WriteString("\t")
	}

	for _, char := range str {
		if escapeBlocker {
			if string(char) == "u" {
				escapeBuffer.WriteString(string(char))
				unicodeBlocker = true
				unicodeCounter = unicodeCounter + 2
			} else {
				if unicodeBlocker {
					if unicodeCounter == 5 {
						escapeBuffer.WriteString(string(char) + "</span>")
						stringBuffer.WriteString(escapeBuffer.String())
						escapeBuffer.Reset()
						escapeBlocker = false
						unicodeBlocker = false
						unicodeCounter = 0
					} else {
						escapeBuffer.WriteString(string(char))
						unicodeCounter++
					}
				} else {
					escapeBuffer.WriteString(string(char) + "</span>")
					stringBuffer.WriteString(escapeBuffer.String())
					escapeBuffer.Reset()
					escapeBlocker = false
				}
			}
		} else {
			switch string(char) {
			case "\\":
				escapeBuffer.WriteString("<span style='color:black'>" + string(char))
				escapeBlocker = true
			case "<":
				stringBuffer.WriteString("&lt;")
			case ">":
				stringBuffer.WriteString("&gt;")
			case "&":
				stringBuffer.WriteString("&amp;")
			case "\"":
				stringBuffer.WriteString("&quot;")
			case "'":
				stringBuffer.WriteString("&apos;")
			default:
				stringBuffer.WriteString(string(char))
			}
		}
	}

	if toIndent {
		fmt.Printf("%s", indentBuffer.String()+"<span style='color:red'>"+stringBuffer.String()+"</span>")
		stringBuffer.Reset()
	} else {
		fmt.Printf("%s", "<span style='color:red'>"+stringBuffer.String()+"</span>")
		stringBuffer.Reset()
	}
}

func formatInt(indent int, toIndent bool, str string) {
	var indentBuffer bytes.Buffer
	for i := 0; i < indent; i++ {
		indentBuffer.WriteString("\t")
	}

	if toIndent {
		fmt.Printf(indentBuffer.String() + "<span style='color:mediumpurple'>" + str + "</span>")
	} else {
		fmt.Printf("<span style='color:mediumpurple'>" + str + "</span>")
	}
}

func formatBool(indent int, toIndent bool, str string) {
	var indentBuffer bytes.Buffer
	for i := 0; i < indent; i++ {
		indentBuffer.WriteString("\t")
	}

	if toIndent {
		fmt.Printf(indentBuffer.String() + "<span style='color:cyan'>" + str + "</span>")
	} else {
		fmt.Printf("<span style='color:cyan'>" + str + "</span>")
	}
}

func formatFile(tokens []TokenList) {
	toIndent := false
	wasSquare := false

	if len(tokens) < 3 { // input is {}
		fmt.Printf("{\n}")
	} else {
		for _, token := range tokens {
			switch token.token {
			case 0:
				// ignore all whitespace
			case 1:
				// Left curly bracket
				toIndent = true
				formatBracket(indentLevel, toIndent, wasSquare, squareOne, token.char)
				squareOne++
				indentLevel++
			case 2:
				// Right curly bracket
				indentLevel--
				toIndent = false
				formatBracket(indentLevel, toIndent, wasSquare, squareOne, token.char)
				if wasSquare {
					indentLevel++
				}
			case 3:
				// Left square bracket
				toIndent = true
				wasSquare = true
				squareOne = 0
				formatBracket(indentLevel, toIndent, wasSquare, squareOne, token.char)
				indentLevel++
			case 4:
				// Right square bracket
				indentLevel--
				toIndent = false
				wasSquare = false
				squareOne = 0
				formatBracket(indentLevel, toIndent, wasSquare, squareOne, token.char)
			case 5:
				// Colon
				toIndent = false
				fmt.Printf("<span style='color:darkgray'>:</span> ")
			case 6:
				// Comma
				toIndent = true
				fmt.Printf("<span style='color:darkorange'>,</span> \n")
			case 7:
				// Bool
				formatBool(indentLevel, toIndent, token.char)
			case 8:
				// String
				formatString(indentLevel, toIndent, token.char)
			case 9:
				// Int
				formatInt(indentLevel, toIndent, token.char)
			}
		}
	}
}

func main() {
	if len(os.Args) < 2 {
		fmt.Printf("Filename not provided.")
	}
	content, err := ioutil.ReadFile(os.Args[1])
	if err != nil {
		log.Fatal(err)
	} else {
		fmt.Printf("<span style='font-family:monospace; white-space:pre'>\n")

		// Tokenize
		tokenArr := scanFile(content)

		// Format
		formatFile(tokenArr)

		fmt.Printf("</span>")
	}
}
