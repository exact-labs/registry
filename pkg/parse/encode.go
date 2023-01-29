package parse

import (
	"errors"
	"regexp"
	"github.com/jxskiss/base62"
)

func EncodeName(package_name string) (string, error) {
	matched, _ := regexp.MatchString(`^[a-zA-Z0-9@][a-zA-Z0-9:_.-]*$`, package_name)
	if matched {
		return base62.EncodeToString([]byte(package_name)), nil
	} else {
		return "", errors.New("package name contains invalid characters")
	}
}

func OriginalName(encoded_name string) string {
	if encoded_name == "just_auth_system" {
		return "just_auth_system"
	} else {
		decoded_name, _ := base62.DecodeString(encoded_name)
		return string(decoded_name)
	}
}
