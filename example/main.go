package main

import (
	"log"

	"github.com/bnb-chain/greenfield-go-sdk/client"
	"github.com/bnb-chain/greenfield-go-sdk/types"
	storageTestUtil "github.com/bnb-chain/greenfield/testutil/storage"
)

func main() {
	account, err := types.NewAccountFromPrivateKey("test", privateKey)
	if err != nil {
		log.Fatalf("New account from private key error, %v", err)
	}
	cli, err := client.New(chainId, rpcAddr, client.Option{DefaultAccount: account})
	if err != nil {
		log.Fatalf("unable to new greenfield client, %v", err)
	}
	bucketName := storageTestUtil.GenRandomBucketName()
	objectName := storageTestUtil.GenRandomObjectName()
	groupName := storageTestUtil.GenRandomGroupName()

	testBasic(cli)
	testStorage(cli, bucketName, objectName)
	testGroup(cli, groupName)
	testPermission(cli, bucketName, objectName)

	crossChainTransfer(cli)
	mirrorBucket(cli)
	payment(cli)
}
