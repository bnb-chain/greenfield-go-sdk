package client

import (
	"context"
	"encoding/xml"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"

	"cosmossdk.io/math"
	gnfdSdkTypes "github.com/bnb-chain/greenfield/sdk/types"
	paymentTypes "github.com/bnb-chain/greenfield/x/payment/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/rs/zerolog/log"

	"github.com/bnb-chain/greenfield-go-sdk/pkg/utils"
	"github.com/bnb-chain/greenfield-go-sdk/types"
)

type IPaymentClient interface {
	GetStreamRecord(ctx context.Context, streamAddress string) (*paymentTypes.StreamRecord, error)

	Deposit(ctx context.Context, toAddress string, amount math.Int, txOption gnfdSdkTypes.TxOption) (string, error)
	Withdraw(ctx context.Context, fromAddress string, amount math.Int, txOption gnfdSdkTypes.TxOption) (string, error)
	DisableRefund(ctx context.Context, paymentAddress string, txOption gnfdSdkTypes.TxOption) (string, error)
	ListUserPaymentAccounts(ctx context.Context, opts types.ListUserPaymentAccountsOptions) (types.ListUserPaymentAccountsResult, error)
}

// GetStreamRecord retrieves stream record information for a given stream address.
func (c *Client) GetStreamRecord(ctx context.Context, streamAddress string) (*paymentTypes.StreamRecord, error) {
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
func (c *Client) Deposit(ctx context.Context, toAddress string, amount math.Int, txOption gnfdSdkTypes.TxOption) (string, error) {
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
func (c *Client) Withdraw(ctx context.Context, fromAddress string, amount math.Int, txOption gnfdSdkTypes.TxOption) (string, error) {
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
func (c *Client) DisableRefund(ctx context.Context, paymentAddress string, txOption gnfdSdkTypes.TxOption) (string, error) {
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

// ListUserPaymentAccounts list payment info by user address
func (c *Client) ListUserPaymentAccounts(ctx context.Context, opts types.ListUserPaymentAccountsOptions) (types.ListUserPaymentAccountsResult, error) {
	params := url.Values{}
	params.Set("user-payments", "")

	account := opts.Account
	if account == "" {
		acc, err := c.GetDefaultAccount()
		if err != nil {
			log.Error().Msg(fmt.Sprintf("failed to get default account: %s", err.Error()))
			return types.ListUserPaymentAccountsResult{}, err
		}
		account = acc.GetAddress().String()
	}

	reqMeta := requestMeta{
		urlValues:     params,
		contentSHA256: types.EmptyStringSHA256,
		userAddress:   account,
	}

	sendOpt := sendOptions{
		method:           http.MethodGet,
		disableCloseBody: true,
	}

	endpoint, err := c.getEndpointByOpt(&types.EndPointOptions{
		Endpoint:  opts.Endpoint,
		SPAddress: opts.SPAddress,
	})
	if err != nil {
		log.Error().Msg(fmt.Sprintf("get endpoint by option failed %s", err.Error()))
		return types.ListUserPaymentAccountsResult{}, err
	}

	resp, err := c.sendReq(ctx, reqMeta, &sendOpt, endpoint)
	if err != nil {
		return types.ListUserPaymentAccountsResult{}, errors.New("send request error" + err.Error())
	}
	defer utils.CloseResponse(resp)

	paymentAccounts := types.ListUserPaymentAccountsResult{}
	// unmarshal the json content from response body
	buf := new(strings.Builder)
	_, err = io.Copy(buf, resp.Body)
	if err != nil {
		return types.ListUserPaymentAccountsResult{}, errors.New("unmarshal response error" + err.Error())
	}

	bufStr := buf.String()
	err = xml.Unmarshal([]byte(bufStr), &paymentAccounts)

	if err != nil {
		return types.ListUserPaymentAccountsResult{}, err
	}

	return paymentAccounts, nil
}
