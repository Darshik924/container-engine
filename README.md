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
