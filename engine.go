package main

import (
	"fmt"
	"os"
) 

func main() {
	if !(os.Getenv("SUDO_USER") != "") {
		fmt.Printf("please run as sudo user\n")
		os.Exit(1)
	}

	if len(os.Args) < 2 {
		fmt.Printf("incorrect usage: go run engine.go run <command>\n")
		os.Exit(1)
	}

	var err error
	switch os.Args[1] {
	case "run":
		err = Parent()
	case "child":
		err = Child()
	default:
		fmt.Printf("incorrect usage: run <command>\n")
		err = nil
	}
	
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
