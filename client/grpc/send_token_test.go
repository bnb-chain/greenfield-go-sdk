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

func TestSendTokenSucceed(t *testing.T) {
	km, err := keys.NewPrivateKeyManager("ab463aca3d2965233da3d1d6108aa521274c5ddc2369ff72970a52a451863fbf")
	assert.NoError(t, err)
	gnfdCli := NewGreenfieldClientWithKeyManager(GrpcConn, ChainId, km)

	sendTokenReq := types.SendTokenRequest{
		Token:     "bnb",
		Amount:    10,
		ToAddress: "0x76d244CE05c3De4BbC6fDd7F56379B145709ade9",
	}
	response, err := gnfdCli.SendToken(sendTokenReq, nil)
	assert.NoError(t, err)
	assert.Equal(t, true, response.Ok)
	t.Log(response.TxHash)
}

func TestSendTokenWithTxOptionSucceed(t *testing.T) {
	km, err := keys.NewPrivateKeyManager("ab463aca3d2965233da3d1d6108aa521274c5ddc2369ff72970a52a451863fbf")
	assert.NoError(t, err)
	gnfdCli := NewGreenfieldClientWithKeyManager(GrpcConn, ChainId, km)

	sendTokenReq := types.SendTokenRequest{
		Token:     "bnb",
		Amount:    10,
		ToAddress: "0x76d244CE05c3De4BbC6fDd7F56379B145709ade9",
	}

	txOpt := &types.TxOption{
		GasLimit: 210000,
	}
	response, err := gnfdCli.SendToken(sendTokenReq, txOpt)
	assert.NoError(t, err)
	t.Log(response.TxHash)
	assert.Equal(t, true, response.Ok)

}

func TestSendTokenFailedWithoutInitKeyManager(t *testing.T) {
	gnfdCli := NewGreenfieldClient(GrpcConn, ChainId)
	sendTokenReq := types.SendTokenRequest{
		Token:     "bnb",
		Amount:    10,
		ToAddress: "0x76d244CE05c3De4BbC6fDd7F56379B145709ade9",
	}
	_, err := gnfdCli.SendToken(sendTokenReq, nil)
	assert.Error(t, err)
	assert.Equal(t, types.KeyManagerNotInitError, err)
}
