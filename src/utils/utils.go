package utils

import (
	"encoding/json"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
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

func GetFilesOfDir(dirpath, root, path string) ([]string, error) {
	var dir string
	if _, err := os.Stat(filepath.Join(root, "libs", dirpath)); err == nil {
		dir = filepath.Join(root, "libs", dirpath)
	} else if _, err := os.Stat(filepath.Join(path, dirpath)); err == nil {
		dir = filepath.Join(path, dirpath)
	} else {
		return nil, fmt.Errorf("%s was not found in %s or %s", dirpath, root, path)
	}

	files, err := fs.Glob(os.DirFS(dir), "*.li")
	for i, file := range files {
		files[i] = filepath.Join(dir, file)
	}
	if err != nil {
		return nil, err
	}
	return files, nil
}

func PrintData(d any) {
	data, _ := json.MarshalIndent(d, "", "  ")
	fmt.Println(string(data))
}
