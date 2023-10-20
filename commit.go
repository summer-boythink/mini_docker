package main

import (
	"fmt"
	log "github.com/sirupsen/logrus"
	"github.com/summer-boythink/mydocker/devconst"
	"os/exec"
)

func commitContainer(containerName, imageName string) {
	mntURL := fmt.Sprintf(devconst.MntURL, containerName)
	mntURL += "/"
	imageTar := devconst.RootURL + imageName + ".tar"

	if _, err := exec.Command("tar", "-czf", imageTar, "-C", mntURL, ".").CombinedOutput(); err != nil {
		log.Errorf("Tar folder %s error %v", mntURL, err)
	}
}
