package main

import (
	"os"
	"os/exec"
	"strings"

	"github.com/dolab/logger"
	"github.com/golib/cli"
)

const (
	VERSION = "1.0.0"
)

var (
	log *logger.Logger
)

func init() {
	log, _ = logger.New("stdout")
	log.SetColor(true)
	log.SetFlag(3)
}

func main() {
	app := cli.NewApp()
	app.Name = "godisk"
	app.Version = VERSION
	app.Authors = []cli.Author{
		{
			Name:  "Xue Jin",
			Email: "csxuejin@gmail.com",
		},
	}

	app.Commands = []cli.Command{
		{
			Name:   "diskinfo",
			Usage:  "get disk info",
			Action: getDiskInfo(log),
		},
		{
			Name:   "partition",
			Usage:  "disk partition",
			Action: diskPartition(log),
		},
	}

	err := app.Run(os.Args)
	if err != nil {
		log.Fatal(err)
	} else {
		log.Info("ok!")
	}
}

func getDiskInfo(log *logger.Logger) cli.ActionFunc {
	return func(ctx *cli.Context) (err error) {
		data, err := exec.Command("fdisk", "-l").Output()
		if err != nil {
			log.Errorf("cmd.CombinedOutput(): %v\n", err)
		} else {
			log.Infof("data is : %#v\n", string(data))
			infos := strings.Split(string(data), "\n")
			parseDisk(infos)
		}

		return nil
	}
}

func diskPartition(log *logger.Logger) cli.ActionFunc {
	return func(ctx *cli.Context) error {

		return nil
	}
}
