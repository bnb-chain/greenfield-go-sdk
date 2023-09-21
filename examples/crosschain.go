package main

import (
	"context"
	"log"

	"cosmossdk.io/math"
	"github.com/bnb-chain/greenfield-go-sdk/client"
	"github.com/bnb-chain/greenfield-go-sdk/types"
	gnfdSdkTypes "github.com/bnb-chain/greenfield/sdk/types"
	storageTestUtil "github.com/bnb-chain/greenfield/testutil/storage"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// it is the example of cross-chain SDKs usage
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

	// cross chain transfer to an account in BSC
	transferOut(cli, ctx)

	// mirror resource to BSC
	mirrorBucket(cli, ctx)
}

func transferOut(cli client.IClient, ctx context.Context) {
	// cross chain transfer to BSC
	txResp, err := cli.TransferOut(ctx, toAddress, math.NewInt(123456), gnfdSdkTypes.TxOption{})
	handleErr(err, "crossChainTransfer")
	waitForTx, _ := cli.WaitForTx(ctx, txResp.TxHash)
	log.Printf("Wait for tx: %s", waitForTx.TxResult.String())
	log.Printf("the tx log is %s", txResp.String())
}

func mirrorBucket(cli client.IClient, ctx context.Context) {
	// get storage providers list
	spLists, err := cli.ListStorageProviders(ctx, true)
	if err != nil {
		log.Fatalf("fail to list in service sps")
	}
	// choose the first sp to be the primary SP
	primarySP := spLists[0].GetOperatorAddress()

	bucketName := storageTestUtil.GenRandomBucketName()

	txHash, err := cli.CreateBucket(ctx, bucketName, primarySP, types.CreateBucketOptions{})
	handleErr(err, "CreateBucket")
	log.Printf("create bucket %s on SP: %s successfully \n", bucketName, spLists[0].Endpoint)

	waitForTx, _ := cli.WaitForTx(ctx, txHash)
	log.Printf("Wait for tx: %s", waitForTx.TxResult.String())

	// head bucket
	bucketInfo, err := cli.HeadBucket(ctx, bucketName)
	handleErr(err, "HeadBucket")
	log.Println("bucket info:", bucketInfo.String())

	// mirror bucket
	txResp, err := cli.MirrorBucket(ctx, sdk.ChainID(crossChainDestBsChainId), bucketInfo.Id, bucketName, gnfdSdkTypes.TxOption{})
	handleErr(err, "MirrorBucket")
	waitForTx, _ = cli.WaitForTx(ctx, txResp.TxHash)
	log.Printf("Wait for tx: %s", waitForTx.TxResult.String())
	log.Printf("successfully mirrored bucket wiht bucket id %s to BSC", bucketInfo.Id)
}
