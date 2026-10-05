package main

import (
	"container-engine/cgroup"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"syscall"

	"golang.org/x/sys/unix"
)

const rootfs = "/home/darshik/Dev/container-engine"
const StateDir = "/home/darshik/Dev/container-engine"
/* rootfs should change here to be the alpine image */
const MemoryLim = 50 * 1024 * 1024 // that is 50MBs
const CPULim = 50 // 50 percentage 
const cname = "contain3"

func Child() (retErr error) {
	// child goes here
	fmt.Printf("Child envorioment with PID: %d\n", os.Getpid())

	// Container Now believes its PID is 1 but it is actually something else for the Host machine
	containerDir := filepath.Join(StateDir, cname)

    upper := filepath.Join(containerDir, "upperdir")
    work := filepath.Join(containerDir, "workdir")
    merged := filepath.Join(containerDir, "merged")

    lower := os.Getenv("ALPINE_ROOT")
    if lower == "" {
        lower = filepath.Join(StateDir, "images", "lowerdir")
    }
    lower, err := filepath.Abs(lower)
    if err != nil {
        return fmt.Errorf("resolve Alpine root path: %w", err)
    }

    for _, dir := range []string{upper, work, merged} {
        if err := os.MkdirAll(dir, 0755); err != nil {
            return fmt.Errorf("create directory %q: %w", dir, err)
        }
    }

    info, err := os.Stat(lower)
    if err != nil {
        return fmt.Errorf("Alpine root filesystem %q not found; set ALPINE_ROOT to its extracted directory: %w", lower, err)
    }
    if !info.IsDir() {
        return fmt.Errorf("Alpine root filesystem %q is not a directory", lower)
    }

	oldRoot, err := os.Open("/")
    if err != nil {
        return fmt.Errorf("open original root: %w", err)
    }
    defer oldRoot.Close()

    overlayMounted := true
    chrooted := false
    procMounted := false

	// cleanup function
	defer func() {
        setCleanupErr := func(err error) {
            if err != nil && retErr == nil {
                retErr = err
            }
        }

        if procMounted {
            if err := unix.Unmount("/proc", 0); err != nil {
                setCleanupErr(fmt.Errorf("unmount /proc: %w", err))
            }
        }

        if chrooted {
            if err := unix.Fchdir(int(oldRoot.Fd())); err != nil {
                setCleanupErr(fmt.Errorf("restore original directory: %w", err))
                return
            }
            if err := syscall.Chroot("."); err != nil {
                setCleanupErr(fmt.Errorf("restore original root: %w", err))
                return
            }
            chrooted = false
            if err := os.Chdir("/"); err != nil {
                setCleanupErr(fmt.Errorf("change to original root: %w", err))
                return
            }
        }

        if overlayMounted {
            if err := unix.Unmount(merged, 0); err != nil {
                setCleanupErr(fmt.Errorf("unmount overlay: %w", err))
            }
        }
    }()

	if err := unix.Mount("", "/", "", unix.MS_REC|unix.MS_PRIVATE, ""); err != nil {
        return fmt.Errorf("make mounts private: %w", err)
    }

	opts := fmt.Sprintf("lowerdir=%s,upperdir=%s,workdir=%s", lower, upper, work)
    if err := unix.Mount("overlay", merged, "overlay", 0, opts); err != nil {
        return fmt.Errorf("mount overlay: %w", err)
    }
    overlayMounted = true
    
	/* Child feels like it is isolated due to NEW PID, NETWORK isolation, UTS Namespace it cannot really see so it is isolated */
	if err := syscall.Chroot(merged); err != nil {
		return fmt.Errorf("chroot to %q: %w", rootfs, err)
	}
    chrooted = true

	if err := os.Chdir("/"); err != nil {
		return err 	
	}
	// Child Believes that its Directory is /

	if err := os.MkdirAll("/proc", 0555); err != nil {
        return fmt.Errorf("create /proc: %w", err)
    }
    if err := unix.Mount("proc", "/proc", "proc", 0, ""); err != nil {
        return fmt.Errorf("mount /proc: %w", err)
    }
	procMounted = true
	if len(os.Args) < 3 {
        return fmt.Errorf("usage: run <command> [args...]")
    }

	targetCmd := os.Args[2]
	targetArgs := os.Args[3:]

	cmd := exec.Command(targetCmd, targetArgs...)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("could not run the child command: %w\n", err)
	}
	
    defer os.RemoveAll(containerDir) // cleanup all directories 
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

	// Start the process directly inside cgroup beforehand to make the logically flow correct instead of starting it first and then moving it inside
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

	waitErr := cmd.Wait()

	if waitErr != nil {
		return fmt.Errorf("error: child execution failed: %w", waitErr)
	}

	return nil
}
