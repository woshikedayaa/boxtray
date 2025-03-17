package common

import (
	"net/url"
	"os"
	"path/filepath"
	"strings"
)

func CombineArgs(values ...url.Values) url.Values {
	if len(values) == 0 {
		return nil
	}
	if len(values) == 1 {
		return values[0]
	}
	ret := url.Values{}
	// shit code
	for _, value := range values {
		if value == nil {
			continue
		}
		for k, v := range value {
			for _, vv := range v {
				ret.Add(k, vv)
			}
		}
	}
	return ret
}

// ExpandHomePath replaces the ~ symbol in paths with the user's home directory
func ExpandHomePath(path string) (string, error) {
	if !strings.HasPrefix(path, "~") {
		return path, nil
	}

	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}

	if path == "~" {
		return home, nil
	}

	if strings.HasPrefix(path, "~/") {
		return filepath.Join(home, path[2:]), nil
	}

	// Handle the ~username/path case
	// Note: Go standard library cannot directly get other users' home directories
	// This case requires special handling, typically using /home/username on Unix systems
	return path, nil
}
