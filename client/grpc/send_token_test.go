package client

import (
	"github.com/bnb-chain/gnfd-go-sdk/keys"
	"github.com/bnb-chain/gnfd-go-sdk/types"
	"github.com/stretchr/testify/assert"
	"testing"
)

var (
	GrpcConn = "localhost:9090"
	ChainId  = "greenfield_9000-121"
)

func TestSendToken(t *testing.T) {
	km, err := keys.NewPrivateKeyManager("ab463aca3d2965233da3d1d6108aa521274c5ddc2369ff72970a52a451863fbf")
	assert.NoError(t, err)
	gnfdCli := NewGreenlandClientWithKeyManager(GrpcConn, ChainId, km)

	sendTokenReq := types.SendTokenRequest{
		"bnb",
		10,
		"0x76d244CE05c3De4BbC6fDd7F56379B145709ade9",
	}
	response, err := gnfdCli.SendToken(sendTokenReq, false)
	assert.NoError(t, err)
	assert.Equal(t, true, response.Ok)
	t.Log(response.TxHash)
}

func TestSendTokenFailedWithoutInitKeyManager(t *testing.T) {
	gnfdCli := NewGreenlandClient(GrpcConn, ChainId)
	sendTokenReq := types.SendTokenRequest{
		Token:     "bnb",
		Amount:    10,
		ToAddress: "0x76d244CE05c3De4BbC6fDd7F56379B145709ade9",
	}
	_, err := gnfdCli.SendToken(sendTokenReq, false)
	assert.Error(t, err)
	assert.Equal(t, types.KeyManagerNotInitError, err)
}
