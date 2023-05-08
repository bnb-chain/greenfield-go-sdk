package main

import (
	"context"
	"log"

	"github.com/bnb-chain/greenfield-go-sdk/client"
)

func testBasic(cli client.Client) {
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
