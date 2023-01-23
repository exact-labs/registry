package templates

import (
	"archive/zip"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/go-git/go-git/v5"
)

func recursiveZip(pathToZip, destinationPath string) error {
	destinationFile, err := os.Create(destinationPath)
	if err != nil {
		return err
	}
	myZip := zip.NewWriter(destinationFile)
	err = filepath.Walk(pathToZip, func(filePath string, info os.FileInfo, err error) error {
		if info.IsDir() {
			return nil
		}
		if err != nil {
			return err
		}
		relPath := strings.TrimPrefix(filePath, filepath.Dir(pathToZip))
		zipFile, err := myZip.Create(relPath)
		if err != nil {
			return err
		}
		fsFile, err := os.Open(filePath)
		if err != nil {
			return err
		}
		_, err = io.Copy(zipFile, fsFile)
		if err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		return err
	}
	err = myZip.Close()
	if err != nil {
		return err
	}
	return nil
}

func Dir() string {
	if strings.HasPrefix(os.Args[0], os.TempDir()) {
		return "./templates"
	}
	return filepath.Join(os.Args[0], "../templates")
}

func ensureDir(dirName string) error {
	err := os.Mkdir(dirName, 0700)
	if err == nil {
		return nil
	}
	if os.IsExist(err) {
		info, err := os.Stat(dirName)
		if err != nil {
			return err
		}
		if !info.IsDir() {
			return errors.New("path exists but is not a directory")
		}
		return nil
	}
	return err
}

func listDir(dirName string) ([]string, error) {
	file, err := os.Open(dirName)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	list, err := file.Readdirnames(0)
	if err != nil {
		return nil, err
	}

	return list, nil
}

func Copy() error {
	if err := os.RemoveAll(Dir()); err != nil {
		return err
	}

	if err := ensureDir(Dir()); err != nil {
		return err
	}

	if _, err := git.PlainClone(Dir(), false, &git.CloneOptions{URL: "https://github.com/exact-rs/templates", Progress: os.Stdout}); err != nil {
		return err
	}

	if err := os.RemoveAll(fmt.Sprintf("%s/.git", Dir())); err != nil {
		return err
	}

	files, err := listDir(Dir())
	if err != nil {
		return err
	}

	for _, folder := range files {
		recursiveZip(fmt.Sprintf("%s/%s", Dir(), folder), fmt.Sprintf("%s/%s.zip", Dir(), folder))
      if err := os.RemoveAll(fmt.Sprintf("%s/%s", Dir(), folder)); err != nil {
         return err
      }
	}

	return nil
}
