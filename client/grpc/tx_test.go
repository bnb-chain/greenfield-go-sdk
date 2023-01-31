package client

import (
	"github.com/bnb-chain/gnfd-go-sdk/keys"
	sdk "github.com/cosmos/cosmos-sdk/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	"testing"
)

func TestBroadcastTx(t *testing.T) {

	km, _ := keys.NewPrivateKeyManager("ab463aca3d2965233da3d1d6108aa521274c5ddc2369ff72970a52a451863fbf")
	gnfdCli := NewGreenlandClientWithKeyManager("localhost:9090", "greenfield_9000-121", km)

	toAddr := "0x76d244CE05c3De4BbC6fDd7F56379B145709ade9"
	to, err := sdk.AccAddressFromHexUnsafe(toAddr)
	if err != nil {
		return
	}

	msg1 := banktypes.NewMsgSend(km.GetAddr(), to, sdk.NewCoins(sdk.NewInt64Coin("bnb", 12)))

	msgs := []sdk.Msg{msg1}

	tx, err := gnfdCli.BroadcastTx(true, msgs...)
	if err != nil {
		println(err.Error())
	}
	println(tx.TxHash)
}
