package database

import "os"

func mkdir(path string) error {
	if path == "" {
		return os.ErrInvalid
	}
	return os.MkdirAll(path, 0o750)
}
