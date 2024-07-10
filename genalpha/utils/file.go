package utils

import (
	"io"
	"os"
)

func ReadContents(filename string) string {
	file, err := os.Open(filename)
	if err != nil {
		panic("Error opening file")
	}
	defer file.Close()

	contents, err := io.ReadAll(file)
	if err != nil {
		panic("Error reading file")
	}

	return string(contents)
}
