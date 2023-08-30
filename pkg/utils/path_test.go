package utils

import (
	"fmt"
	"log"
	"path/filepath"
	"testing"
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
			if (err != nil) != tt.wantErr {
				t.Errorf("ExpandPath() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("ExpandPath() = %v, want %v", got, tt.want)
			}
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
