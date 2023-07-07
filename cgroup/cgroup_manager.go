package cgroup

import (
	"github.com/sirupsen/logrus"
	"github.com/summer-boythink/mydocker/cgroup/subsystems"
)

type CgroupManager struct {
	Path string
	Resource *subsystems.ResourceConfig
}

func NewManager(path string) *CgroupManager {
	return &CgroupManager{
		Path:path,
	}
}

func (c *CgroupManager) Apply(pid int) error {
	for _, subSysIns := range subsystems.SubSystemIns {
		err := subSysIns.Apply(c.Path, pid)
		if err != nil {
			return err
		}
	}
	return nil
}

// Set 设置cgroup资源限制
func (c *CgroupManager) Set(res *subsystems.ResourceConfig) error {
	for _, subSysIns := range subsystems.SubSystemIns {
		err := subSysIns.Set(c.Path, res)
		if err != nil {
			return err
		}
	}
	return nil
}

// Destroy 释放cgroup
func (c *CgroupManager) Destroy() error {
	for _, subSysIns := range subsystems.SubSystemIns {
		if err := subSysIns.Remove(c.Path); err != nil {
			logrus.Warnf("remove cgroup fail %v", err)
		}
	}
	return nil
}