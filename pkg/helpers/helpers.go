package helpers

import (
	"archive/tar"
	"compress/gzip"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing/fstest"
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

func ReadTar(source string) (fstest.MapFS, error) {
	file, err := os.Open(source)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	gzRead, err := gzip.NewReader(file)
	if err != nil {
		return nil, err
	}

	tarRead := tar.NewReader(gzRead)
	files := make(fstest.MapFS)

	for {
		cur, err := tarRead.Next()
		if err == io.EOF {
			break
		} else if err != nil {
			return nil, err
		}
		if cur.Typeflag != tar.TypeReg {
			continue
		}
		data, err := io.ReadAll(tarRead)
		if err != nil {
			return nil, err
		}
		files[cur.Name] = &fstest.MapFile{Data: data}
	}

	return files, nil
}
