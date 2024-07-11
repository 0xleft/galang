package utils

import (
	"io"
	"os"
)

func ReadContents(filename string) string {
	file, err := os.Open(filename)
	if err != nil {
		panic(err)
		panic("Error opening file " + filename)
	}
	defer file.Close()

	contents, err := io.ReadAll(file)
	if err != nil {
		panic(err)
		panic("Error reading file " + filename)
	}

	return string(contents)
}
