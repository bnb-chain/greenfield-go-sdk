package client

import (
	"context"
	"github.com/bnb-chain/gnfd-go-sdk/types"
	"github.com/cosmos/cosmos-sdk/client"
	clitx "github.com/cosmos/cosmos-sdk/client/tx"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/tx"
	"github.com/cosmos/cosmos-sdk/types/tx/signing"
	xauthsigning "github.com/cosmos/cosmos-sdk/x/auth/signing"
	authtx "github.com/cosmos/cosmos-sdk/x/auth/tx"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	"google.golang.org/grpc"
)

type TransactionClient interface {
	BroadcastTx(msgs []sdk.Msg, txOpt *types.TxOption, opts ...grpc.CallOption) (*tx.BroadcastTxResponse, error)
	SimulateTx(msgs []sdk.Msg) ([]byte, error)
	SignTx(msgs []sdk.Msg) ([]byte, error)
}

// BroadcastTx will sign and broadcast a tx with simulated gas(if not provided in txOpt)
func (c *GreenfieldClient) BroadcastTx(msgs []sdk.Msg, txOpt *types.TxOption, opts ...grpc.CallOption) (*tx.BroadcastTxResponse, error) {

	txConfig := authtx.NewTxConfig(types.Cdc(), []signing.SignMode{signing.SignMode_SIGN_MODE_EIP_712})
	txBuilder := txConfig.NewTxBuilder()

	// Build tx and inject it into txBuilder
	if err := c.buildTxWithGasLimit(msgs, txOpt, txConfig, txBuilder); err != nil {
		return nil, err
	}

	// do the actual signing of tx
	txSignedBytes, err := c.signTx(txConfig, txBuilder)
	if err != nil {
		return nil, err
	}

	mode := tx.BroadcastMode_BROADCAST_MODE_SYNC
	if txOpt != nil && txOpt.Async {
		mode = tx.BroadcastMode_BROADCAST_MODE_ASYNC
	}
	txRes, err := c.TxClient.BroadcastTx(
		context.Background(),
		&tx.BroadcastTxRequest{
			Mode:    mode,
			TxBytes: txSignedBytes,
		},
		opts...)
	println(txRes.TxResponse.RawLog)
	if err != nil {
		return nil, err
	}
	return txRes, nil
}

// SimulateTx is for simulating tx to get Gas info
func (c *GreenfieldClient) SimulateTx(msgs []sdk.Msg, txOpt *types.TxOption, opts ...grpc.CallOption) (*tx.SimulateResponse, error) {
	txConfig := authtx.NewTxConfig(types.Cdc(), []signing.SignMode{signing.SignMode_SIGN_MODE_EIP_712})
	txBuilder := txConfig.NewTxBuilder()
	err := c.buildTx(msgs, txOpt, txBuilder)
	if err != nil {
		return nil, err
	}
	txBytes, err := txConfig.TxEncoder()(txBuilder.GetTx())
	if err != nil {
		return nil, err
	}
	simulateResponse, err := c.simulateTx(txBytes, opts...)
	if err != nil {
		return nil, err
	}
	return simulateResponse, nil
}

func (c *GreenfieldClient) simulateTx(txBytes []byte, opts ...grpc.CallOption) (*tx.SimulateResponse, error) {
	simulateResponse, err := c.TxClient.Simulate(
		context.Background(),
		&tx.SimulateRequest{
			TxBytes: txBytes,
		},
		opts...,
	)
	if err != nil {
		return nil, err
	}
	return simulateResponse, nil
}

func (c *GreenfieldClient) SignTx(msgs []sdk.Msg, txOpt *types.TxOption, opts ...grpc.CallOption) ([]byte, error) {
	txConfig := authtx.NewTxConfig(types.Cdc(), []signing.SignMode{signing.SignMode_SIGN_MODE_EIP_712})
	txBuilder := txConfig.NewTxBuilder()
	if err := c.buildTxWithGasLimit(msgs, txOpt, txConfig, txBuilder); err != nil {
		return nil, err
	}
	// sign the tx with signer info
	return c.signTx(txConfig, txBuilder)
}

func (c *GreenfieldClient) signTx(txConfig client.TxConfig, txBuilder client.TxBuilder) ([]byte, error) {
	km, err := c.GetKeyManager()
	if err != nil {
		return nil, err
	}
	account, err := c.getAccount()
	if err != nil {
		return nil, err
	}
	sig := signing.SignatureV2{}
	signerData := xauthsigning.SignerData{
		ChainID:       types.ChainId,
		AccountNumber: account.GetAccountNumber(),
		Sequence:      account.GetSequence(),
	}
	sig, err = clitx.SignWithPrivKey(signing.SignMode_SIGN_MODE_EIP_712,
		signerData,
		txBuilder,
		km.GetPrivKey(),
		txConfig,
		account.GetSequence(),
	)
	if err != nil {
		return nil, err
	}
	err = txBuilder.SetSignatures(sig)
	if err != nil {
		return nil, err
	}
	txSignedBytes, err := txConfig.TxEncoder()(txBuilder.GetTx())
	if err != nil {
		return nil, err
	}
	return txSignedBytes, nil
}

// setSingerInfo gather the signer info by doing "empty signature" hack, and inject it into txBuilder
func (c *GreenfieldClient) setSingerInfo(txBuilder client.TxBuilder) error {
	km, err := c.GetKeyManager()
	if err != nil {
		return err
	}
	account, err := c.getAccount()
	if err != nil {
		return err
	}
	sig := signing.SignatureV2{
		PubKey: km.GetPrivKey().PubKey(),
		Data: &signing.SingleSignatureData{
			SignMode: signing.SignMode_SIGN_MODE_EIP_712,
		},
		Sequence: account.GetSequence(),
	}
	if err := txBuilder.SetSignatures(sig); err != nil {
		return err
	}
	return nil
}

func (c *GreenfieldClient) buildTx(msgs []sdk.Msg, txOpt *types.TxOption, txBuilder client.TxBuilder) error {
	for _, m := range msgs {
		if err := m.ValidateBasic(); err != nil {
			return err
		}
	}
	if err := txBuilder.SetMsgs(msgs...); err != nil {
		return err
	}
	if txOpt != nil {
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
	// inject signer info into txBuilder, it is needed for simulating and signing
	return c.setSingerInfo(txBuilder)
}

func (c *GreenfieldClient) buildTxWithGasLimit(msgs []sdk.Msg, txOpt *types.TxOption, txConfig client.TxConfig, txBuilder client.TxBuilder) error {
	err := c.buildTx(msgs, txOpt, txBuilder)
	if err != nil {
		return err
	}
	if txOpt != nil && txOpt.GasLimit != 0 {
		txBuilder.SetGasLimit(txOpt.GasLimit)
	} else {
		txBytes, err := txConfig.TxEncoder()(txBuilder.GetTx())
		if err != nil {
			return err
		}
		simulateRes, err := c.simulateTx(txBytes)
		if err != nil {
			return err
		}
		txBuilder.SetGasLimit(simulateRes.GasInfo.GetGasUsed())
	}
	return nil
}
func (c *GreenfieldClient) getAccount() (authtypes.AccountI, error) {
	km, err := c.GetKeyManager()
	if err != nil {
		return nil, err
	}
	address := km.GetAddr().String()
	acct, err := c.AuthQueryClient.Account(context.Background(), &authtypes.QueryAccountRequest{Address: address})
	if err != nil {
		return nil, err
	}
	var account authtypes.AccountI
	if err := types.Cdc().InterfaceRegistry().UnpackAny(acct.Account, &account); err != nil {
		return nil, err
	}
	return account, nil
}