package main

import (
	"container-engine/cgroup"
	"golang.org/x/sys/unix"
	"fmt"
	"os"
	"os/exec"
	"syscall"
) 

const rootfs = "~/Dev/container-engine"
const MemoryLim = 50 * 1024 * 1024 // that is 50MBs
const CPULim = 50 // 50 percentage 

func main()  {
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
		fmt.Printf("incorrect usage: go run engine.go run <command>\n")
		err = nil
	}
	
	if err != nil {
		fmt.Print(err)
	}
}

func Child() error {
	// child goes here
	fmt.Printf("Child envorioment with PID: %d\n", os.Getpid())

	// Container Now believes its PID is 1 but it is actually something else for the Host machine

	/* Child feels like it is isolated due to NEW PID, NETWORK isolation, UTS Namespace it cannot really see so it is isolated */
	if err := syscall.Chroot(rootfs); err != nil {
		fmt.Printf("chroot: %v\n", err)
		return err
	}

	if err := os.Chdir("/"); err != nil {
		fmt.Printf("chdir: %v\n", err)
		return err 	
	}

	os.MkdirAll("/proc", 0555)
	unix.Mount("proc", "/proc", "proc", 0, "")

	targetCmd := os.Args[2]
	targetArgs := os.Args[3:]

	cmd := exec.Command(targetCmd, targetArgs...)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("could not run the child command: %w", err)
	}
	return nil
}

func Parent() error {
	// parent execution here
	if len(os.Args) < 3 {
		return fmt.Errorf("usage: go run engine.go run <command> [args...]")
	}

	fmt.Printf("Parent with PID: %d\n", os.Getpid())

	pPid := os.Getpid()
	// the parent process ID
	cg := cgroup.GetNewManager("mini-cont")

	if err := cg.Create(); err != nil {
		// This creates the directory inside the kernel managed system /sys/fs/cgroup
		return err
	}

	// Add a process is actually writing the process ID in cgroup.procs
	if err := cg.AddProcess(pPid); err != nil {
		return err
	}

	childArgs := append([]string{"child"}, os.Args[2:]...)
	cmd := exec.Command("/proc/self/exe", childArgs...)
	// Load the child process arguments and exec the required command

	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	cmd.SysProcAttr = &syscall.SysProcAttr{
		Cloneflags: syscall.CLONE_NEWUSER | syscall.CLONE_NEWPID | syscall.CLONE_NEWUTS | syscall.CLONE_NEWNET | syscall.CLONE_NEWNS,
	}

	if err := cg.SetMemoryLimit(MemoryLim); err != nil {
		return err
	}
	if err := cg.SetCPU(CPULim); err != nil {
		return err
	}

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("error: could not start the child execution: %v\n", err)
	}

	// executed here - remove the cgroup 
	if err := cg.DeleteGroup(); err != nil {
		return err
	}
	return nil
}