package main

import (
	"fmt"
	"os"
	"os/exec"
	"syscall"
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
	fmt.Printf("Initializing container envorioment with PID: %d\n", os.Getpid())

	targetCmd := os.Args[2]
	targetArgs := os.Args[3:]

	cmd := exec.Command(targetCmd, targetArgs...)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		fmt.Printf("error: could not run the child commands: %v\n", err)
		os.Exit(1)
	}
}

func Parent()  {
	// parent goes down here
	fmt.Printf("Initializing Parent with PID: %d\n", os.Getpid())

	args := append([]string{"child"}, os.Args[2:]...)
	cmd := exec.Command("/proc/self/exe", args...)
	// Load the child process arguments and exec the required command

	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	cmd.SysProcAttr = &syscall.SysProcAttr{
		Cloneflags: syscall.CLONE_NEWUSER | syscall.CLONE_NEWPID | syscall.CLONE_NEWUTS | syscall.CLONE_NEWNET,
	}
	/* CLONE_NEWUSER is the only namespace flag that doesn't require root. Once the child is inside a user namespace, it gets CAP_SYS_ADMIN within that namespace, which is enough to create the other namespaces (mount, PID, net, etc.) */

	if err := cmd.Run(); err != nil {
		fmt.Printf("error: could not start the child execution: %v\n", err)
		os.Exit(1)
	}
}