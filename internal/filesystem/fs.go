package filesystem

import (
	"os/user"
	"path/filepath"
)

func userHome() (string, error) {
	usr, err := user.Current()
	if err != nil {
		return "", err
	}
	return usr.HomeDir, nil
}

func userCache() (string, error) {
	usr, err := userHome()
	if err != nil {
		return "", err
	}
	return filepath.Join(usr, ".cache"), nil
}
