package pkg

import (
	"fmt"
	"strings"

	"bobik.squidwock.com/root/gal/genalpha/utils"
	"gopkg.in/yaml.v3"
)

type Author struct {
	Name  string
	Email string
}

type Executable struct {
	Name string
	Path string
}

type Dependency struct {
	Name    string
	Version string
}

type Package struct {
	Name         string       `yaml:"name"`
	Version      string       `yaml:"version"`
	Authors      []Author     `yaml:"authors"`
	License      string       `yaml:"license"`
	Homepage     string       `yaml:"homepage"`
	Description  string       `yaml:"description"`
	Keywords     []string     `yaml:"keywords"`
	Dependencies []Dependency `yaml:"dependencies"`
	Executables  []Executable `yaml:"executables"`
}

func ParsePackage(filename string) (Package, error) {
	contents := utils.ReadContents(filename)
	var pkg Package
	err := yaml.Unmarshal([]byte(contents), &pkg)
	if err != nil {
		return Package{}, err
	}
	return pkg, nil
}

// returns the path to the built package and error if any
// building the package means install all the executables defined in the package.yml and dependencies
func BuildPackage(pkg Package) (string, error) {
	// todo
	return "", nil
}

func InstallPackage(name string) error {
	// todo check if the package is already installed

	path := utils.GetPackagePath(name)
	if utils.FileExists(path) {
		fmt.Println("Package", name, "is already installed")
		return nil
	}

	if strings.HasPrefix(name, "git+") {
		// clone the repository
		// build the package
		// install the package
	}

	// get the package from the package repository packages.gal.squidwock.com
	// build the package
	// install the package

	return nil
}

func UninstallPackage(name string) error {

	path := utils.GetPackagePath(name)
	if !utils.FileExists(path) {
		fmt.Println("Package", name, "is not installed")
	}

	// remove the package directory
	// todo
	return nil
}
