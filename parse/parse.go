package parse

import (
	"errors"
	"regexp"
	"strings"
)

func EncodeName(package_name string) (string, error) {
	matched, _ := regexp.MatchString(`^[a-zA-Z0-9@][a-zA-Z0-9\/_.-]*$`, package_name)
	format := strings.NewReplacer("-", "_JFMTdash_", ".", "_JFMTdot_", "@", "_JFMTat_", "/", "_JFMTslash_")

	if matched {
		return format.Replace(package_name), nil
	} else {
		return "", errors.New("package name contains invalid characters")
	}
}

func OriginalName(encoded_name string) string {
	format := strings.NewReplacer("_JFMTdash_", "-", "_JFMTdot_", ".", "_JFMTat_", "@", "_JFMTslash_", "/")
	return format.Replace(encoded_name)
}
