package main

import (
	log "github.com/sirupsen/logrus"
	"github.com/summer-boythink/mydocker/cgroup"
	"github.com/summer-boythink/mydocker/cgroup/subsystems"
	"github.com/summer-boythink/mydocker/container"
	"github.com/summer-boythink/mydocker/devconst"
	"os"
	"strings"
)

func Run(tty bool, commands []string, resConf *subsystems.ResourceConfig, volume string) {
	parent, writePipe := container.NewParentProcess(tty, volume)
	if parent == nil {
		log.Errorf("New parent process error")
		return
	}
	if err := parent.Start(); err != nil {
		log.Error(err)
	}

	cgroupManager := cgroup.NewManager("my-group")
	defer func(cgroupManager *cgroup.CgroupManager) {
		err := cgroupManager.Destroy()
		if err != nil {
			log.Error(err)
		}
	}(cgroupManager)
	cgroupManager.Set(resConf)
	cgroupManager.Apply(parent.Process.Pid)
	sendInitCommand(commands, writePipe)
	parent.Wait()
	container.DeleteWorkSpace(devconst.RootURL, devconst.MntURL, volume)
	os.Exit(0)
}

func sendInitCommand(commands []string, writePipe *os.File) {
	command := strings.Join(commands, " ")
	log.Infof("command all is %s", command)
	writePipe.WriteString(command)
	writePipe.Close()
}
