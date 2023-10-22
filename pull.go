package main

import (
	"io"
	"net/http"
	"os"
	"path/filepath"

	log "github.com/sirupsen/logrus"
	"github.com/summer-boythink/mydocker/devconst"

	"github.com/cheggaaa/pb/v3"
)

func pullImage(imageName string) {
	imageStore := devconst.RootURL
	tarFile := filepath.Join(imageStore, imageName+".tar")
	if _, err := os.Stat(imageName); err != nil {
		if os.IsNotExist(err) {
			if err := os.MkdirAll(imageStore, 0755); err == nil {
			} else {
				log.Errorf("Create imageStore failed. %v", err)
			}
		}
	}
	out, err := os.Create(tarFile)
	if err != nil {
		log.Errorf("Failed to create file. %v", err)
		return
	}
	defer out.Close()

	// 一个提供tar image服务的服务器
	resp, err := http.Get("http://110.40.204.35:13229/" + imageName + ".tar")
	if err != nil {
		log.Errorf("Failed to download the image. %v", err)
		return
	}
	defer resp.Body.Close()

	// 创建一个新的进度条
	bar := pb.Full.Start64(resp.ContentLength)
	barReader := bar.NewProxyReader(resp.Body)

	// 将响应体复制到文件，同时更新进度条
	if _, err = io.Copy(out, barReader); err != nil {
		log.Errorf("Failed to save the image. %v", err)
		return
	}

	// 完成进度条
	bar.Finish()

	log.Infof("Successfully downloaded image %s", imageName)
}
