package main

import (
	"container-engine/cgroup"
	"fmt"
	"os"
	"os/exec"
	"strconv"
	"syscall"

	"golang.org/x/sys/unix"
)

const rootfs = "/home/darshik/Dev/container-engine"
const MemoryLim = 50 * 1024 * 1024 // that is 50MBs
const CPULim = 50 // 50 percentage 
const cname = "contain2"

func Child() error {
	// child goes here
	fmt.Printf("Child envorioment with PID: %d\n", os.Getpid())

	// Container Now believes its PID is 1 but it is actually something else for the Host machine

	/* Child feels like it is isolated due to NEW PID, NETWORK isolation, UTS Namespace it cannot really see so it is isolated */
	if err := syscall.Chroot(rootfs); err != nil {
		return fmt.Errorf("chroot to %q: %w", rootfs, err)
	}

	if err := os.Chdir("/"); err != nil {
		return err 	
	}
	// Child Believes that its Directory is /

	os.MkdirAll("/proc", 0555)
	unix.Mount("proc", "/proc", "proc", 0, "")

	targetCmd := os.Args[2]
	targetArgs := os.Args[3:]

	cmd := exec.Command(targetCmd, targetArgs...)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("could not run the child command: %w\n", err)
	}
	return nil
}

func Parent() (err error) {
	// parent execution here
	if len(os.Args) < 3 {
		return fmt.Errorf("usage: run <command> [args...]")
	}
	
	uid, err := strconv.Atoi(os.Getenv("SUDO_UID"))
	if err != nil {
		return fmt.Errorf("invalid SUDO_UID: %w", err)
	}
	gid, err := strconv.Atoi(os.Getenv("SUDO_GID"))
	if err != nil {
		return fmt.Errorf("invalid SUDO_GID: %w", err)
	}
	fmt.Printf("Parent with PID: %d\n", os.Getpid())

	cg := cgroup.GetNewManager(cname)

	if err := cg.Create(); err != nil {
		// This creates the directory inside the kernel managed system /sys/fs/cgroup
		return err
	}

	defer func() {
		if cleanupErr := cg.DeleteGroup(); cleanupErr != nil && err == nil {
			err = cleanupErr
		}
	}()

	childArgs := append([]string{"child"}, os.Args[2:]...)
	cmd := exec.Command("/proc/self/exe", childArgs...)
	// Load the child process arguments and exec the required command

	// Redirect the file descriptors
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	cmd.SysProcAttr = &syscall.SysProcAttr{
		Cloneflags: syscall.CLONE_NEWUSER | syscall.CLONE_NEWPID | syscall.CLONE_NEWUTS | syscall.CLONE_NEWNET | syscall.CLONE_NEWNS,
		Credential: &syscall.Credential{Uid: 0, Gid: 0},
		/* The child was reaching chroot but getting EPERM, so the sample wasn’t starting. I updated container.go to explicitly run the child as container UID/GID 0, matching the existing namespace maps, and made chroot errors include the attempted path. */
		UidMappings: []syscall.SysProcIDMap{{ContainerID: 0, HostID: uid, Size: 1}},
		GidMappings: []syscall.SysProcIDMap{{ContainerID: 0, HostID: gid, Size: 1}},
		GidMappingsEnableSetgroups: false,
	}

	if err := cg.SetMemoryLimit(MemoryLim); err != nil {
		return err
	}
	if err := cg.SetCPU(CPULim); err != nil {
		return err
	}

	if err := cg.SetSwapLimit("0"); err != nil {
		return err
	}

	cgDir, err := os.Open(cgroup.DefaultRoot + "/" + cname)
	if err != nil {
		return err
	}
	defer cgDir.Close()

	cmd.SysProcAttr.UseCgroupFD = true
	cmd.SysProcAttr.CgroupFD = int(cgDir.Fd())

	if err := cmd.Start(); err != nil {
		return fmt.Errorf("error: could not start the child execution: %v\n", err)
	}

	waitErr := cmd.Wait()

	if waitErr != nil {
		return fmt.Errorf("error: child execution failed: %w", waitErr)
	}

	return nil
}