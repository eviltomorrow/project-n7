package fs

import (
	"fmt"
	"os"
)

func CreateDir(dir string) error {
	fi, err := os.Stat(dir)
	if err != nil {
		if !os.IsNotExist(err) {
			return err
		}
		return os.MkdirAll(dir, 0755)
	}
	if !fi.IsDir() {
		return fmt.Errorf("already exist same file, path: %v", dir)
	}
	return nil
}
