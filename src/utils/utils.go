package utils

import (
	"encoding/json"
	"fmt"
	"os"
)

func IsLetter(char rune) bool {
	return (char >= 'a' && char <= 'z') || (char >= 'A' && char <= 'Z') || char == '_'
}

func IsDigit(char rune) bool {
	return (char >= '0' && char <= '9')
}

var EscapeSeq map[rune]rune = map[rune]rune{
	'b':  '\b',
	'f':  '\f',
	'n':  '\n',
	'r':  '\r',
	't':  '\t',
	'v':  '\v',
	'\\': '\\',
	'\'': '\'',
	'"':  '"',
}

func GetFilesOfDir(dir string) ([][]byte, error) {
	entries, err := os.ReadDir(dir)
	for _, entry := range entries {
		if !entry.IsDir() {
		}
	}
	return nil, err
}

func PrintData(d any) {
	data, _ := json.MarshalIndent(d, "", "  ")
	fmt.Println(string(data))
}
