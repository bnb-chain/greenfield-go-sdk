package main

import (
	"context"
	"log"

	"github.com/bnb-chain/greenfield-go-sdk/bsc"
	"github.com/bnb-chain/greenfield-go-sdk/bsctypes"
	"github.com/bnb-chain/greenfield/x/payment/types"
)

func main() {
	account, err := bsctypes.NewBscAccountFromPrivateKey("barry", privateKey)
	if err != nil {
		log.Fatalf("New account from private key error, %v", err)
	}

	client, err := bsc.New(rpcAddr, "qa-net", bsc.Option{DefaultAccount: account})
	if err != nil {
		log.Fatalf("unable to new greenfield client, %v", err)
	}

	relayFee, minAckRelayFee, err := client.GetMinAckRelayFee(context.Background())

	messages := bsctypes.NewExecutorBatchedMessage(client.GetDeployment(), relayFee, minAckRelayFee)
	messages.CreatePaymentAccount(&types.MsgCreatePaymentAccount{Creator: account.GetAddress().String()})

	tx, err := client.Execute(context.Background(), messages.Build())
	if err != nil {
		log.Fatalf("unable to send messages, %v", err)
	}

	success, err := client.CheckTxStatus(context.Background(), tx)
	if err != nil {
		log.Fatalf("unable to check tx status, %v", err)
	}

	if success {
		log.Println("successfully sent the tx")
	} else {
		log.Println("failed to send the tx")
	}
}
