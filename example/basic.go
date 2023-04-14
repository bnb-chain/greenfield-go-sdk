package main

import (
	"context"
	"log"

	"github.com/bnb-chain/greenfield-go-sdk/client"
	"github.com/bnb-chain/greenfield-go-sdk/types"
)

func main() {
	privateKey := "<Your own private key>"
	account, err := types.NewAccountFromPrivateKey("test", privateKey)
	if err != nil {
		log.Fatalf("New account from private key error, %v", err)
	}
	cli, err := client.New("greenfield_9000-121", "localhost:9090", client.Option{DefaultAccount: account})
	if err != nil {
		log.Fatalf("unable to new greenfield client, %v", err)
	}
	ctx := context.Background()
	nodeInfo, versionInfo, err := cli.GetNodeInfo(ctx)
	if err != nil {
		log.Fatalf("unable to get node info, %v", err)
	}
	log.Printf("nodeInfo moniker: %s, go version: %s", nodeInfo.Moniker, versionInfo.GoVersion)
	latestBlock, err := cli.GetLatestBlock(ctx)
	if err != nil {
		log.Fatalf("unable to get latest block, %v", err)
	}
	log.Printf("latestBlock header: %s", latestBlock.Header.String())

	heightBefore := latestBlock.Header.Height
	log.Printf("Wait for block height: %d", heightBefore)
	err = cli.WaitForBlockHeight(ctx, heightBefore+10)
	if err != nil {
		log.Fatalf("unable to wait for block height, %v", err)
	}
	height, err := cli.GetLatestBlockHeight(ctx)
	if err != nil {
		log.Fatalf("unable to get latest block height, %v", err)
	}

	log.Printf("Current block height: %d", height)
}
