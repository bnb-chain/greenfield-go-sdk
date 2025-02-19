package main

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"log"
	"time"

	"github.com/bnb-chain/greenfield-go-sdk/client"
	"github.com/bnb-chain/greenfield-go-sdk/types"
)

// The example demonstrates the use of the Delegate upload functionality, which streamlines the process compared to the
// traditional two-stage method of creating and then uploading. With Delegated upload, the uploader interacts solely with
// the SP without the need to calculate the checksum and create the object meta on-chain initially.
// Instead, the primary SP handles these on behalf of the uploader.
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

	// get storage providers list
	spLists, err := cli.ListStorageProviders(ctx, true)
	if err != nil {
		log.Fatalf("fail to list in service sps")
	}
	// choose the first sp to be the primary SP
	primarySP := spLists[0].GetOperatorAddress()

	// create bucket
	_, err = cli.CreateBucket(ctx, bucketName, primarySP, types.CreateBucketOptions{})
	handleErr(err, "CreateBucket")
	log.Printf("create bucket %s on SP: %s successfully \n", bucketName, spLists[0].Endpoint)

	// head bucket
	bucketInfo, err := cli.HeadBucket(ctx, bucketName)
	handleErr(err, "HeadBucket")
	log.Println("bucket info:", bucketInfo.String())

	// Create object content
	var buffer bytes.Buffer
	line := `0123456789`
	for i := 0; i < 1024; i++ {
		buffer.WriteString(fmt.Sprintf("%s", line))
	}

	err = cli.DelegatePutObject(ctx, bucketName, objectName, int64(buffer.Len()), bytes.NewReader(buffer.Bytes()), types.PutObjectOptions{})
	handleErr(err, "DelegatePutObject")

	log.Printf("object: %s has been uploaded to SP\n", objectName)

	waitObjectSeal(cli, bucketName, objectName)
	// wait for block_syncer to sync up data from chain
	time.Sleep(time.Second * 5)

	// get object
	reader, info, err := cli.GetObject(ctx, bucketName, objectName, types.GetObjectOptions{})
	handleErr(err, "GetObject")
	log.Printf("get object %s successfully, size %d \n", info.ObjectName, info.Size)
	handleErr(err, "GetObject")
	objectBytes, err := io.ReadAll(reader)
	if !bytes.Equal(objectBytes, buffer.Bytes()) {
		handleErr(errors.New("download content not same"), "GetObject")
	}

	// the updated content of object
	var newBuffer bytes.Buffer
	for i := 0; i < 2048; i++ {
		newBuffer.WriteString(fmt.Sprintf("%s", line))
	}

	err = cli.DelegateUpdateObjectContent(ctx, bucketName, objectName, int64(newBuffer.Len()), bytes.NewReader(newBuffer.Bytes()), types.PutObjectOptions{})
	handleErr(err, "DelegateUpdateObjectContent")

	log.Printf("object: %s has been updated to SP\n", objectName)

	waitObjectSeal(cli, bucketName, objectName)
	// wait for block_syncer to sync up data from chain
	time.Sleep(time.Second * 5)

	// get object
	reader, info, err = cli.GetObject(ctx, bucketName, objectName, types.GetObjectOptions{})
	handleErr(err, "GetObject")
	log.Printf("get object %s successfully, size %d \n", info.ObjectName, info.Size)
	handleErr(err, "GetObject")
	objectBytes, err = io.ReadAll(reader)
	if !bytes.Equal(objectBytes, newBuffer.Bytes()) {
		handleErr(errors.New("download content not same"), "GetObject")
	}

}
