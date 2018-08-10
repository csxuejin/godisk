package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"strconv"
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

type FstabContent struct {
	DeviceName string
	FolderName string
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

		cnt := 1
		fstabContents := make([]FstabContent, 0)

		for _, v := range tmpResult.Disks {
			if !v.Formated {
				diskName := v.Name
				deviceName := diskName + "1" // 一块盘只有一个分区
				folderName := "/disk" + strconv.Itoa(cnt)

				// step 1 : remove all partitions
				if err := removeAllPartitions(diskName); err != nil {
					log.Errorf("removeAllPartitions(%v): %v\n", diskName, err)
					continue
				}
				log.Infof("successfully remove all partitions in disk: %v \n", diskName)

				// step 2 : make new partition，the default partition name is 'diskName+1', eg : /dev/vdb1
				if err := makeNewPartition(diskName); err != nil && !strings.Contains(err.Error(), "127") {
					log.Errorf("partition disk(%v) failed : %v\n", diskName, err)
					continue
				}
				log.Infof("partition disk(%v) successfully\n", diskName)

				// step 3: format the new partition
				if err := formatPartition(deviceName); err != nil {
					log.Errorf("mkfs.ext4 %v failed: %v \n", deviceName, err)
					continue
				}
				log.Infof("mkfs.ext4 %v successfully! \n", deviceName)

				// step 4: create folder to mount
				if err := createFolder(folderName); err != nil {
					log.Infof("createFolder %v failed: %v\n", folderName, err)
					continue
				}
				log.Infof("createFolder %v successfully! \n", folderName)

				fstabContents = append(fstabContents, FstabContent{
					DeviceName: deviceName,
					FolderName: folderName,
				})

				cnt++
			}
		}

		newData := ""

		// step 5: modify /etc/fstab file
		data, err = ioutil.ReadFile("/etc/fstab")
		if err != nil {
			log.Errorf("ioutil.ReadFile(/etc/fstab): %v\n", err)
		}

		for _, v := range strings.Split(string(data), "\n") {
			exist := false
			if !strings.HasPrefix(v, "#") {
				for _, dev := range fstabContents {
					if strings.HasPrefix(v, dev.DeviceName) {
						exist = true
						break
					}
				}
			}

			if !exist {
				newData = newData + v + "\n"
			}
		}

		for _, dev := range fstabContents {
			tmp := fmt.Sprintf("%v	%v	ext4	defaults,noatime	0	0\n", dev.DeviceName, dev.FolderName)
			newData = newData + tmp
		}

		if err := ioutil.WriteFile("/etc/fstab", []byte(newData), 0644); err != nil {
			log.Errorf("ioutil.WriteFile(/etc/fstab): %v", 0644)
			return nil
		}
		log.Infof("write /etc/fstab successfully!\n")

		// step 6: exec mount command
		if err := mountPartitions(); err != nil {
			log.Errorf("mountPartitions(): %v\n", err)
			return nil
		}
		log.Infof("mount partitions successfully!\n")

		return nil
	}
}
