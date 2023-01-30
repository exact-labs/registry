package parse

import (
	"regexp"
   "strings"
   
	"registry/pkg/helpers"
)


func HasSemVersion(packageName string) bool {
	semVerCheck := regexp.MustCompile(`^(0|[1-9][0-9]*)\.(0|[1-9][0-9]*)\.(0|[1-9][0-9]*)(?:-((?:0|[1-9][0-9]*|[0-9]*[a-zA-Z-][0-9a-zA-Z-]*)(?:\.(?:0|[1-9][0-9]*|[0-9]*[a-zA-Z-][0-9a-zA-Z-]*))*))?(?:\+([0-9a-zA-Z-]+(?:\.[0-9a-zA-Z-]+)*))?$`).MatchString

	return semVerCheck(helpers.SplitLast(strings.Split(packageName, "@")))
}

func GetSemVer(semString string) string {
   semFind := regexp.MustCompile(`([0-9]+(\.[0-9]+)+).*[A-Za-z0-9]+`).FindAllString
   
   return semFind(semString, -1)[0]
}