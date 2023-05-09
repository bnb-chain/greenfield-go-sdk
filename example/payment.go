package main

import (
	"context"
	"cosmossdk.io/math"
	"github.com/bnb-chain/greenfield-go-sdk/client"
	gnfdsdktypes "github.com/bnb-chain/greenfield/sdk/types"
	"log"
)

func payment(cli client.Client) {
	ctx := context.Background()
	account, _ := cli.GetDefaultAccount()
	// create a payment account
	txHash, err := cli.CreatePaymentAccount(context.Background(), account.GetAddress().String(), gnfdsdktypes.TxOption{})
	HandleErr(err, "CreatePaymentAccount")
	waitForTx, err := cli.WaitForTx(ctx, txHash)
	log.Printf("Wait for tx: %s", waitForTx.String())

	paymentAccounts, err := cli.GetPaymentAccountsByOwner(ctx, account.GetAddress().String())

	// deposit
	paymentAddr := paymentAccounts[len(paymentAccounts)-1].Addr
	depositAmount := math.NewIntFromUint64(100)
	depositTxHash, err := cli.Deposit(ctx, paymentAddr, depositAmount, gnfdsdktypes.TxOption{})
	HandleErr(err, "Deposit")
	waitForTx, err = cli.WaitForTx(ctx, txHash)
	log.Printf("Wait for tx: %s", waitForTx.String())
	log.Printf("deposited %s to payment account %s, txHash=%s", depositAmount.String(), paymentAddr, depositTxHash)

	// get stream record
	streamRecord, err := cli.GetStreamRecord(ctx, paymentAddr)
	HandleErr(err, "GetStreamRecord")
	log.Printf("stream record has balance %s", streamRecord.StaticBalance)

	// withdraw
	withdrawAmount := math.NewIntFromUint64(50)
	withdrawTxHash, err := cli.Withdraw(ctx, paymentAddr, withdrawAmount, gnfdsdktypes.TxOption{})
	HandleErr(err, "Withdraw")
	log.Printf("withdraw tx: %s", withdrawTxHash)

	waitForTx, err = cli.WaitForTx(ctx, txHash)
	log.Printf("Wait for tx: %s", waitForTx.String())

	streamRecordAfterWithdraw, err := cli.GetStreamRecord(ctx, paymentAddr)
	HandleErr(err, "GetStreamRecord")
	log.Printf("stream record has balance %s", streamRecordAfterWithdraw.StaticBalance)
}
