package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"syscall"

	"golang.org/x/sys/unix"
)

func Child() (retErr error) {
	fmt.Printf("Child envorioment with PID: %d\n", os.Getpid())
	
	// Container Now believes its PID is 1 but it is actually something else for the Host machine

    // cDir is the container dir inside out rootfs
	cDir := filepath.Join(rootfs, cname)

    upper := filepath.Join(cDir, "upperdir")
    work := filepath.Join(cDir, "workdir")
    merged := filepath.Join(cDir, "merged")

    lowerdir := os.Getenv("ALPINE_ROOT")
    if lowerdir == "" {
        lowerdir = filepath.Join(rootfs, "images", "alpine")
		// Default
    }

	/* Make the directories */
    for _, dir := range []string{upper, work, merged} {
        if err := os.MkdirAll(dir, 0755); err != nil {
            return fmt.Errorf("create directory failed %q: %w", dir, err)
        }
    }

	oldRoot, err := os.Open("/")
    if err != nil {
        return fmt.Errorf("could not open original root: %w", err)
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

	// mark host system as MS_PRIVATE, this prevents mount namespace events inside the container from moving back to the host OS
	if err := unix.Mount("", "/", "", unix.MS_REC|unix.MS_PRIVATE, ""); err != nil {
        return fmt.Errorf("could not make mounts private: %w", err)
    }

	// Mount overlay - Linux Kernel mounts the virtual overlay fs onto merged.
	opts := fmt.Sprintf("lowerdir=%s,upperdir=%s,workdir=%s", lowerdir, upper, work)
    if err := unix.Mount("overlay", merged, "overlay", 0, opts); err != nil {
        // Mount syscall equivalent to mount -t overlay.... shell command
        return fmt.Errorf("OverlayFS mount has error: %w", err)
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
        return fmt.Errorf("create /proc error: %w", err)
    }
	// Mount a fresh proc (process) fs inside the new container's /proc, else the container inherits the host's proc: ps shows all the processes
    if err := unix.Mount("proc", "/proc", "proc", 0, ""); err != nil {
        return fmt.Errorf("mount /proc error: %w", err)
    }
	procMounted = true

	if len(os.Args) < 3 {
        return fmt.Errorf("usage: run <command> args...")
    }

	targetCmd := os.Args[2]
	targetArgs := os.Args[3:]

	cmd := exec.Command(targetCmd, targetArgs...)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("could not run the child command (inside child): %w\n", err)
	}
	
    defer os.RemoveAll(cDir) // cleanup all directories 
	return nil
}


