package main

import (
	"container-engine/cgroup"
	"fmt"
	"os"
	"os/exec"
	"strconv"
	"syscall"
)

func Parent() (err error) {
	// parent execution here
	if len(os.Args) < 3 {
		return fmt.Errorf("usage: run <command> [args...]")
	}
	
	uid, err := strconv.Atoi(os.Getenv("SUDO_UID"))
	if err != nil {
		return fmt.Errorf("invalid SUDO_UID: %w", err)
	}
	gid, err := strconv.Atoi(os.Getenv("SUDO_GID")) // 27 on Ubuntu Linux
	if err != nil {
		return fmt.Errorf("invalid SUDO_GID: %w", err)
	}

	fmt.Printf("Parent with PID: %d\n", os.Getpid())

	cg := cgroup.GetNewManager(cname)

	if err := cg.Create(); err != nil {
		// This method creates the directory inside the kernel managed system /sys/fs/cgroup
		return err
	}

	defer func() {
		if cleanupErr := cg.DeleteGroup(); cleanupErr != nil && err == nil {
			err = cleanupErr
		}
	}()

	childArgs := append([]string{"child"}, os.Args[2:]...)
	cmd := exec.Command("/proc/self/exe", childArgs...)
	// self execute using required agrs

	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	cmd.SysProcAttr = &syscall.SysProcAttr{
		Cloneflags: syscall.CLONE_NEWUSER | syscall.CLONE_NEWPID | syscall.CLONE_NEWUTS | syscall.CLONE_NEWNET | syscall.CLONE_NEWNS,
		Credential: &syscall.Credential{Uid: 0, Gid: 0}, // calls setuid(0), and setgid(0) so that child's actual UID / GID become 0,0 in the namespace
		/* The child was reaching chroot but getting EPERM, so the sample wasn’t starting. Here we want to run the child as container UID/GID 0.*/

		// Writes /proc/<pid>/uid_map and /proc/<pid>/gid_map after clone() but before the child runs. This is the kernel's mechanism to say "container UID 0, host UID uid"
		UidMappings: []syscall.SysProcIDMap{{ContainerID: 0, HostID: uid, Size: 1}},
		GidMappings: []syscall.SysProcIDMap{{ContainerID: 0, HostID: gid, Size: 1}},
		GidMappingsEnableSetgroups: false, // no supplementary groups here as this was failing with EPERM.
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

	// When we start the process, do that directly inside cgroup beforehand to make the logically flow correct instead of starting it first and then moving it inside, otherwise memory test fails.
	cgDir, err := os.Open(cgroup.DefaultRoot + "/" + cname)

	if err != nil {
		return err
	}
	defer cgDir.Close()

	cmd.SysProcAttr.UseCgroupFD = true
	cmd.SysProcAttr.CgroupFD = int(cgDir.Fd())
	// File descriptor for the cgroup is dir Fd

	if err := cmd.Start(); err != nil {
		return fmt.Errorf("error: could not start the child execution: %v\n", err)
	}

	errCaused := cmd.Wait()

	if errCaused != nil {
		return fmt.Errorf("error: child execution failed: %w", errCaused)
	}

	return nil
}
