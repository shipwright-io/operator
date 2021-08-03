package controllers

import (
	"fmt"
	"os"
)

// koDataPathEnv ko data-path environment variable.
const koDataPathEnv = "KO_DATA_PATH"

// koDataPath retrieve the data path environment variable, returning error when not found.
func koDataPath() (string, error) {
	dataPath, exists := os.LookupEnv(koDataPathEnv)
	if !exists {
		return "", fmt.Errorf("'%s' is not set", koDataPathEnv)
	}
	return dataPath, nil
}

// contains returns true if the string if found in the slice.
func contains(slice []string, str string) bool {
	for _, s := range slice {
		if s == str {
			return true
		}
	}
	return false
}
