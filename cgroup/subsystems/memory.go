package subsystems

import (
	"fmt"
	"os"
	"path"
	"strconv"
)

type MemorySubSystem struct {
}

func (m *MemorySubSystem) Name() string {
	return "memory"
}

func (m *MemorySubSystem) Set(cgroupPath string, res *ResourceConfig) error {
	if sysCgroupPath, err := GetCgroupPath(m.Name(), cgroupPath, true); err == nil {
		if res.MemoryLimit != "" {
			if err := os.WriteFile(path.Join(sysCgroupPath, "memory.limit_in_bytes"),
				[]byte(res.MemoryLimit), 0644); err != nil {
				return fmt.Errorf("set memory fail %v", err)
			}
		}
		return nil
	} else {
		return err
	}
}

func (m *MemorySubSystem) Apply(cgroupPath string, pid int) error {
	if sysCgroupPath, err := GetCgroupPath(m.Name(), cgroupPath, true); err == nil {
		if err := os.WriteFile(path.Join(sysCgroupPath, "tasks"),
			[]byte(strconv.Itoa(pid)), 0644); err != nil {
			return fmt.Errorf("set cgroup proc fail %v", err)
		}
		return nil
	} else {
		return err
	}
}

func (m *MemorySubSystem) Remove(cgroupPath string) error {
	if sysCgroupPath, err := GetCgroupPath(m.Name(), cgroupPath, true); err == nil {
		return os.Remove(sysCgroupPath)
	} else {
		return err
	}
}
