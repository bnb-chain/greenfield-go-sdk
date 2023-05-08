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

func testStorage(cli client.Client, bucketName, objectName string) {
	ctx := context.Background()

	// get storage providers list
	spLists, err := cli.ListStorageProviders(ctx, true)
	if err != nil {
		log.Fatalf("fail to list in service sps")
	}

	primarySP := spLists[3].GetOperatorAddress()

	// create bucket
	_, err = cli.CreateBucket(ctx, bucketName, primarySP, types.CreateBucketOptions{})
	HandleErr(err, "CreateBucket")
	log.Printf("create bucket %s on SP: %s successfully \n", bucketName, spLists[3].Endpoint)

	// head bucket
	bucketInfo, err := cli.HeadBucket(ctx, bucketName)
	HandleErr(err, "HeadBucket")
	log.Println("bucket info:", bucketInfo.String())

	// Create object content
	var buffer bytes.Buffer
	line := `0123456789`
	for i := 0; i < objectSize/10; i++ {
		buffer.WriteString(fmt.Sprintf("%s", line))
	}

	// create and put object
	txnHash, err := cli.CreateObject(ctx, bucketName, objectName, bytes.NewReader(buffer.Bytes()), types.CreateObjectOptions{})
	HandleErr(err, "CreateObject")

	err = cli.PutObject(ctx, bucketName, objectName, int64(buffer.Len()),
		bytes.NewReader(buffer.Bytes()), types.PutObjectOptions{TxnHash: txnHash})
	HandleErr(err, "PutObject")

	waitObjectSeal(cli, bucketName, objectName)

	// get object
	reader, info, err := cli.GetObject(ctx, bucketName, objectName, types.GetObjectOption{})
	HandleErr(err, "GetObject")
	log.Printf("get object %s successfully, size %d \n", info.ObjectName, info.Size)
	HandleErr(err, "GetObject")
	objectBytes, err := io.ReadAll(reader)
	if !bytes.Equal(objectBytes, buffer.Bytes()) {
		HandleErr(errors.New("download content not same"), "GetObject")
	}

	// list object
	objects, err := cli.ListObjects(ctx, bucketName, types.ListObjectsOptions{})
	log.Println("list objects")
	for _, obj := range objects.Objects {
		i := obj.ObjectInfo
		log.Printf("object: %s, status: %s\n", i.ObjectName, i.ObjectStatus)
	}

}

func testGroup(cli client.Client, groupName string) {
	ctx := context.Background()

	// create group
	groupTx, err := cli.CreateGroup(ctx, groupName, types.CreateGroupOptions{})
	HandleErr(err, "CreateGroup")
	_, err = cli.WaitForTx(ctx, groupTx)
	if err != nil {
		log.Fatalln("txn fail")
	}

	log.Printf("create group %s successfully \n", groupName)

	// head group
	creator, err := cli.GetDefaultAccount()
	HandleErr(err, "GetDefaultAccount")
	groupInfo, err := cli.HeadGroup(ctx, groupName, creator.GetAddress().String())
	HandleErr(err, "HeadGroup")
	log.Println("head group info:", groupInfo.String())

	// update group member
	updateTx, err := cli.UpdateGroupMember(ctx, groupName, creator.GetAddress().String(), []string{groupMember}, []string{},
		types.UpdateGroupMemberOption{})
	HandleErr(err, "UpdateGroupMember")
	_, err = cli.WaitForTx(ctx, updateTx)
	if err != nil {
		log.Fatalln("txn fail")
	}

	// head group member
	memIsExist := cli.HeadGroupMember(ctx, groupName, creator.GetAddress().String(), groupMember)
	if !memIsExist {
		log.Fatalf("head group member %s fail \n", groupMember)
	}
}

func waitObjectSeal(cli client.Client, bucketName, objectName string) {
	ctx := context.Background()
	// wait for the object to be sealed
	timeout := time.After(15 * time.Second)
	ticker := time.NewTicker(2 * time.Second)

	for {
		select {
		case <-timeout:
			err := errors.New("object not sealed after 15 seconds")
			HandleErr(err, "HeadObject")
		case <-ticker.C:
			objectInfo, err := cli.HeadObject(ctx, bucketName, objectName)
			HandleErr(err, "HeadObject")
			if objectInfo.GetObjectStatus().String() == "OBJECT_STATUS_SEALED" {
				ticker.Stop()
				fmt.Printf("put object %s successfully \n", objectName)
				return
			}
		}
	}
}
