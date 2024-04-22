package util

import (
	"archive/tar"
	"bytes"
	"deploybot-service-agent/model"
	"fmt"
	"io"
	"io/fs"
	"log"
	"os"
	"path/filepath"
	"syscall"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing/transport/http"

	"github.com/go-git/go-git/v5/plumbing"
	"gopkg.in/mgo.v2/bson"
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
	// Ensure the directory exists
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}

	// Open the file with truncation flag to ensure it's overwritten
	file, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0644)
	if err != nil {
		return err
	}
	defer file.Close()

	if _, err := file.WriteString(content); err != nil {
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

func TarFiles(dir string) (io.Reader, error) {
	var buf bytes.Buffer
	tw := tar.NewWriter(&buf)

	fileSystem := os.DirFS(dir)

	fs.WalkDir(fileSystem, ".", func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			log.Fatal(err)
		}

		if d.IsDir() {
			return nil
		}

		bytes, err := os.ReadFile(dir + path)
		if err != nil {
			log.Fatal(err)
		}

		hdr := &tar.Header{
			Name: path,
			Mode: 0600,
			Size: int64(len(bytes)),
		}
		if err := tw.WriteHeader(hdr); err != nil {
			log.Fatal(err)
		}
		if _, err := tw.Write(bytes); err != nil {
			log.Fatal(err)
		}

		return nil
	})

	if err := tw.Close(); err != nil {
		log.Fatal(err)
	}

	return &buf, nil
}

type GitCredentials struct {
	Username string
	Password string
}

func CloneRepo(path, cloneUrl, branch string, cred GitCredentials) error {
	_, err := git.PlainClone(path, false, &git.CloneOptions{
		URL:               cloneUrl,
		ReferenceName:     plumbing.NewBranchReferenceName(branch),
		Progress:          os.Stdout,
		RecurseSubmodules: 1,
		Auth: &http.BasicAuth{
			Username: cred.Username,
			Password: cred.Password,
		},
	})

	if err != nil {
		return err
	}

	return nil
}

func StructToBsonDoc(source interface{}) bson.M {
	bytes, err := bson.Marshal(source)

	if err != nil {
		return nil
	}

	doc := bson.M{}
	err = bson.Unmarshal(bytes, &doc)

	if err != nil {
		return nil
	}

	return doc
}
