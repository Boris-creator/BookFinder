package fileutils

import (
	"io"
	"io/fs"
	"net/http"
	"os"
	"path/filepath"
)

func DownloadFile(url string, filePath string) error {
	response, err := http.Get(url)
	if err != nil {
		return err
	}
	defer response.Body.Close()

	out, err := CreateFile(filePath)
	if err != nil {
		return err
	}
	defer func() {
		err = out.Close()
	}()

	_, err = io.Copy(out, response.Body)
	return err
}

func FetchFile(url string) ([]byte, error) {
	response, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer response.Body.Close()
	return io.ReadAll(response.Body)
}

func CreateFile(filePath string) (*os.File, error) {
	if err := os.MkdirAll(filepath.Dir(filePath), fs.ModePerm); err != nil {
		return nil, err
	}
	return os.Create(filePath)
}
