package util

import (
	"bytes"
	"context"
	"testing"

	"github.com/docker/docker/pkg/stdcopy"
)

func TestLogContainer(t *testing.T) {
	h := NewContainerHelper()
	out, err := h.LogContainer(context.TODO(), "hello_world")

	if err != nil {
		t.Error(err)
	}

	var buf bytes.Buffer

	stdcopy.StdCopy(&buf, &buf, out)

	t.Log(buf.String())
}
