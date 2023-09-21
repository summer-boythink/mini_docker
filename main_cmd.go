package main

import (
	"fmt"
	"github.com/summer-boythink/mydocker/cgroup/subsystems"
	"github.com/summer-boythink/mydocker/container"

	log "github.com/sirupsen/logrus"
	"github.com/urfave/cli"
)

var runCommand = cli.Command{
	Name:  "run",
	Usage: `Create a new Container`,
	Flags: []cli.Flag{
		cli.BoolFlag{
			Name:  "ti",
			Usage: "enable tty",
		},
		cli.StringFlag{
			Name:  "memory",
			Usage: "memory limit",
		},
		cli.StringFlag{
			Name:  "cpuShare",
			Usage: "cpuShare limit",
		},
		cli.StringFlag{
			Name:  "cpuSet",
			Usage: "cpuSet limit",
		},
		cli.StringFlag{
			Name:  "v",
			Usage: "volume",
		},
		cli.BoolFlag{
			Name:  "d",
			Usage: "detach container",
		},
		cli.StringFlag{
			Name:  "name",
			Usage: "container name",
		},
	},
	Action: func(context *cli.Context) error {
		if len(context.Args()) < 1 {
			return fmt.Errorf("missing container command")
		}

		var cmdArray []string
		for _, arg := range context.Args() {
			cmdArray = append(cmdArray, arg)
		}
		tty := context.Bool("ti")
		detach := context.Bool("d")

		if tty && detach {
			return fmt.Errorf("ti and d can't both provide")
		}

		resConf := &subsystems.ResourceConfig{
			MemoryLimit: context.String("memory"),
			CpuSet:      context.String("cpuSet"),
			CpuShare:    context.String("cpuShare"),
		}
		volume := context.String("v")
		log.Infof("createTty %v", tty)
		containerName := context.String("name")
		Run(tty, cmdArray, resConf, volume, containerName)
		return nil
	},
}

var initCommand = cli.Command{
	Name:  "init",
	Usage: "Init container",
	Action: func(context *cli.Context) error {
		log.Infof("init come")
		err := container.RunContainerInit()
		return err
	},
}

var commitCommand = cli.Command{
	Name:  "commit",
	Usage: "commit a running container to image",
	Action: func(context *cli.Context) error {
		if len(context.Args()) < 1 {
			return fmt.Errorf("missing container name")
		}
		imageName := context.Args().Get(0)
		commitContainer(imageName)
		return nil
	},
}

var listCommand = cli.Command{
	Name:  "ps",
	Usage: "list all the containers",
	Action: func(context *cli.Context) error {
		ListContainers()
		return nil
	},
}

var logCommand = cli.Command{
	Name:  "logs",
	Usage: "print logs of a container",
	Action: func(context *cli.Context) error {
		if len(context.Args()) < 1 {
			return fmt.Errorf("Please input your container name")
		}
		containerName := context.Args().Get(0)
		logContainer(containerName)
		return nil
	},
}
