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
			Name: "memory",
			Usage: "memory limit",
		},
		cli.StringFlag{
			Name: "cpuShare",
			Usage: "cpuShare limit",
		},
		cli.StringFlag{
			Name: "cpuSet",
			Usage: "cpuSet limit",
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
		resConf := &subsystems.ResourceConfig{
			MemoryLimit: context.String("memory"),
			CpuSet: context.String("cpuSet"),
			CpuShare:context.String("cpuShare"),
		}
		Run(tty, cmdArray,resConf)
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
