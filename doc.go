// Package sdk is the official GreenField SDK for the Go programming language.
//
// # Getting started
//
// To get started working with the SDK setup your project for Go modules, and retrieve the SDK dependencies with `go get`.
// This example shows how you can use the SDK to make an API request using the SDK's client.
//
// # Hello GreenField
//
//	package main
//
//	import (
//
//		"context"
//		"log"
//		"github.com/bnb-chain/greenfield-go-sdk/client"
//		"github.com/bnb-chain/greenfield-go-sdk/types"
//		"google.golang.org/grpc"
//		"google.golang.org/grpc/credentials/insecure"
//
//	)
//
//	func main() {
//		privateKey := "9579fff0cab07a4379e845a890105004ba4c8276f8ad9d22082b2acbf02d884b"
//		account, err := types.NewAccountFromPrivateKey("test", privateKey)
//		if err != nil {
//			log.Fatalf("New account from private key error, %v", err)
//		}
//		cli, err := client.New("greenfield_9000-121", "localhost:9090", account, &client.Option{GrpcDialOption: grpc.WithTransportCredentials(insecure.NewCredentials())})
//		if err != nil {
//			log.Fatalf("unable to new greenfield client, %v", err)
//		}
//		ctx := context.Background()
//		nodeInfo, versionInfo, err := cli.GetNodeInfo(ctx)
//		if err != nil {
//			log.Fatalf("unable to get node info, %v", err)
//		}
//		log.Printf("nodeInfo moniker: %s, go version: %s", nodeInfo.Moniker, versionInfo.GoVersion)
//		latestBlock, err := cli.LatestBlock(ctx)
//		if err != nil {
//			log.Fatalf("unable to get latest block, %v", err)
//		}
//		log.Printf("latestBlock header: %s", latestBlock.Header.String())
//
//		heightBefore := latestBlock.Header.Height
//		log.Printf("Wait for block height: %d", heightBefore)
//		err = cli.WaitForBlockHeight(ctx, heightBefore+10)
//		if err != nil {
//			log.Fatalf("unable to wait for block height, %v", err)
//		}
//		height, err := cli.LatestBlockHeight(ctx)
//		if err != nil {
//			log.Fatalf("unable to get latest block height, %v", err)
//		}
//		log.Printf("Current block height: %d", height)
//	}
package sdk
