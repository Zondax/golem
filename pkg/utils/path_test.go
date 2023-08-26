package utils

import (
	"fmt"
	"log"
	"path/filepath"
)

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
