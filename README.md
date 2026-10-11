# Container Engine

A small Linux container runtime built in Go to explore how container isolation works at the kernel level. This project shows a minimum container implementation that uses Linux namespaces, mount management, chroot, overlay filesystem layering, and cgroup resource limits.

## Overview

Project creates a lightweight isolated environment for running a process in a sandbox root fs. It is designed as an educational prototype only.

The implementation includes:

- Process isolation with PID, network, UTS, mount, and user namespaces
- A chrooted root filesystem using an overlay mount
- A mounted proc filesystem inside the container
- Resource limits via cgroup v2 for memory, CPU, and swap
- A simple CLI to launch commands inside the container

## What was implemented

- A command-line entrypoint in `main.go` that dispatches to `run` and `child` modes
- Parent/child process separation in `container.go`
- Cgroup management logic in `cgroup/cgroup.go`
- A base Alpine fs under `images/alpine/`
- Configure cname in container.go and Manage overlay fs directories with that directory name.

## Architecture

The workflow is roughly:

1. The parent process validates sudo access and creates a cgroup.
2. A child process is spawned with Linux namespace isolation enabled.
3. The container root is prepared with overlay mounts.
4. The process is chrooted into the merged overlay directory.
5. `/proc` is mounted inside the container.
6. The target command is executed inside the isolated environment.

## Project structure

```text
.
├── main.go                  # CLI entry point
├── container.go             # Container setup and execution logic
├── cgroup/
│   └── cgroup.go            # Cgroup v2 helpers
├── images/
│   └── alpine/              # Minimal root filesystem image
├── contain3/                # Overlay mount working directories
├── tests/
│   └── tests.go             # Simple Tests written that were used during debug and memory simulation
├── go.mod                   # Go module definition
└── README.md                # Project documentation
```

## Prerequisites

- Linux host with kernel support for namespaces and overlayfs
- Go installed
- Root or sudo access
- cgroup v2 enabled on the system

## Usage

Run a process inside the container:

```bash
sudo go run . run /bin/sh
```

You can also run other binaries from the Alpine image, for example:

```bash
sudo go run . run /bin/ls -l /
```

You can also use the pushed binary [engine](./cmd/engine) and run this directly without go build/run

```bash
sudo ./cmd/engine run /bin/sh    # or whatever u want container to do
```

## Resource limits

I have basic control limits (using /cgroup inside the kernel fs) to the container:

- Memory limit: 50 MB
- CPU limit: 50%
- Swap limit: disabled

These values are defined in `container.go` and managed via the cgroup helpers.

## Notes

Repo is a simple learning project for understanding how container runtimes work internally. It is not a complete, production-ready container engine and does not aim to match the complexity of Docker, containerd, or runc.

## License

Intended for educational use and experimentation.
