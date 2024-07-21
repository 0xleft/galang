package utils

import (
	"os"
	"runtime"
)

func GetInstalledPackagesDirectory() string {
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

	return userdir + "/.gal/packages/"
}
