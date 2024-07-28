package pkg

import (
	"fmt"
	"os"
	"runtime"
	"strings"

	"bobik.squidwock.com/root/gal/genalpha/utils"
	"github.com/go-git/go-git/v5"
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

func GetGalDirectory() string {
	system := runtime.GOOS
	userdir := os.Getenv("HOME")

	if userdir == "" {
		switch system {
		case "windows":
			userdir = os.Getenv("APPDATA")
		case "darwin":
			userdir = os.Getenv("HOME")
		default:
			userdir = os.Getenv("HOME")
		}
	}

	return userdir + "/.gal/"
}

func GetInstalledPackagesDirectory() string {
	return GetGalDirectory() + "packages/"
}

func GetPackagePath(name string) string {
	return GetInstalledPackagesDirectory() + name + "/"
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
	path := GetPackagePath(pkg.Name)
	for _, dep := range pkg.Dependencies {
		err := InstallPackage(dep.Name)
		if err != nil {
			return "", err
		}
	}

	// add all executables to path
	execPath := GetGalDirectory() + "executables/"

	for _, exe := range pkg.Executables {
		if runtime.GOOS == "windows" {
			// add a .bat file to the Package directory
			batPath := execPath + exe.Name + ".bat"
			batContents := `
@echo off
set EXECUTABLE_PATH="` + path + exe.Path + `"
gal.exe run %EXECUTABLE_PATH% %*
`
			utils.WriteContents(batPath, batContents)
		} else {
			// create a executable file
			filePath := execPath + exe.Name
			fileContents := `
#!/bin/bash

export EXECUTABLE_PATH=` + path + exe.Path + `
gal run $EXECUTABLE_PATH $@
`
			utils.WriteContents(filePath, fileContents)
			os.Chmod(filePath, 0755)
		}
	}

	return path, nil
}

func InstallPackage(name string) error {
	if strings.HasPrefix(name, ".") {
		pack, err := ParsePackage(name + "/package.yml")
		if err != nil {
			return err
		}

		err = utils.CopyDir(name, GetInstalledPackagesDirectory()+pack.Name)
		if err != nil {
			return err
		}

		_, err = BuildPackage(pack)
		fmt.Println("Installed package", pack.Name)
		return err
	}

	if strings.HasPrefix(name, "git+") {
		pkgName := strings.Split(name, "/")[len(strings.Split(name, "/"))-1]
		UninstallPackage(pkgName)
		path := GetPackagePath(pkgName)
		// clone the repository
		_, err := git.PlainClone(path, false, &git.CloneOptions{
			URL:          name[4:],
			Depth:        0,
			SingleBranch: true,
		})
		if err != nil {
			return err
		}
		// build the package
		pack, err := ParsePackage(path + "package.yml")
		if err != nil {
			return err
		}
		// install the package
		_, err = BuildPackage(pack)
		fmt.Println("Installed package", pkgName)
		return err
	}

	// get the package from the package repository packages.gal.squidwock.com
	// build the package
	// install the package
	panic("not implemented")

	return nil
}

func UninstallPackage(name string) error {
	path := GetPackagePath(name)
	if !utils.FileExists(path) {
		return fmt.Errorf("Package %s is not installed", name)
	}

	// remove executables
	pack, err := ParsePackage(path + "package.yml")
	if err != nil {
		return err
	}

	for _, exe := range pack.Executables {
		execPath := GetGalDirectory() + "executables/"
		if os.Getenv("GOOS") == "windows" {
			batPath := execPath + exe.Name + ".bat"
			os.Remove(batPath)
		} else {
			filePath := execPath + exe.Name
			os.Remove(filePath)
		}
	}

	err = os.RemoveAll(path)
	if err != nil {
		return err
	}

	fmt.Println("Uninstalled package", name)
	return err
}
