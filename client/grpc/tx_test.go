package client

import (
	"github.com/bnb-chain/gnfd-go-sdk/keys"
	"github.com/bnb-chain/gnfd-go-sdk/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	"github.com/stretchr/testify/assert"
	"testing"
)

var (
	GrpcConn   = "localhost:9090"
	ChainId    = "greenfield_9000-121"
	ToAddr     = "0x76d244CE05c3De4BbC6fDd7F56379B145709ade9"
	PrivateKey = "ab463aca3d2965233da3d1d6108aa521274c5ddc2369ff72970a52a451863fbf"
)

func TestSendTokenSucceedWithSimulatedGas(t *testing.T) {
	km, err := keys.NewPrivateKeyManager(PrivateKey)
	assert.NoError(t, err)
	gnfdCli := NewGreenfieldClientWithKeyManager(GrpcConn, ChainId, km)
	to, err := sdk.AccAddressFromHexUnsafe(ToAddr)
	assert.NoError(t, err)
	transfer := banktypes.NewMsgSend(km.GetAddr(), to, sdk.NewCoins(sdk.NewInt64Coin("bnb", 12)))
	response, err := gnfdCli.BroadcastTx([]sdk.Msg{transfer}, nil)
	assert.NoError(t, err)
	assert.Equal(t, uint32(0), response.TxResponse.Code)
	t.Log(response.TxResponse.TxHash)
}

func TestSendTokenWithTxOptionSucceed(t *testing.T) {
	km, err := keys.NewPrivateKeyManager(PrivateKey)
	assert.NoError(t, err)
	gnfdCli := NewGreenfieldClientWithKeyManager(GrpcConn, ChainId, km)
	to, err := sdk.AccAddressFromHexUnsafe(ToAddr)
	assert.NoError(t, err)
	transfer := banktypes.NewMsgSend(km.GetAddr(), to, sdk.NewCoins(sdk.NewInt64Coin("bnb", 100)))
	payerAddr, err := sdk.AccAddressFromHexUnsafe(km.GetAddr().String())
	txOpt := &types.TxOption{
		Async:     true,
		GasLimit:  123456,
		Memo:      "test",
		FeeAmount: sdk.Coins{{"bnb", sdk.NewInt(1)}},
		FeePayer:  payerAddr,
	}
	response, err := gnfdCli.BroadcastTx([]sdk.Msg{transfer}, txOpt)
	assert.NoError(t, err)
	assert.Equal(t, uint32(0), response.TxResponse.Code)
}

func TestSimulateTx(t *testing.T) {
	km, err := keys.NewPrivateKeyManager(PrivateKey)
	assert.NoError(t, err)
	gnfdCli := NewGreenfieldClientWithKeyManager(GrpcConn, ChainId, km)
	to, err := sdk.AccAddressFromHexUnsafe(ToAddr)
	assert.NoError(t, err)
	transfer := banktypes.NewMsgSend(km.GetAddr(), to, sdk.NewCoins(sdk.NewInt64Coin("bnb", 100)))
	simulateRes, err := gnfdCli.SimulateTx([]sdk.Msg{transfer}, nil)
	assert.NoError(t, err)
	t.Log(simulateRes.GasInfo.String())
}
