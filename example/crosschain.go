package main

import (
	"context"
	"cosmossdk.io/math"
	"github.com/bnb-chain/greenfield-go-sdk/client"
	"github.com/bnb-chain/greenfield-go-sdk/types"
	gnfdSdkTypes "github.com/bnb-chain/greenfield/sdk/types"
	storageTestUtil "github.com/bnb-chain/greenfield/testutil/storage"
	"log"
)

func crossChainTransfer(cli client.Client) {
	ctx := context.Background()
	txResp, err := cli.TransferOut(ctx, toAddress, math.NewInt(123456), gnfdSdkTypes.TxOption{})
	HandleErr(err, "crossChainTransfer")
	waitForTx, _ := cli.WaitForTx(ctx, txResp.TxHash)
	log.Printf("Wait for tx: %s", waitForTx.String())
	log.Printf("the tx log is %s", txResp.String())
}

func mirrorBucket(cli client.Client) {
	ctx := context.Background()

	// get storage providers list
	spLists, err := cli.ListStorageProviders(ctx, true)
	if err != nil {
		log.Fatalf("fail to list in service sps")
	}
	// choose the first sp to be the primary SP
	primarySP := spLists[0].GetOperatorAddress()

	bucketName := storageTestUtil.GenRandomBucketName()

	txHash, err := cli.CreateBucket(ctx, bucketName, primarySP, types.CreateBucketOptions{})
	HandleErr(err, "CreateBucket")
	log.Printf("create bucket %s on SP: %s successfully \n", bucketName, spLists[0].Endpoint)

	waitForTx, _ := cli.WaitForTx(ctx, txHash)
	log.Printf("Wait for tx: %s", waitForTx.String())

	// head bucket
	bucketInfo, err := cli.HeadBucket(ctx, bucketName)
	HandleErr(err, "HeadBucket")
	log.Println("bucket info:", bucketInfo.String())

	// mirror bucket
	txResp, err := cli.MirrorBucket(ctx, bucketInfo.Id, gnfdSdkTypes.TxOption{})
	HandleErr(err, "MirrorBucket")
	waitForTx, _ = cli.WaitForTx(ctx, txResp.TxHash)
	log.Printf("Wait for tx: %s", waitForTx.String())
	log.Printf("successfully mirrored bucket wiht bucket id %s to BSC", bucketInfo.Id)
}
