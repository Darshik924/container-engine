/_ CLONE_NEWUSER is the only namespace flag that doesn't require root. Once the child is inside a user namespace, it gets CAP_SYS_ADMIN within that namespace, which is enough to create the other namespaces (mount, PID, net, etc.) _/

                 Your Go program
                       │
                       ▼
              Create cgroup
                       │
              ┌────────┴────────┐
              ▼                 ▼
        memory.max          cpu.max
          100 MB              50%
              │                 │
              └────────┬────────┘
                       ▼
                 Start process
                       │
                       ▼
                 Get its PID
                       │
                       ▼
              cgroup.procs
                       │
                       ▼
               Process is limited



cg.AddProcess(pPid) → parent joins mini-cont cgroup
cmd.Run() → forks child → child inherits cgroup membership
CLONE_NEW* flags → child gets new PID/UTS/net/mnt/user namespaces
The CLONE_NEWPID flag does not affect cgroup membership — namespaces and cgroups are independent mechanisms. The child is in the cgroup regardless.


┌─────────────────────────────────────────────────────────────────┐
│ 1. CREATE                                                        │
│    mkdir upper/ work/ merged/                                    │
│    (lowerdir already exists from image extraction)               │
├─────────────────────────────────────────────────────────────────┤
│ 2. MOUNT                                                         │
│    mount -t overlay overlay                                      │
│      -o lowerdir=/images/alpine/layer2:/images/alpine/layer1,   │
│         upperdir=/containers/abc123/upper,                       │
│         workdir=/containers/abc123/work                          │
│      /containers/abc123/merged                                   │
│                                                                  │
│    Now merged/ shows the full Alpine filesystem.                 │
├─────────────────────────────────────────────────────────────────┤
│ 3. RUN                                                           │
│    • Enter new mount namespace                                   │
│    • pivot_root (or chroot) into merged/                         │
│    • Exec the container's entrypoint                             │
│                                                                  │
│    All reads → lowerdir (shared, cached)                         │
│    All writes → upperdir (isolated, per-container)               │
│    Atomic ops → workdir (kernel-internal)                        │
├─────────────────────────────────────────────────────────────────┤
│ 4. STOP / REMOVE                                                 │
│    • Kill container process                                      │
│    • umount /containers/abc123/merged                            │
│    • rm -rf /containers/abc123/upper                             │
│    • rm -rf /containers/abc123/work                              │
│    • rm -rf /containers/abc123/merged                            │
│                                                                  │
│    lowerdir is untouched. Other containers are unaffected.       │
└─────────────────────────────────────────────────────────────────┘   