package godisk

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"sort"
	"strconv"
	"strings"
)

type _Helper struct{}

var (
	Helper *_Helper
)

type DiskInfo struct {
	Name       string  `json:"name"`
	Capacity   float64 `json:"capacity"`
	DiskType   int     `json:"disk_type"` // 0: SSD, 1: SATA
	Formated   bool    `json:"formated"`
	NeedFormat bool    `json:"need_format"`
}

type DiskInfos []*DiskInfo

func (s DiskInfos) Len() int {
	return len(s)
}
func (s DiskInfos) Swap(i, j int) {
	s[i], s[j] = s[j], s[i]
}
func (s DiskInfos) Less(i, j int) bool {
	return s[i].DiskType < s[j].DiskType
}

type Result struct {
	System     string    `json:"system"`
	FormatType string    `json:"format_type"`
	Disks      DiskInfos `json:"disks"`
}

func getDiskType() map[string]int {
	data, err := exec.Command("bash", "-c", `lsblk -d -o name,rota`).Output()
	if err != nil {
		log.Errorf("exec.Command(lsblk -d -o name,rota): %v\n", err)
		return nil
	}

	res := make(map[string]int)
	disks := strings.Split(string(data), "\n")
	for i, v := range disks {
		if i == 0 || v == "" {
			continue
		}

		strs := strings.Split(v, " ")
		num, _ := strconv.Atoi(strs[len(strs)-1])
		res[strs[0]] = num
	}

	return res
}

func parseDisk(infos []string) (data []byte, err error) {
	diskTypeMap := getDiskType()

	result.Disks = make([]*DiskInfo, 0)
	devices := make([]string, 0)
	bootDevice := ""
	for _, v := range infos {
		if strings.HasPrefix(v, "Disk /dev") || strings.HasPrefix(v, "磁盘 /dev") {
			nameBegin := strings.Index(v, "/dev")
			nameEnd := strings.Index(v, ":")
			name := v[nameBegin:nameEnd]

			capacityBegin := strings.Index(v, ", ")
			capacityEnd := strings.Index(v, " bytes")
			capacityStr := v[capacityBegin+2 : capacityEnd]
			capacity, err := strconv.ParseFloat(capacityStr, 64)
			if err != nil {
				log.Errorf("strconv.ParseFloat(%v, 64)\n", capacityStr, 64)
				continue
			}
			result.Disks = append(result.Disks, &DiskInfo{
				Name:       name,
				Capacity:   convertToGB(capacity),
				DiskType:   diskTypeMap[name],
				NeedFormat: true,
			})

		} else if strings.HasPrefix(v, "/dev") {
			deviceName := strings.Split(v, " ")[0]
			devices = append(devices, deviceName)
			if strings.Contains(v, "*") {
				bootDevice = deviceName
			}
		}
	}

	// remove Boot disk from the result
	if bootDevice != "" {
		for i, v := range result.Disks {
			if strings.HasPrefix(bootDevice, v.Name) {
				fmt.Println("boot disk is ", v.Name)
				result.Disks = append(result.Disks[:i], result.Disks[i+1:]...)
				break
			}
		}
	}

	for _, device := range devices {
		for _, disk := range result.Disks {
			if strings.HasPrefix(device, disk.Name) && !disk.Formated {
				disk.Formated = true
			}
		}
	}

	sort.Sort(result.Disks)

	data, err = json.MarshalIndent(result, "", "  ")
	if err != nil {
		log.Errorf("json.Marshal(): %v\n", err)
		return
	}

	return
}

func convertToGB(size float64) float64 {
	return size / 1024 / 1024 / 1024
}

func getMountInfo() ([]string, error) {
	data, err := exec.Command("df").Output()
	if err != nil {
		log.Errorf("exec df failed: %v\n", err)
		return nil, err
	}

	infos := strings.Split(string(data), "\n")
	return infos, nil
}

func removeAllPartitions(diskName string) error {
	mountInfo, err := getMountInfo()
	if err == nil {
		for _, v := range mountInfo {
			args := strings.Split(v, " ")
			if len(args) > 1 && strings.HasPrefix(args[0], diskName) {
				exec.Command("umount", args[0]).Run()
			}
		}
	}

	return exec.Command("dd", "if=/dev/zero", "of="+diskName, "count=1", "conv=notrunc").Run()
}

func makeNewPartition(diskName string) error {
	return exec.Command("bash", "-c", `parted -a opt `+diskName+` --script mklabel gpt --script mkpart primary 2048s 100%`).Run()
}

func formatPartition(deviceName string) error {
	return exec.Command("bash", "-c", `echo -e "\ny" | mkfs.ext4 `+deviceName).Run()
}

func createFolder(folerName string) error {
	return os.MkdirAll(folerName, 0700)
}

func mountPartitions() error {
	return exec.Command("mount", "-a").Run()
}
