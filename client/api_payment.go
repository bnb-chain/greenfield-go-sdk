package client

import (
	"context"
	"cosmossdk.io/math"
	gnfdSdkTypes "github.com/bnb-chain/greenfield/sdk/types"
	sdk "github.com/cosmos/cosmos-sdk/types"

	paymentTypes "github.com/bnb-chain/greenfield/x/payment/types"
)

type Payment interface {
	GetStreamRecord(ctx context.Context, streamAddress string) (*paymentTypes.StreamRecord, error)

	Deposit(ctx context.Context, toAddress string, amount math.Int, txOption *gnfdSdkTypes.TxOption) (string, error)
	Withdraw(ctx context.Context, fromAddress string, amount math.Int, txOption *gnfdSdkTypes.TxOption) (string, error)
	DisableRefund(ctx context.Context, paymentAddress string, txOption *gnfdSdkTypes.TxOption) (string, error)
}

// GetStreamRecord retrieves stream record information for a given stream address.
func (c *client) GetStreamRecord(ctx context.Context, streamAddress string) (*paymentTypes.StreamRecord, error) {
	accAddress, err := sdk.AccAddressFromHexUnsafe(streamAddress)
	if err != nil {
		return nil, err
	}
	pa, err := c.chainClient.StreamRecord(ctx, &paymentTypes.QueryGetStreamRecordRequest{Account: accAddress.String()})
	if err != nil {
		return nil, err
	}
	return &pa.StreamRecord, nil
}

// Deposit deposits BNB to a stream account.
func (c *client) Deposit(ctx context.Context, toAddress string, amount math.Int, txOption gnfdSdkTypes.TxOption) (string, error) {
	accAddress, err := sdk.AccAddressFromHexUnsafe(toAddress)
	if err != nil {
		return "", err
	}
	msgDeposit := &paymentTypes.MsgDeposit{
		Creator: c.MustGetDefaultAccount().GetAddress().String(),
		To:      accAddress.String(),
		Amount:  amount,
	}
	tx, err := c.chainClient.BroadcastTx(ctx, []sdk.Msg{msgDeposit}, &txOption)
	if err != nil {
		return "", err
	}
	return tx.TxResponse.TxHash, nil
}

// Withdraw withdraws BNB from a stream account.
func (c *client) Withdraw(ctx context.Context, fromAddress string, amount math.Int, txOption gnfdSdkTypes.TxOption) (string, error) {
	accAddress, err := sdk.AccAddressFromHexUnsafe(fromAddress)
	if err != nil {
		return "", err
	}
	msgWithdraw := &paymentTypes.MsgWithdraw{
		Creator: c.MustGetDefaultAccount().GetAddress().String(),
		From:    accAddress.String(),
		Amount:  amount,
	}
	tx, err := c.chainClient.BroadcastTx(ctx, []sdk.Msg{msgWithdraw}, &txOption)
	if err != nil {
		return "", err
	}
	return tx.TxResponse.TxHash, nil
}

// DisableRefund disables refund for a stream account.
func (c *client) DisableRefund(ctx context.Context, paymentAddress string, txOption gnfdSdkTypes.TxOption) (string, error) {
	accAddress, err := sdk.AccAddressFromHexUnsafe(paymentAddress)
	if err != nil {
		return "", err
	}
	msgDisableRefund := &paymentTypes.MsgDisableRefund{
		Owner: c.MustGetDefaultAccount().GetAddress().String(),
		Addr:  accAddress.String(),
	}
	tx, err := c.chainClient.BroadcastTx(ctx, []sdk.Msg{msgDisableRefund}, &txOption)
	if err != nil {
		return "", err
	}
	return tx.TxResponse.TxHash, nil
}
