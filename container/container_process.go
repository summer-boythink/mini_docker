package container

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
	"syscall"

	log "github.com/sirupsen/logrus"
	"github.com/summer-boythink/mydocker/devconst"
)

func NewParentProcess(tty bool, volume string, containerName string, imageName string) (*exec.Cmd, *os.File) {
	readPipe, writePipe, err := NewPipe()
	if err != nil {
		log.Errorf("New pipe error %v", err)
		return nil, nil
	}
	cmd := exec.Command("/proc/self/exe", "init")
	cmd.SysProcAttr = &syscall.SysProcAttr{
		Cloneflags: syscall.CLONE_NEWUTS | syscall.CLONE_NEWPID | syscall.CLONE_NEWNS |
			syscall.CLONE_NEWNET | syscall.CLONE_NEWIPC,
	}

	if tty {
		cmd.Stdin = os.Stdin
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
	} else {
		dirURL := fmt.Sprintf(DefaultInfoLocation, containerName)
		if err := os.MkdirAll(dirURL, 0622); err != nil {
			log.Errorf("NewParentProcess mkdir %s error %v", dirURL, err)
			return nil, nil
		}
		stdLogFilePath := dirURL + ContainerLogFile
		stdLogFile, err := os.Create(stdLogFilePath)
		if err != nil {
			log.Errorf("NewParentProcess create file %s error %v", stdLogFilePath, err)
			return nil, nil
		}
		cmd.Stdout = stdLogFile
	}
	cmd.ExtraFiles = []*os.File{readPipe}
	NewWorkSpace(volume, imageName, containerName)
	cmd.Dir = fmt.Sprintf(devconst.MntURL, containerName)
	return cmd, writePipe
}

func NewPipe() (*os.File, *os.File, error) {
	read, write, err := os.Pipe()
	if err != nil {
		return nil, nil, err
	}
	return read, write, nil
}

// NewWorkSpace Create a AUFS filesystem
func NewWorkSpace(volume string, imageName string, containerName string) {
	CreateReadOnlyLayer(imageName)
	CreateWriteLayer(containerName)
	CreateTmpLayer(containerName)
	CreateMountPoint(containerName, imageName)

	if volume != "" {
		volumeURLs := strings.Split(volume, ":")
		length := len(volumeURLs)
		if length == 2 && volumeURLs[0] != "" && volumeURLs[1] != "" {
			MountVolume(volumeURLs, containerName)
			log.Infof("%q", volumeURLs)
		} else {
			log.Infof("Volume parameter input is not correct.")
		}
	}
}

func MountVolume(volumeURLs []string, containerName string) {
	// check external url
	parentUrl := volumeURLs[0]
	_, err := os.Stat(parentUrl)
	if os.IsNotExist(err) {
		if err := os.MkdirAll(parentUrl, 0777); err != nil {
			log.Infof("Mkdir parent dir %s error. %v", parentUrl, err)
		}
	}

	// check container url
	containerUrl := volumeURLs[1]
	mntURL := fmt.Sprintf(devconst.MntURL, containerName)
	// writeURL := fmt.Sprintf(devconst.WriteURL, containerName)
	containerVolumeURL := mntURL + containerUrl
	if err := os.MkdirAll(containerVolumeURL, 0777); err != nil {
		log.Infof("Mkdir container dir %s error. %v", containerVolumeURL, err)
	}

	tmpURL := fmt.Sprintf(devconst.TmpURL, containerName)
	// mount dir
	dirs := fmt.Sprintf("lowerdir=%s,upperdir=%s,workdir=%s", containerVolumeURL,
		parentUrl, tmpURL)
	cmd := exec.Command("mount", "-t", "overlay", "-o", dirs, "overlay", containerVolumeURL)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		log.Errorf("Mount volume failed. %v", err)
	}
}

func volumeUrlExtract(volume string) []string {
	var volumeURLs []string
	volumeURLs = strings.Split(volume, ":")
	return volumeURLs
}

func CreateTmpLayer(containerName string) {
	tmpURL := fmt.Sprintf(devconst.TmpURL, containerName)
	if err := os.MkdirAll(tmpURL, 0777); err != nil {
		log.Errorf("Mkdir dir %s error. %v", tmpURL, err)
	}
}

func CreateReadOnlyLayer(imageName string) {
	unTarFolderUrl := devconst.RootURL + imageName + "/"
	imageUrl := devconst.RootURL + imageName + ".tar"
	exist, err := PathExists(unTarFolderUrl)
	if err != nil {
		log.Infof("Fail to judge whether dir %s exists. %v", unTarFolderUrl, err)
	}
	if !exist {
		if err := os.MkdirAll(unTarFolderUrl, 0777); err != nil {
			log.Errorf("Mkdir dir %s error. %v", unTarFolderUrl, err)
		}
		if _, err := exec.Command("tar", "-xvf", imageUrl, "-C", unTarFolderUrl).CombinedOutput(); err != nil {
			log.Errorf("Untar dir %s error %v", unTarFolderUrl, err)
		}
	}
}

func CreateMountPoint(containerName, imageName string) error {
	mntUrl := fmt.Sprintf(devconst.MntURL, containerName)
	if err := os.MkdirAll(mntUrl, 0777); err != nil {
		log.Errorf("Mkdir dir %s error. %v", devconst.MntURL, err)
		return err
	}
	writeURL := fmt.Sprintf(devconst.WriteURL, containerName)
	tmpURL := fmt.Sprintf(devconst.TmpURL, containerName)

	dirs := fmt.Sprintf("lowerdir=%s,upperdir=%s,workdir=%s", devconst.RootURL+imageName+"/",
		writeURL+"/", tmpURL+"/")
	_, err := exec.Command("mount", "-t", "overlay", "-o", dirs, "overlay", mntUrl).CombinedOutput()
	// cmd.Stdout = os.Stdout
	// cmd.Stderr = os.Stderr
	// if err := cmd.Run(); err != nil {
	// 	log.Errorf("%v", err)
	// }
	if err != nil {
		log.Errorf("Run command for creating mount point failed %v", err)
		return err
	}
	return nil
}

func CreateWriteLayer(containerName string) {
	writeURL := fmt.Sprintf(devconst.WriteURL, containerName)
	if err := os.MkdirAll(writeURL, 0777); err != nil {
		log.Errorf("Mkdir dir %s error. %v", writeURL, err)
	}
}

// DeleteWorkSpace Delete the AUFS filesystem while container exit
func DeleteWorkSpace(volume, containerName string) {
	if volume != "" {
		volumeURLs := volumeUrlExtract(volume)
		length := len(volumeURLs)
		if length == 2 && volumeURLs[0] != "" && volumeURLs[1] != "" {
			DeleteMountPointWithVolume(volumeURLs, containerName)
		} else {
			DeleteMountPoint(containerName)
		}
	} else {
		DeleteMountPoint(containerName)
	}
	DeleteWriteLayer(containerName)
	DeleteTmpLayer(containerName)
}

func DeleteMountPointWithVolume(volumeURLs []string, containerName string) {
	mntURL := fmt.Sprintf(devconst.MntURL, containerName)
	containerUrl := mntURL + volumeURLs[1]
	cmd := exec.Command("umount", containerUrl)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		log.Errorf("Umount volume failed. %v", err)
	}

	cmd = exec.Command("umount", mntURL)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		log.Errorf("Umount mountpoint failed. %v", err)
	}

	if err := os.RemoveAll(mntURL); err != nil {
		log.Infof("Remove mountpoint dir %s error %v", mntURL, err)
	}
}

func DeleteTmpLayer(containerName string) {
	tmpURL := fmt.Sprintf(devconst.TmpURL, containerName)
	if err := os.RemoveAll(tmpURL); err != nil {
		log.Errorf("Remove dir %s error %v", tmpURL, err)
	}
}

func DeleteMountPoint(containerName string) error {
	mntURL := fmt.Sprintf(devconst.MntURL, containerName)
	_, err := exec.Command("umount", mntURL).CombinedOutput()
	if err != nil {
		log.Errorf("Unmount %s error %v", mntURL, err)
		return err
	}
	if err := os.RemoveAll(mntURL); err != nil {
		log.Errorf("Remove mountpoint dir %s error %v", mntURL, err)
		return err
	}
	return nil
}

func DeleteWriteLayer(containerName string) {
	writeURL := fmt.Sprintf(devconst.WriteURL, containerName)
	if err := os.RemoveAll(writeURL); err != nil {
		log.Infof("Remove writeLayer dir %s error %v", writeURL, err)
	}
}

func PathExists(path string) (bool, error) {
	_, err := os.Stat(path)
	if err == nil {
		return true, nil
	}
	if os.IsNotExist(err) {
		return false, nil
	}
	return false, err
}
