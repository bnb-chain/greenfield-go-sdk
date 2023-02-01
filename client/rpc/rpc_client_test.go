package client

import (
	"context"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestGetABCIInfo(t *testing.T) {
	client := NewRPCClient("http://0.0.0.0:26750")
	_, err := client.RpcClient.ABCIInfo(context.Background())
	assert.NoError(t, err)
}

func TestGetStatus(t *testing.T) {
	client := NewRPCClient("http://0.0.0.0:26750")
	_, err := client.RpcClient.Status(context.Background())
	assert.NoError(t, err)
}
