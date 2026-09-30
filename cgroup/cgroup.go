package cgroup

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"
)

const DefaultRoot = "/sys/fs/cgroup"

type ControlManager struct {
	root string
	path string
	Name string
}

// cg.GetNewManager("container-name") will establish a cg state for us
func GetNewManager(cname string) *ControlManager {
	return &ControlManager{
		root: DefaultRoot,
		path: filepath.Join(DefaultRoot, cname),
		Name: cname,
	}
}

func (m *ControlManager) CheckExists() error {
	info, err := os.Stat(m.root)
	if err != nil {
		return fmt.Errorf("error: cgroup root not available: %w\n", err)
	}

	if !info.IsDir() {
		return fmt.Errorf("fatal: %s is not a directory\n", m.root)
	}
	controllers := filepath.Join(m.root, "cgroup.controllers")

	if _, err := os.Stat(controllers); err != nil {
		return fmt.Errorf("error: cgroup v2 does not appear to be mounted at %s: %w\n", m.root, err)
	}
	return nil
}

// cg.Create will perform the sudo mkdir /sys/fs/cgroup/container-name
func (m *ControlManager) Create() error {
	if err := os.Mkdir(m.path, 0755); err != nil {
		return fmt.Errorf("error: could not create a control group (directory inside the /sys/fs/cgroup ): %w\n", err)
	}
	return nil
}

// Helper to write to files of /sys/fs/cgroup/container-name
func (m *ControlManager) WriteFile(fname string, value string) error {
	pathh := filepath.Join(m.path, fname)

	if err := os.WriteFile(pathh, []byte(value), 0644); err != nil {
		return fmt.Errorf("fatal: could not write %s: %w\n", value, err)
	}
	return nil
}

// usage: cg.SetMemoryLimit(50 * 1024 * 1024) --->> That is 50 MBs
func (m *ControlManager) SetMemoryLimit(bytes int64) error {
	return m.WriteFile("memory.max",strconv.FormatInt(bytes, 10),)
}

// cg.SetCPU(50) writes a 50000 100000
func (m *ControlManager) SetCPU(percent int) error {
	if percent <= 0 || percent >= 100 {
		return fmt.Errorf("error: percent cannot be invalid and must be in between 0 and 100\n");
	}
	period := 100000
	quota := period * percent / 100
	value := fmt.Sprintf("%d %d", quota, period)
	return m.WriteFile("cpu.max", value)
}

// Removal of the group
func (m *ControlManager) DeleteGroup() error {
	if err := os.Remove(m.path); err != nil {
		return fmt.Errorf("error: could not delete the cgroup: %w", err)
	}
	return nil
}

// Important Method to add a Process in the container
func (m *ControlManager) AddProcess(pid int) error {
	return m.WriteFile("cgroup.procs",strconv.Itoa(pid))
}	