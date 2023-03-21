package chain

import (
	"testing"

	"github.com/bnb-chain/greenfield-go-sdk/client/test"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func TestGnfdCompositeClient(t *testing.T) {
	gnfdClients := NewGnfdCompositClients(
		[]string{test.TEST_GRPC_ADDR, test.TEST_GRPC_ADDR2, test.TEST_GRPC_ADDR3},
		[]string{test.TEST_RPC_ADDR, test.TEST_RPC_ADDR2, test.TEST_RPC_ADDR3},
		test.TEST_CHAIN_ID,
		WithGrpcDialOption(grpc.WithTransportCredentials(insecure.NewCredentials())))
	client := gnfdClients.GetClient()
	t.Log(client.Height)
}
