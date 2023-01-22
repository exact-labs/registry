package parse

import (
	"errors"
	"regexp"
	// "strings"
	"github.com/jxskiss/base62"
)

func EncodeName(package_name string) (string, error) {
	matched, _ := regexp.MatchString(`^[a-zA-Z0-9@][a-zA-Z0-9\/_.-]*$`, package_name)
	// format := strings.NewReplacer("-", "_JFMTdash_", ".", "_JFMTdot_", "@", "_JFMTat_", "/", "_JFMTslash_")
	if matched {
		// return format.Replace(package_name), nil
		return base62.EncodeToString([]byte(package_name)), nil
	} else {
		return "", errors.New("package name contains invalid characters")
	}
}

func OriginalName(encoded_name string) string {
	// format := strings.NewReplacer("_JFMTdash_", "-", "_JFMTdot_", ".", "_JFMTat_", "@", "_JFMTslash_", "/")
	// return format.Replace(encoded_name)
	if encoded_name == "just_auth_system" {
		return "just_auth_system"
	} else {
		decoded_name, _ := base62.DecodeString(encoded_name)
		return string(decoded_name)
	}
}
