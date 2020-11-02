package cache

import (
	"fmt"
	"os/user"
	"path/filepath"
)

func userCache() (string, error) {
	usr, err := user.Current()
	if err != nil {
		return "", fmt.Errorf("Could not find the current user")
	}
	return filepath.Join(usr.HomeDir, ".cache"), nil
}
