package godisk

import (
	"fmt"
	"io/ioutil"
	"os/exec"
	"strconv"
	"strings"

	"github.com/dolab/logger"
)

type DiskClient struct{}

func New() *DiskClient {
	return &DiskClient{}
}

const (
	CommonDiskFromatType = "ext4"
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

type FstabContent struct {
	DeviceName string
	FolderName string
}

func (_ *DiskClient) GetDiskInfo() ([]byte, error) {
	data, err := exec.Command("fdisk", "-l").Output()
	if err != nil {
		log.Errorf("fdisk -l: %v\n", err)
		return nil, err
	}

	infos := strings.Split(string(data), "\n")
	return parseDisk(infos)
}

func (_ *DiskClient) DiskPartitioDiskPartitionn(result *Result) error {
	if result == nil {
		return nil
	}

	cnt := 1
	fstabContents := make([]FstabContent, 0)

	for _, v := range result.Disks {
		if v.NeedFormat {
			log.Infof("now start to format disk %v\n", v.Name)
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
			if err := makeNewPartition(diskName); err != nil && !strings.Contains(err.Error(), "127") && !strings.Contains(err.Error(), "exit status 1") {
				log.Errorf("partition disk(%v) failed : %v\n", diskName, err)
				continue
			}
			log.Infof("partition disk(%v) successfully\n", diskName)

			// step 3: format the new partition
			if err := formatPartition(deviceName); err != nil && !strings.Contains(err.Error(), "exit status 1") {
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

	if len(fstabContents) <= 0 {
		return nil
	}

	newData := ""

	// step 5: modify /etc/fstab file
	data, err := ioutil.ReadFile("/etc/fstab")
	if err != nil {
		log.Errorf("ioutil.ReadFile(/etc/fstab): %v\n", err)
		return err
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
		tmp := fmt.Sprintf("%v	%v	%v	defaults,noatime	0	0\n", dev.DeviceName, dev.FolderName, result.FormatType)
		newData = newData + tmp
	}

	if err := ioutil.WriteFile("/etc/fstab", []byte(newData), 0644); err != nil {
		log.Errorf("ioutil.WriteFile(/etc/fstab): %v", 0644)
		return err
	}
	log.Infof("write /etc/fstab successfully!\n")

	// step 6: exec mount command
	if err := mountPartitions(); err != nil {
		log.Errorf("mountPartitions(): %v\n", err)
		return err
	}
	log.Infof("mount partitions successfully!\n")

	return nil
}
