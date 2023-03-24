package main

import (
	"io/ioutil"
	"net/http"
	"testing"
)

func TestServiceLogHandler(t *testing.T) {

	res, err := http.Get("http://localhost:8083/serviceLogs?name=hello_world")
	if err != nil {
		t.Error(err)
	}

	defer res.Body.Close()
	out, err := ioutil.ReadAll(res.Body)
	if err != nil {
		t.Log(err)
	}

	t.Log(out)
}
