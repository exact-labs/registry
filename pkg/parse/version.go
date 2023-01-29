package parse

import (
	"regexp"
)

func HasVersion(packageName string) bool {
   notVersion := regexp.MustCompile(`^[a-zA-Z0-9@][a-zA-Z0-9.-]*$`).MatchString
   isVersion := regexp.MustCompile(`^[a-zA-Z0-9@][a-zA-Z0-9.-@]*$`).MatchString

   if notVersion(packageName) {
      return false
   } else {
      if isVersion(packageName) {
         return true
      } else {
         return false
      }
   }
   
   return false
}