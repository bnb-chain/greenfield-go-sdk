package client

import (
	"github.com/bnb-chain/gnfd-go-sdk/client/testutil"
	"github.com/bnb-chain/gnfd-go-sdk/keys"
	"github.com/bnb-chain/gnfd-go-sdk/types"
	"github.com/stretchr/testify/assert"
	"testing"
)

var (
	GrpcConn = testutil.TEST_GRPC_ADDR
	ChainId  = testutil.TEST_CHAIN_ID
	PrivKey  = testutil.TEST_PRIVATE_KEY
)

func TestSendTokenSucceed(t *testing.T) {
	km, err := keys.NewPrivateKeyManager(PrivKey)
	assert.NoError(t, err)
	gnfdCli := NewGreenfieldClientWithKeyManager(GrpcConn, ChainId, km)

	sendTokenReq := types.SendTokenRequest{
		Token:     "bnb",
		Amount:    10,
		ToAddress: testutil.TEST_ADDR,
	}
	response, err := gnfdCli.SendToken(sendTokenReq, nil)
	assert.NoError(t, err)
	assert.Equal(t, true, response.Ok)
	t.Log(response.TxHash)
}

func TestSendTokenWithTxOptionSucceed(t *testing.T) {
	km, err := keys.NewPrivateKeyManager(PrivKey)
	assert.NoError(t, err)
	gnfdCli := NewGreenfieldClientWithKeyManager(GrpcConn, ChainId, km)

	sendTokenReq := types.SendTokenRequest{
		Token:     "bnb",
		Amount:    10,
		ToAddress: testutil.TEST_ADDR,
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
		ToAddress: testutil.TEST_ADDR,
	}
	_, err := gnfdCli.SendToken(sendTokenReq, nil)
	assert.Error(t, err)
	assert.Equal(t, types.KeyManagerNotInitError, err)
}
