package main

import (
	"fmt"
	"os"
)

func noFile() {
}

func withFile(filename string) {

}

func main() {
	defer func() {
		if r := recover(); r != nil {
			fmt.Println("Execution failed: ", r)
		}
	}()

	if len(os.Args) < 2 {
		noFile()
		return
	}

	filename := os.Args[1]
	withFile(filename)
}
