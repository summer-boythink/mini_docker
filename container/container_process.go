package container

import (
	"fmt"
	log "github.com/sirupsen/logrus"
	"os"
	"os/exec"
	"strings"
	"syscall"
)

func NewParentProcess(tty bool, volume string) (*exec.Cmd, *os.File) {
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
	}
	cmd.ExtraFiles = []*os.File{readPipe}
	rootURL := "/root/gos_open/mydocker/"
	mntURL := "/root/gos_open/mydocker/mnt/"
	tmpURL := rootURL + "myTmp/"
	writeURL := rootURL + "writeLayer/"
	NewWorkSpace(rootURL, mntURL, tmpURL, writeURL, volume)
	cmd.Dir = mntURL
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
func NewWorkSpace(rootURL string, mntURL string, tmpURL string, writeURL string, volume string) {
	CreateReadOnlyLayer(rootURL)
	CreateWriteLayer(writeURL)
	CreateTmpLayer(tmpURL)
	CreateMountPoint(rootURL, mntURL, tmpURL, writeURL)

	if volume != "" {
		volumeURLs := volumeUrlExtract(volume)
		length := len(volumeURLs)
		if length == 2 && volumeURLs[0] != "" && volumeURLs[1] != "" {
			MountVolume(mntURL, tmpURL, volumeURLs)
			log.Infof("%q", volumeURLs)
		} else {
			log.Infof("Volume parameter input is not correct.")
		}
	}
}

func MountVolume(mntURL string, tmpURL string, volumeURLs []string) {
	// check external url
	parentUrl := volumeURLs[0]
	_, err := os.Stat(parentUrl)
	if os.IsNotExist(err) {
		if err := os.Mkdir(parentUrl, 0777); err != nil {
			log.Infof("Mkdir parent dir %s error. %v", parentUrl, err)
		}
	}

	// check container url
	containerUrl := volumeURLs[1]
	containerVolumeURL := mntURL + containerUrl
	if err := os.Mkdir(containerVolumeURL, 0777); err != nil {
		log.Infof("Mkdir container dir %s error. %v", containerVolumeURL, err)
	}

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

func CreateTmpLayer(tmpURL string) {
	if err := os.Mkdir(tmpURL, 0777); err != nil {
		log.Errorf("Mkdir dir %s error. %v", tmpURL, err)
	}
}

func CreateReadOnlyLayer(rootURL string) {
	busyboxURL := rootURL + "busybox/"
	busyboxTarURL := rootURL + "busybox.tar"
	exist, err := PathExists(busyboxURL)
	if err != nil {
		log.Infof("Fail to judge whether dir %s exists. %v", busyboxURL, err)
	}
	if exist == false {
		if err := os.Mkdir(busyboxURL, 0777); err != nil {
			log.Errorf("Mkdir dir %s error. %v", busyboxURL, err)
		}
		if _, err := exec.Command("tar", "-xvf", busyboxTarURL, "-C", busyboxURL).CombinedOutput(); err != nil {
			log.Errorf("Untar dir %s error %v", busyboxURL, err)
		}
	}
}

func CreateMountPoint(rootURL string, mntURL string, tmpURL string, writeURL string) {
	if err := os.Mkdir(mntURL, 0777); err != nil {
		log.Errorf("Mkdir dir %s error. %v", mntURL, err)
	}
	dirs := fmt.Sprintf("lowerdir=%s,upperdir=%s,workdir=%s", rootURL+"busybox/",
		writeURL, tmpURL)
	cmd := exec.Command("mount", "-t", "overlay", "-o", dirs, "overlay", mntURL)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		log.Errorf("%v", err)
	}
}

func CreateWriteLayer(writeURL string) {
	if err := os.Mkdir(writeURL, 0777); err != nil {
		log.Errorf("Mkdir dir %s error. %v", writeURL, err)
	}
}

// DeleteWorkSpace Delete the AUFS filesystem while container exit
func DeleteWorkSpace(rootURL string, mntURL string, volume string) {
	if volume != "" {
		volumeURLs := volumeUrlExtract(volume)
		length := len(volumeURLs)
		if length == 2 && volumeURLs[0] != "" && volumeURLs[1] != "" {
			DeleteMountPointWithVolume(mntURL, volumeURLs)
		} else {
			DeleteMountPoint(mntURL)
		}
	} else {
		DeleteMountPoint(mntURL)
	}
	DeleteWriteLayer(rootURL)
	DeleteTmpLayer(rootURL)
}

func DeleteMountPointWithVolume(mntURL string, volumeURLs []string) {
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

func DeleteTmpLayer(rootURL string) {
	tmpURL := rootURL + "myTmp/"
	if err := os.RemoveAll(tmpURL); err != nil {
		log.Errorf("Remove dir %s error %v", tmpURL, err)
	}
}

func DeleteMountPoint(mntURL string) {
	cmd := exec.Command("umount", mntURL)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		log.Errorf("%v", err)
	}
	if err := os.RemoveAll(mntURL); err != nil {
		log.Errorf("Remove dir %s error %v", mntURL, err)
	}
}

func DeleteWriteLayer(rootURL string) {
	writeURL := rootURL + "writeLayer/"
	if err := os.RemoveAll(writeURL); err != nil {
		log.Errorf("Remove dir %s error %v", writeURL, err)
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
