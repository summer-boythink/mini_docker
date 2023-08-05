package main

import (
	"fmt"
	log "github.com/sirupsen/logrus"
	"github.com/summer-boythink/mydocker/devconst"
	"os/exec"
)

func commitContainer(imageName string) {
	imageTar := devconst.RootURL + imageName + ".tar"
	fmt.Printf("%s\n", imageTar)
	if _, err := exec.Command("tar", "-czf", imageTar, "-C", devconst.MntURL, ".").CombinedOutput(); err != nil {
		log.Errorf("Tar folder %s error %v", devconst.MntURL, err)
	}
}
