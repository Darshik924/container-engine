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