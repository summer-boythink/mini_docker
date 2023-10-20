package main

import (
	"encoding/json"
	"fmt"
	"math/rand"
	"os"
	"strconv"
	"strings"
	"time"

	log "github.com/sirupsen/logrus"
	"github.com/summer-boythink/mydocker/cgroup"
	"github.com/summer-boythink/mydocker/cgroup/subsystems"
	"github.com/summer-boythink/mydocker/container"
)

func Run(tty bool, commands []string, resConf *subsystems.ResourceConfig, volume string, containerName string, imageName string) {
	id := randStringBytes(10)
	if containerName == "" {
		containerName = id
	}
	parent, writePipe := container.NewParentProcess(tty, volume, containerName, imageName)
	if parent == nil {
		log.Errorf("New parent process error")
		return
	}
	if err := parent.Start(); err != nil {
		log.Error(err)
	}

	containerName, err := recordContainerInfo(parent.Process.Pid, commands, containerName, id, volume)
	if err != nil {
		log.Errorf("record err:%v", err)
		return
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
	if tty {
		parent.Wait()
		deleteContainerInfo(containerName)
		container.DeleteWorkSpace(volume, containerName)
	}
	// container.DeleteWorkSpace(volume, containerName)
	os.Exit(0)
}

func deleteContainerInfo(name string) {
	dirURL := fmt.Sprintf(container.DefaultInfoLocation, name)
	if err := os.RemoveAll(dirURL); err != nil {
		log.Errorf("Remove dir %s error %v", dirURL, err)
	}
}

func sendInitCommand(commands []string, writePipe *os.File) {
	command := strings.Join(commands, " ")
	log.Infof("command all is %s", command)
	writePipe.WriteString(command)
	writePipe.Close()
}

func recordContainerInfo(containerPid int, commandArr []string, containerName string, id string, volume string) (string, error) {
	createTime := time.Now().Format("2006-01-02 15:04:05")
	commands := strings.Join(commandArr, "")
	containerInfo := container.Info{
		Pid:         strconv.Itoa(containerPid),
		Id:          id,
		Name:        containerName,
		Command:     commands,
		CreatedTime: createTime,
		Status:      container.RUNNING,
		Volume:      volume,
	}

	jsonBytes, err := json.Marshal(containerInfo)
	if err != nil {
		log.Errorf("Record container info error %v", err)
		return "", err
	}
	jsonStr := string(jsonBytes)
	dirUrl := fmt.Sprintf(container.DefaultInfoLocation, containerName)
	if err := os.MkdirAll(dirUrl, 0622); err != nil {
		log.Errorf("Mkdir error path %s,error:%s", dirUrl, err)
		return "", err
	}
	fileName := dirUrl + "/" + container.ConfigName
	file, err := os.Create(fileName)
	defer func(file *os.File) {
		err := file.Close()
		if err != nil {
			panic(err)
		}
	}(file)
	if err != nil {
		log.Errorf("Create file %s error %v", fileName, err)
		return "", err
	}
	if _, err := file.WriteString(jsonStr); err != nil {
		log.Errorf("File write string error %v", err)
		return "", err
	}

	return containerName, nil
}

func randStringBytes(n int) string {
	letterBytes := "1234567890"
	rand.Seed(time.Now().UnixNano())
	b := make([]byte, n)
	for i := range b {
		b[i] = letterBytes[rand.Intn(len(letterBytes))]
	}
	return string(b)
}
