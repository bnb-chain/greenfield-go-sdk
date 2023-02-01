package client

import (
	"context"
	"github.com/bnb-chain/gnfd-go-sdk/types"
	clitx "github.com/cosmos/cosmos-sdk/client/tx"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/tx"
	"github.com/cosmos/cosmos-sdk/types/tx/signing"
	xauthsigning "github.com/cosmos/cosmos-sdk/x/auth/signing"
	authtx "github.com/cosmos/cosmos-sdk/x/auth/tx"
	"google.golang.org/grpc"
)

type TransactionClient interface {
	BroadcastTx(msgs []sdk.Msg, txOpt *types.TxOption, opts ...grpc.CallOption) (*types.TxBroadcastResponse, error)
	SendToken(req types.SendTokenRequest, txOpt *types.TxOption, opts ...grpc.CallOption) (*types.TxBroadcastResponse, error)
}

func (c *GreenfieldClient) BroadcastTx(msgs []sdk.Msg, txOpt *types.TxOption, opts ...grpc.CallOption) (*types.TxBroadcastResponse, error) {

	txConfig := authtx.NewTxConfig(types.Cdc(), authtx.DefaultSignModes)
	txBuilder := txConfig.NewTxBuilder()

	err := txBuilder.SetMsgs(msgs...)
	if err != nil {
		return nil, err
	}

	txBuilder.SetGasLimit(types.DefaultGasLimit)
	mode := tx.BroadcastMode_BROADCAST_MODE_SYNC

	if txOpt != nil {
		if txOpt.Async {
			mode = tx.BroadcastMode_BROADCAST_MODE_ASYNC
		}
		if txOpt.GasLimit != 0 {
			txBuilder.SetGasLimit(txOpt.GasLimit)
		}
		if txOpt.Memo != "" {
			txBuilder.SetMemo(txOpt.Memo)
		}
		if !txOpt.FeeAmount.IsZero() {
			txBuilder.SetFeeAmount(txOpt.FeeAmount)
		}
		if !txOpt.FeePayer.Empty() {
			txBuilder.SetFeePayer(txOpt.FeePayer)
		}
	}

	km, err := c.GetKeyManager()
	if err != nil {
		return nil, err
	}

	address := km.GetAddr().String()
	account, err := c.Account(address)
	if err != nil {
		return nil, err
	}
	accountNum := account.GetAccountNumber()
	accountSeq := account.GetSequence()

	sig := signing.SignatureV2{
		PubKey: km.GetPrivKey().PubKey(),
		Data: &signing.SingleSignatureData{
			SignMode:  signing.SignMode_SIGN_MODE_EIP_712,
			Signature: nil,
		},
		Sequence: accountSeq,
	}

	err = txBuilder.SetSignatures(sig)
	if err != nil {
		return nil, err
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
		km.GetPrivKey(),
		txConfig,
		accountSeq,
	)
	if err != nil {
		return nil, err
	}

	err = txBuilder.SetSignatures(sig)
	if err != nil {
		return nil, err
	}

	txBytes, err := txConfig.TxEncoder()(txBuilder.GetTx())
	if err != nil {
		return nil, err
	}

	txRes, err := c.TxClient.BroadcastTx(
		context.Background(),
		&tx.BroadcastTxRequest{
			Mode:    mode,
			TxBytes: txBytes,
		},
		opts...)

	if err != nil {
		return nil, err
	}
	txResponse := txRes.TxResponse
	return &types.TxBroadcastResponse{
		Ok:     txResponse.Code == 0,
		Log:    txRes.TxResponse.RawLog,
		TxHash: txResponse.TxHash,
		Code:   txResponse.Code,
		Data:   txResponse.Data,
	}, nil
}
