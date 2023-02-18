package chain

import (
	"github.com/bnb-chain/greenfield-go-sdk/client/test"
	"github.com/stretchr/testify/assert"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"testing"
)

func TestMultipleGrpcAddresses(t *testing.T) {
	gnfdClients := NewGreenfieldClients(
		[]string{test.TEST_GRPC_ADDR, test.TEST_GRPC_ADDR2, test.TEST_GRPC_ADDR3},
		[]string{test.TEST_RPC_ADDR, test.TEST_RPC_ADDR2, test.TEST_RPC_ADDR3},
		test.TEST_CHAIN_ID,
		WithGrpcDialOption(grpc.WithTransportCredentials(insecure.NewCredentials())))
	client, err := gnfdClients.GetClient()
	assert.NoError(t, err)
	t.Log(client.Height)
}
