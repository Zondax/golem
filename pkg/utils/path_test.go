package utils

import (
	"fmt"
	"log"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestExpandPath(t *testing.T) {
	tests := []struct {
		name    string
		path    string
		want    string
		wantErr bool
	}{
		{"empty", "", "", false},
		{"non existing user", "~NONEXISTING/", "~NONEXISTING/", false},
		{"any path", "/patgh/to/file", "/patgh/to/file", false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ExpandPath(tt.path)
			if tt.wantErr {
				assert.NotNil(t, err)
				return
			}

			assert.Nil(t, err)
			assert.Equal(t, tt.want, got)
		})
	}
}

func ExampleExpandPath() {
	p, err := ExpandPath("~/test")
	if err != nil {
		log.Fatal(err)
	}

	_, last := filepath.Split(p)
	fmt.Println(last)

	p, _ = ExpandPath("~NONEXISTINGUSER/path/to/file")
	fmt.Println(p)

	p, _ = ExpandPath("/path/to/file/tilde/~")
	fmt.Println(p)

	p, _ = ExpandPath("")
	fmt.Println(p)

	p, _ = ExpandPath("")
	fmt.Println(p)

	// Output:
	// test
	// ~NONEXISTINGUSER/path/to/file
	// /path/to/file/tilde/~
	//
}
