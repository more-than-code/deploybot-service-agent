package util

import (
	"bytes"
	"context"
	types "deploybot-service-launcher/deploybot-types"
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

func TestStartContainer(t *testing.T) {
	h := NewContainerHelper()

	h.StartContainer(&types.DeployConfig{
		ImageName: "binartist/messaging",
		ImageTag:  "latest",
		Ports:     map[string]string{"8002": "8002"},
		Env: []string{
			"SERVER_PORT=:8002",
			"SMS_TEMPLATE_CODE=SMS_181853823",
			"SMS_GLOBE_ACCESS_KEY_ID=LTAI5tM4BQSASzQVpGFStfQK",
			"SMS_GLOBE_ACCESS_KEY_SECRET=1CBanvughH7tNkceUxWNjra6zO2Urd",
			"POSTMARK_API_KEY=6fc1d464-ada4-4ae8-a7da-315ef8d83de6",
			"POSTMARK_MAIL_SENDER=cn@woofaa.com",
			"REDIS_URI=redis:6379",
			"EMAIL_DOMAINS=''",
			"BYPASS_CODE=''",
		},
		NetworkId:   "b602edbe428ec9e5549e1fefa0f9d56fbd5955ec3b10d46c8d2097057ecdc362",
		NetworkName: "geoy-network",
		ServiceName: "messaging",
	})
}
