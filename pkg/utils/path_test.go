package utils

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestExpandPath(t *testing.T) {
	home := getHomeDir(t)

	tests := []struct {
		name    string
		path    string
		want    string
		wantErr bool
	}{
		{"empty", "", "", false},
		{"non existing user", "~NONEXISTING/", fmt.Sprintf("%sNONEXISTING", home), false},
		{"home's subdir", "~/subdir", filepath.Join(home, "subdir"), false},
		{"any path", "/patgh/to/file", "/patgh/to/file", false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t2 *testing.T) {
			got, err := ExpandPath(tt.path)

			if tt.want == "" {
				curdir, err := os.Getwd()
				if err != nil {
					t2.Skipf("couldn't2 get the current working directory")
				}

				tt.want = curdir
			}

			if tt.wantErr {
				assert.NotNil(t2, err)
				return
			}

			assert.Nil(t2, err)
			assert.Equal(t2, tt.want, got)
		})
	}
}

func TestExpandPathPOSIX(t *testing.T) {
	home := getHomeDir(t)

	tests := []struct {
		name    string
		path    string
		want    string
		wantErr bool
	}{
		{"empty", "", "", false},
		{"non existing user", "~NONEXISTING/", "~NONEXISTING/", false},
		{"home's subdir", "~/subdir", filepath.Join(home, "subdir"), false},
		{"any path", "/patgh/to/file", "/patgh/to/file", false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t2 *testing.T) {
			got, err := ExpandPathPOSIX(tt.path)
			if tt.wantErr {
				assert.NotNil(t2, err)
				return
			}

			assert.Nil(t2, err)
			assert.Equal(t2, tt.want, got)
		})
	}
}

func ExampleExpandPathPOSIX() {
	p, err := ExpandPathPOSIX("~/test")
	if err != nil {
		log.Fatal(err)
	}

	_, last := filepath.Split(p)
	fmt.Println(last)

	p, _ = ExpandPathPOSIX("~NONEXISTINGUSER/path/to/file")
	fmt.Println(p)

	p, _ = ExpandPathPOSIX("/path/to/file/tilde/~")
	fmt.Println(p)

	p, _ = ExpandPathPOSIX("")
	fmt.Println(p)

	p, _ = ExpandPathPOSIX("")
	fmt.Println(p)

	// Output:
	// test
	// ~NONEXISTINGUSER/path/to/file
	// /path/to/file/tilde/~
	//
}

func getHomeDir(t *testing.T) string {
	t.Helper()

	home, err := os.UserHomeDir()
	if err != nil {
		t.Skipf("couldn't get $HOME")
	}

	return home
}
