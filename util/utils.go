package util

import (
	"os"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing/transport/http"
	"github.com/kelseyhightower/envconfig"
)

type Config struct {
	RepoUsername string `envconfig:"REPO_USERNAME"`
	RepoPassword string `envconfig:"REPO_PASSWORD"`
}

func CloneRepo(path, cloneUrl string) error {
	var cfg Config
	err := envconfig.Process("", &cfg)
	if err != nil {
		panic(err)
	}

	_, err = git.PlainClone(path, false, &git.CloneOptions{
		URL:               cloneUrl,
		Progress:          os.Stdout,
		RecurseSubmodules: 1,
		Auth: &http.BasicAuth{
			Username: cfg.RepoUsername,
			Password: cfg.RepoPassword,
		},
	})

	if err != nil {
		return err
	}

	return nil
}

func InterfaceOfSliceToMap(source []interface{}) map[string]interface{} {
	m := map[string]interface{}{}

	for _, e := range source {
		e2 := e.(map[string]interface{})
		m[e2["Key"].(string)] = e2["Value"]
	}

	return m
}

func WriteToFile(path, content string) error {
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
