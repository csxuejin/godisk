package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os/exec"
	"strconv"
	"strings"
)

type _Helper struct{}

var (
	Helper *_Helper
)

type Result struct {
	System     string     `json:"system"`
	FormatType string     `json:"format_type"`
	Disks      []DiskInfo `json:"disks"`
}

type DiskInfo struct {
	Name     string  `json:"name"`
	Capacity float64 `json:"capacity"`
	Formated bool    `json:"formated"`
}

func parseDisk(infos []string) {
	result.Disks = make([]DiskInfo, 0)
	for _, v := range infos {
		fmt.Println("str is : ", v)
		if strings.HasPrefix(v, "Disk /dev") {
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

			result.Disks = append(result.Disks, DiskInfo{
				Name:     name,
				Capacity: convertToGB(capacity),
			})
		}
	}

	data, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		log.Errorf("json.Marshal(): %v\n", err)
		return
	}

	if err := ioutil.WriteFile("disk.json", data, 0644); err != nil {
		log.Errorf("ioutil.WriteFile(disk.json): %v\n", err)
	}
	return
}

func convertToGB(size float64) float64 {
	return size / 1024 / 1024 / 1024
}

func removeAllPartitions(diskName string) {
	_, err := exec.Command("dd", "if=/dev/zero", "of="+diskName, "count=1", "conv=notrunc").Output()
	if err != nil {
		log.Errorf("remove all partitions in %v: %v\n", diskName, err)
	} else {
		log.Infof("successfully remove all partitions in disk: %v\n", diskName)
	}
}
