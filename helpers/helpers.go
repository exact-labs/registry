package helpers

import (
   "os"
   "path/filepath"
   "strings"
)

func InspectRuntime() (baseDir string, withGoRun bool) {
   if strings.HasPrefix(os.Args[0], os.TempDir()) {
      withGoRun = true
      baseDir, _ = os.Getwd()
   } else {
      withGoRun = false
      baseDir = filepath.Dir(os.Args[0])
   }
   return
}