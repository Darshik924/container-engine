package main

import (
	"fmt"
	"os"
) 

func main()  {
	if len(os.Args) < 2 {
		fmt.Printf("error usage: go run engine.go run <command>\n")
		os.Exit(1)
	}

	switch os.Args[1] {
	case "run":
		Parent()
	case "child":
		Child()
	}
}

func Child()  {
	// child goes here
	fmt.Printf("[Child] Initializing container envorioment with PID: %d", os.Getpid())
}

func Parent()  {
	// parent goes down here
	fmt.Printf("[Parent] Initializing Parent with PID: %d", os.Getpid())
}