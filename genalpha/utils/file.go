package utils

import (
	"io"
	"io/ioutil"
	"os"
	"os/exec"
	"path"
	"path/filepath"
)

func CopyFile(src, dst string) error {
	sourceFile, err := os.Open(src)
	if err != nil {
		return err
	}
	defer sourceFile.Close()

	destinationFile, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer destinationFile.Close()

	_, err = io.Copy(destinationFile, sourceFile)
	return err
}

// CopyDir copies all files and directories from src to dst
func CopyDir(src string, dst string) error {
	srcInfo, err := os.Stat(src)
	if err != nil {
		return err
	}

	// Create the destination directory
	if err = os.MkdirAll(dst, srcInfo.Mode()); err != nil {
		return err
	}

	// Read the contents of the source directory
	entries, err := ioutil.ReadDir(src)
	if err != nil {
		return err
	}

	for _, entry := range entries {
		srcPath := path.Join(src, entry.Name())
		dstPath := path.Join(dst, entry.Name())

		if entry.IsDir() {
			// Recursively copy the directory
			if err = CopyDir(srcPath, dstPath); err != nil {
				return err
			}
		} else {
			// Copy the file
			if err = CopyFile(srcPath, dstPath); err != nil {
				return err
			}
		}
	}
	return nil
}

func WriteContents(filename string, contents string) {
	os.MkdirAll(filepath.Dir(filename), os.ModeDir)
	file, err := os.Create(filename)

	if err != nil {
		panic("Error creating file " + filename)
	}
	defer file.Close()

	_, err = file.WriteString(contents)
	if err != nil {
		// todo
		panic("Error writing to file " + filename)
	}
}

func ReadContents(filename string) string {
	file, err := os.Open(filename)
	if err != nil {
		// todo
		panic("Error opening file " + filename)
	}
	defer file.Close()

	contents, err := io.ReadAll(file)
	if err != nil {
		// todo
		panic("Error reading file " + filename)
	}

	return string(contents)
}

func FileExists(filename string) bool {
	_, err := os.Stat(filename)
	return err == nil
}

func IsExecutableInPath(name string) bool {
	_, err := exec.LookPath(name)
	return err == nil
}
