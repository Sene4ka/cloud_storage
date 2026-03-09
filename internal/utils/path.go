package utils

import (
	"fmt"
	"regexp"
	"strings"
)

var validPathRegex = regexp.MustCompile(`^[a-zA-Z0-9_ -]+(/[a-zA-Z0-9_ -]+)*$`)

func ValidatePath(path string) error {
	if path == "" {
		return fmt.Errorf("path cannot be empty")
	}
	if path[0] != '/' {
		return fmt.Errorf("path must start with '/'")
	}
	if path != "/" && strings.HasSuffix(path, "/") {
		return fmt.Errorf("path cannot end with '/' (except root)")
	}
	if strings.Contains(path, "//") {
		return fmt.Errorf("path cannot contain consecutive slashes")
	}
	if strings.Contains(path, "..") {
		return fmt.Errorf("path cannot contain '..'")
	}

	trimmed := strings.TrimPrefix(path, "/")
	if trimmed != "" && !validPathRegex.MatchString(trimmed) {
		return fmt.Errorf("path contains invalid characters (allowed: letters, digits, space, underscore, hyphen, slash)")
	}
	return nil
}
