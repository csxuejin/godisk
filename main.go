package main

import (
	"encoding/json"
	"io/ioutil"
	"os"
	"os/exec"
	"strings"

	"github.com/dolab/logger"
	"github.com/golib/cli"
)

const (
	VERSION              = "1.0.0"
	CommonDiskFromatType = "ext4"
	DiskInfoFileName     = "disk.json"
)

var (
	log    *logger.Logger
	result Result
)

func init() {
	log, _ = logger.New("stdout")
	log.SetColor(true)
	log.SetFlag(3)

	data, err := exec.Command("uname", "-a").Output()
	if err != nil {
		log.Errorf("exec.Command(uname -a): %v\n", err)
		return
	}

	result.System = strings.TrimSuffix(string(data), "\n")
	result.FormatType = CommonDiskFromatType
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
		var tmpResult *Result
		data, err := ioutil.ReadFile(DiskInfoFileName)
		if err != nil {
			log.Errorf("ioutil.ReadFile(%v): %v\n", DiskInfoFileName, err)
			return nil
		}

		if err := json.Unmarshal(data, &tmpResult); err != nil {
			log.Errorf("json.Unmarshal(): %v", err)
			return nil
		}

		for _, v := range tmpResult.Disks {
			if !v.Formated {
				removeAllPartitions(v.Name)
			}
		}
		return nil
	}
}
