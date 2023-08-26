package utils

import (
	"fmt"
	"os"
	"os/user"
	"path/filepath"
	"strings"
)

func ExpandPath(path string) (string, error) {
	if len(path) == 0 {
		return path, nil
	}

	if !strings.HasPrefix(path, "~") {
		return filepath.Abs(os.ExpandEnv(path))
	}

	// [...] the characters in the tilde-prefix following the <tilde> are treated
	// as a possible login name from the user database. [...].

	parts := pathSegments(path)
	if len(parts[0]) == 1 {
		// If the login name is null (that is, the tilde-prefix contains only the tilde),
		// the tilde-prefix is replaced by the value of the variable HOME. If HOME is
		// unset, the results are unspecified. [continue]
		home, err := os.UserHomeDir()
		if err != nil {
			return path, err
		}

		path = strings.Replace(path, "~", home, 1)

		return filepath.Abs(os.ExpandEnv(path))
	}

	// Otherwise, the tilde-prefix shall be replaced
	// by a pathname of the initial working directory associated with the login name
	// obtained using the getpwnam() function as defined in the System Interfaces volume
	// of POSIX.1-2017. If the system does not recognize the login name, the results are
	// undefined.

	// treat what follows the tilde as a potentially valid username

	usr, err := user.Lookup(strings.TrimPrefix(parts[0], "~"))
	if err == nil {
		// replace the tilde-prefix with the user's home directory
		return filepath.Abs(os.ExpandEnv(strings.Replace(path, parts[0], usr.HomeDir, 1)))
	}

	switch terr := err.(type) {
	case user.UnknownUserError: // non existing user, move on
	default: // unexpected error
		return path, fmt.Errorf("ExpandPath: got unexpected error %v", terr)
	}

	return os.ExpandEnv(path), nil
}

func pathSegments(path string) []string {
	dir, last := filepath.Split(path)
	if dir == "" {
		return []string{last}
	}
	return append(pathSegments(filepath.Clean(dir)), last)
}
