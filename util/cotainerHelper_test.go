package util

import (
	"context"
	"io"
	"testing"
)

func TestLogContainer(t *testing.T) {
	h := NewContainerHelper()
	out, err := h.LogContainer(context.TODO(), "hello_world")

	if err != nil {
		t.Error(err)
	}

	bs, _ := io.ReadAll(out)

	t.Log(bs)
}
