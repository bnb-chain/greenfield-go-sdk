package client

import (
	"context"
	"errors"
	"fmt"
	"github.com/bnb-chain/gnfd-go-sdk/types"
	clitx "github.com/cosmos/cosmos-sdk/client/tx"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/tx"
	"github.com/cosmos/cosmos-sdk/types/tx/signing"
	xauthsigning "github.com/cosmos/cosmos-sdk/x/auth/signing"
	authtx "github.com/cosmos/cosmos-sdk/x/auth/tx"
)

func (c *GreenfieldClient) BroadcastTx(msgs []sdk.Msg) (string, error) {
	if c.keyManager == nil {
		return "", errors.New("please use the client with key manager to support sending transaction")
	}

	txConfig := authtx.NewTxConfig(types.Cdc(), authtx.DefaultSignModes)
	txBuilder := txConfig.NewTxBuilder()

	err := txBuilder.SetMsgs(msgs...)
	if err != nil {
		return "", err
	}

	txBuilder.SetGasLimit(210000)

	address := c.keyManager.GetAddr().String()
	account, err := c.Account(address)
	if err != nil {
		return "", err
	}
	accountNum := account.GetAccountNumber()
	accountSeq := account.GetSequence()

	sig := signing.SignatureV2{
		PubKey: c.keyManager.GetPrivKey().PubKey(),
		Data: &signing.SingleSignatureData{
			SignMode:  signing.SignMode_SIGN_MODE_EIP_712,
			Signature: nil,
		},
		Sequence: accountSeq,
	}

	err = txBuilder.SetSignatures(sig)
	if err != nil {
		return "", err
	}

	sig = signing.SignatureV2{}

	signerData := xauthsigning.SignerData{
		ChainID:       types.ChainId,
		AccountNumber: accountNum,
		Sequence:      accountSeq,
	}

	sig, err = clitx.SignWithPrivKey(signing.SignMode_SIGN_MODE_EIP_712,
		signerData,
		txBuilder,
		c.keyManager.GetPrivKey(),
		txConfig,
		accountSeq,
	)
	if err != nil {
		return "", err
	}

	err = txBuilder.SetSignatures(sig)
	if err != nil {
		return "", err
	}

	txBytes, err := txConfig.TxEncoder()(txBuilder.GetTx())
	if err != nil {
		return "", err
	}
	// Broadcast transaction
	txRes, err := c.TxClient.BroadcastTx(
		context.Background(),
		&tx.BroadcastTxRequest{
			Mode:    tx.BroadcastMode_BROADCAST_MODE_SYNC,
			TxBytes: txBytes,
		})
	if err != nil {
		return "", err
	}
	if txRes.TxResponse.Code != 0 {
		return "", fmt.Errorf("claim error, code=%d, log=%s", txRes.TxResponse.Code, txRes.TxResponse.RawLog)
	}
	return txRes.TxResponse.TxHash, nil
}
