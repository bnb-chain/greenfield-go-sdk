package main

import (
	"context"
	"log"

	"cosmossdk.io/math"
	"github.com/bnb-chain/greenfield-go-sdk/client"
	"github.com/bnb-chain/greenfield-go-sdk/types"
	gnfdsdktypes "github.com/bnb-chain/greenfield/sdk/types"
)

// it is the example of payment SDKs usage
func main() {
	account, err := types.NewAccountFromPrivateKey("test", privateKey)
	if err != nil {
		log.Fatalf("New account from private key error, %v", err)
	}
	cli, err := client.New(chainId, rpcAddr, client.Option{DefaultAccount: account})
	if err != nil {
		log.Fatalf("unable to new greenfield client, %v", err)
	}
	ctx := context.Background()
	// create a payment account
	txHash, err := cli.CreatePaymentAccount(context.Background(), account.GetAddress().String(), gnfdsdktypes.TxOption{})
	handleErr(err, "CreatePaymentAccount")
	waitForTx, err := cli.WaitForTx(ctx, txHash)
	log.Printf("Wait for tx: %s", waitForTx.String())

	paymentAccounts, err := cli.GetPaymentAccountsByOwner(ctx, account.GetAddress().String())

	// deposit
	paymentAddr := paymentAccounts[len(paymentAccounts)-1].Addr
	depositAmount := math.NewIntFromUint64(100)
	depositTxHash, err := cli.Deposit(ctx, paymentAddr, depositAmount, gnfdsdktypes.TxOption{})
	handleErr(err, "Deposit")
	waitForTx, err = cli.WaitForTx(ctx, txHash)
	log.Printf("Wait for tx: %s", waitForTx.String())
	log.Printf("deposited %s to payment account %s, txHash=%s", depositAmount.String(), paymentAddr, depositTxHash)

	// get stream record
	streamRecord, err := cli.GetStreamRecord(ctx, paymentAddr)
	handleErr(err, "GetStreamRecord")
	log.Printf("stream record has balance %s", streamRecord.StaticBalance)

	// withdraw
	withdrawAmount := math.NewIntFromUint64(50)
	withdrawTxHash, err := cli.Withdraw(ctx, paymentAddr, withdrawAmount, gnfdsdktypes.TxOption{})
	handleErr(err, "Withdraw")
	log.Printf("withdraw tx: %s", withdrawTxHash)

	waitForTx, err = cli.WaitForTx(ctx, txHash)
	log.Printf("Wait for tx: %s", waitForTx.String())

	streamRecordAfterWithdraw, err := cli.GetStreamRecord(ctx, paymentAddr)
	handleErr(err, "GetStreamRecord")
	log.Printf("stream record has balance %s", streamRecordAfterWithdraw.StaticBalance)
}
