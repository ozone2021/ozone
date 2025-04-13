package utils

import (
	"log"
	"strings"
)

func ValidateName(name string) bool {
	// Check if the name contains a forward slash
	if strings.Contains(name, "/") {
		return true
	} else if name == "master" || name == "dev" || name == "main" || name == "develop" {
		return true
	}
	log.Println("Invalid name:", name)
	return false
}
