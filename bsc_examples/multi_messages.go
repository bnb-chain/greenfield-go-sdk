package main

import (
	"context"
	"log"

	"github.com/bnb-chain/greenfield-go-sdk/bsc"
	"github.com/bnb-chain/greenfield-go-sdk/bsctypes"
)

func main() {
	account, err := bsctypes.NewBscAccountFromPrivateKey("barry", " ")
	if err != nil {
		log.Fatalf("New account from private key error, %v", err)
	}

	client, err := bsc.New(rpcAddr, chainId, bsc.Option{DefaultAccount: account})
	if err != nil {
		log.Fatalf("unable to new greenfield client, %v", err)
	}

	relayFee, minAckRelayFee, err := client.GetMinAckRelayFee()

	messages := bsctypes.NewMessages(client.GetDeployment(), relayFee, minAckRelayFee)

	_ = messages.CreateGroup(account.GetAddress(), account.GetAddress(), "barry-test")

	flag, err := client.SendMessages(context.Background(), messages.Build())
	if err != nil {
		log.Fatalf("unable to send messages, %v", err)
	}

	if flag {
		log.Println("successfully send messages")
	}
}
