package util

import (
	"deploybot-service-agent/model"
	"fmt"
	"os"
	"strings"
	"syscall"
)

func InterfaceOfSliceToMap(source []interface{}) map[string]interface{} {
	m := map[string]interface{}{}

	for _, e := range source {
		e2 := e.(map[string]interface{})
		m[e2["Key"].(string)] = e2["Value"]
	}

	return m
}

func WriteToFile(path string, content string) error {
	dir := path[:strings.LastIndex(path, "/")]
	err := CreateDirsIfNotExist(dir)

	if err != nil {
		return err
	}

	file, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE, 0644)
	if err != nil {
		return err
	}
	defer file.Close()

	_, err = file.WriteString(content)
	if err != nil {
		return err
	}

	return nil
}

func CreateDirsIfNotExist(dirPath string) error {
	if _, err := os.Stat(dirPath); os.IsNotExist(err) {
		err := os.MkdirAll(dirPath, os.ModePerm)
		if err != nil {
			return err
		}
	}
	return nil
}

func GetDiskInfo(mountPoint string) (*model.DiskInfo, error) {
	var stat syscall.Statfs_t
	err := syscall.Statfs(mountPoint, &stat)
	if err != nil {
		fmt.Fprintf(os.Stderr, "statfs failed: %v\n", err)
		return nil, err
	}

	totalSize := stat.Blocks * uint64(stat.Bsize)
	availableSpace := stat.Bavail * uint64(stat.Bsize)

	return &model.DiskInfo{TotalSize: totalSize, AvailSize: availableSpace, Path: mountPoint}, nil
}
