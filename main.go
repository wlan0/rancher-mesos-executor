package main

import (
	"os"
	"path/filepath"

	log "github.com/Sirupsen/logrus"
	"github.com/codegangsta/cli"
	"github.com/mesos/mesos-go/executor"
	rancher_mesos "github.com/rancherio/rancher-mesos-executor/executor"
	"github.com/rancherio/rancher-mesos-executor/utils"
)

var (
	GITCOMMIT = "HEAD"
)

func main() {
	app := cli.NewApp()

	app.Name = os.Args[0]
	app.Usage = "Start the rancher-mesos executor"
	app.Version = GITCOMMIT
	app.Author = "Rancher Labs, Inc."
	app.Action = start
	app.Flags = []cli.Flag{
		cli.StringFlag{
			Name:   "work_dir",
			Usage:  "working Directory",
			EnvVar: "RANCHER_MESOS_WORKDIR",
		},
		cli.StringFlag{
			Name:   "bridge_iface",
			Usage:  "the name of the bridge interface",
			EnvVar: "RANCHER_MESOS_BRIDGE",
			Value:  "br0",
		},
		cli.StringFlag{
			Name:   "bridge_cidr",
			Usage:  "CIDR of the bridge interface",
			EnvVar: "RANCHER_MESOS_CIDR",
			Value:  "192.168.11.0/24",
		},
	}

	app.Run(os.Args)
}

func start(c *cli.Context) {
	workDir := c.String("work_dir")
	if workDir == "" {
		log.Fatal("work_dir is not specified")
	}
	err := utils.PerformPreChecksAndPrepareHost(workDir)
	if err != nil {
		log.WithFields(log.Fields{
			"err": err,
		}).Fatal("Environment error, cannot run rancher-mesos-executor")
	}
	log.Info("Setup of env complete. Starting executor driver")
	driver, err := executor.NewMesosExecutorDriver(
		executor.DriverConfig{
			Executor: rancher_mesos.NewRancherExecutor(
				filepath.Join(workDir, "rancheros.iso"),
				c.String("bridge_iface"),
				c.String("bridge_cidr"),
				filepath.Join(workDir, "base-img.img")),
		},
	)
	if err != nil {
		log.WithFields(log.Fields{
			"err": err,
		}).Fatal("Error starting executor")
	}
	_, err = driver.Run()
	if err != nil {
		log.WithFields(log.Fields{
			"err": err,
		}).Fatal("Error starting executor")
	}
}
