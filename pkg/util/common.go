package util

import (
	"fmt"
	"io/ioutil"
	"net/url"
	"os"
	"path"
	"path/filepath"
	"strings"
)

func ParseEndpoint(endpoint string) (string, string, error) {
	u, err := url.Parse(endpoint)
	if err != nil {
		return "", "", fmt.Errorf("could not parse endpoint: %v", err)
	}

	var address string
	if len(u.Host) == 0 {
		address = filepath.FromSlash(u.Path)
	} else {
		address = path.Join(u.Host, filepath.FromSlash(u.Path))
	}

	scheme := strings.ToLower(u.Scheme)
	if scheme == "unix" {
		if err := os.Remove(address); err != nil && !os.IsNotExist(err) {
			return "", "", fmt.Errorf("could not remove unix socket %q: %v", address, err)
		}
	} else {
		return "", "", fmt.Errorf("unsupported protocol: %s", scheme)
	}

	return scheme, address, nil
}

func CreateFile(path, contents string) (string, error) {
	tmpFile, err := ioutil.TempFile(path, "")
	if err != nil {
		return "", fmt.Errorf("error creating file: %s", err)
	}

	filePath := tmpFile.Name()
	fileContents := []byte(contents)

	if _, err := tmpFile.Write(fileContents); err != nil {
		return "", fmt.Errorf("error writing to file %s: %s", filePath, err)
	}

	if err := tmpFile.Close(); err != nil {
		return "", fmt.Errorf("error closing file %s: %s", filePath, err)
	}

	return filePath, nil
}

func CreateDir(d string) error {
	stat, err := os.Lstat(d)

	if os.IsNotExist(err) {
		if err := os.MkdirAll(d, os.ModePerm); err != nil {
			return err
		}
	} else if err != nil {
		return err
	} else if stat != nil && !stat.IsDir() {
		return fmt.Errorf("%s already exists and is not a directory", d)
	}

	return nil
}
