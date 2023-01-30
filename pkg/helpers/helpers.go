package helpers

import (
	"compress/gzip"
	"os"
   "io/fs"
	"path/filepath"
	"strings"

	tarfs "github.com/nlepage/go-tarfs"
)

func SplitLast(string []string) string {
    return string[len(string)-1]
}

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

func TarPath() string {
	_, DebugMode := InspectRuntime()
	if DebugMode {
		return "http://localhost:8090"
	} else {
		return "https://r.justjs.dev"
	}
}

func ReadFromTar(name string, source string) ([]byte, error) {
	pkg, err := os.Open(source)
	if err != nil {
		return nil, err
	}
	defer pkg.Close()

	gz, err := gzip.NewReader(pkg)
	if err != nil {
		return nil, err
	}

	tar, err := tarfs.New(gz)
	if err != nil {
		return nil, err
	}

   bytes, err := fs.ReadFile(tar, name)
   if err != nil {
      return nil, err
   }

	return bytes, nil
}
