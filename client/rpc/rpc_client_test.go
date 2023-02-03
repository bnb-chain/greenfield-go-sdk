package client

import (
	"context"
	"github.com/bnb-chain/gnfd-go-sdk/client/testutil"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestGetABCIInfo(t *testing.T) {
	client := NewRPCClient(testutil.TEST_RPC_ADDR)
	_, err := client.RpcClient.ABCIInfo(context.Background())
	assert.NoError(t, err)
}

func TestGetStatus(t *testing.T) {
	client := NewRPCClient(testutil.TEST_RPC_ADDR)
	_, err := client.RpcClient.Status(context.Background())
	assert.NoError(t, err)
}
