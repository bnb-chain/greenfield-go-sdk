package main

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"github.com/bnb-chain/greenfield-go-sdk/client"
	"github.com/bnb-chain/greenfield-go-sdk/types"
	greenfield_types "github.com/bnb-chain/greenfield/types"
	storagetypes "github.com/bnb-chain/greenfield/x/storage/types"
	"time"

	"log"
	"testing"
)

// local
const rpcAddr = "http://localhost:26750"
const chainId = "greenfield_9000-121"

const privateKey = "35fba35f69ca14054ac66a8be988b02046826e3b2e64c60613c992258c2ca921"
const addressNew = "0x2412c55D1b613C87E37aF2172b05343551D19e98"
const primarySP = "0x5B2aee18007b54a97C899f375194a236368C5df4"
const endpoint = "http://127.0.0.1:9033"

func Test_BucketTags(t *testing.T) {

	account, err := types.NewAccountFromPrivateKey("test", privateKey)
	log.Println("test account addr is " + account.GetAddress().String())
	if err != nil {
		log.Fatalf("New account from private key error, %v", err)
	}

	cli, err := client.New(chainId, rpcAddr, client.Option{
		DefaultAccount: account,
	})
	if err != nil {
		log.Fatalf("unable to new greenfield client, %v", err)
	}
	ctx := context.Background()

	log.Println(primarySP)
	prefix := "b4"

	chargedQuota := uint64(100)
	//ChargedQuota: chargedQuota

	// case1:  create bucket without tag
	_, err = cli.CreateBucket(ctx, prefix+"tag01", primarySP, types.CreateBucketOptions{
		ChargedQuota: chargedQuota,
	})
	if err != nil {
		log.Println(err)
	}

	// case2:  create bucket with tag

	var tags storagetypes.ResourceTags
	tags.Tags = append(tags.Tags, storagetypes.ResourceTags_Tag{Key: "key1", Value: "value1"})
	tags.Tags = append(tags.Tags, storagetypes.ResourceTags_Tag{Key: "key2", Value: "value2"})
	_, err = cli.CreateBucket(ctx, prefix+"tag02", primarySP, types.CreateBucketOptions{
		ChargedQuota: chargedQuota,
		Tags:         &tags,
	})
	if err != nil {
		log.Println(err)
	}
	// case3:  create bucket with empty tag

	var emptyTags storagetypes.ResourceTags
	_, err = cli.CreateBucket(ctx, prefix+"tag03", primarySP, types.CreateBucketOptions{
		ChargedQuota: chargedQuota,
		Tags:         &emptyTags,
	})
	if err != nil {
		log.Println(err)
	}

	// case 4: set tags
	_, err = cli.CreateBucket(ctx, prefix+"tag04", primarySP, types.CreateBucketOptions{
		ChargedQuota: chargedQuota,
	})
	if err != nil {
		log.Println(err)
	}
	// Set tag
	grn := greenfield_types.NewBucketGRN(prefix + "tag04")
	var setTags storagetypes.ResourceTags
	setTags.Tags = append(tags.Tags, storagetypes.ResourceTags_Tag{Key: "tag04", Value: "tag04_value"})
	tx, err := cli.SetTag(ctx, grn.String(), setTags, types.SetTagsOptions{})
	if err != nil {
		log.Println(err)
	} else {
		log.Println("set tag tx: " + tx)
	}
	//===========
}

func Test_ObjectTags(t *testing.T) {

	account, err := types.NewAccountFromPrivateKey("test", privateKey)
	log.Println("test account addr is " + account.GetAddress().String())
	if err != nil {
		log.Fatalf("New account from private key error, %v", err)
	}

	cli, err := client.New(chainId, rpcAddr, client.Option{
		DefaultAccount: account,
	})
	if err != nil {
		log.Fatalf("unable to new greenfield client, %v", err)
	}
	ctx := context.Background()

	log.Println(primarySP)
	prefix := "ot005"
	bucket := prefix + "test-bucket"
	chargedQuota := uint64(100)
	_, err = cli.CreateBucket(ctx, bucket, primarySP, types.CreateBucketOptions{
		ChargedQuota: chargedQuota,
	})
	// Create object content
	var buffer bytes.Buffer
	line := `0123456789`
	// objectSize := 1 * 1024 * 1024 * 100 //  1*1024*1024*100 *  10 == 1GB
	objectSize := 1 * 100 //  1*1024*1024*100 *  10 == 1GB
	for i := 0; i < objectSize; i++ {
		buffer.WriteString(fmt.Sprintf("%s", line))
	}

	// case1:  create object without tag
	// create and put object
	txnHash, err := cli.CreateObject(ctx, bucket, prefix+"obj001", bytes.NewReader(buffer.Bytes()), types.CreateObjectOptions{
		ContentType: "text/plain",
		Visibility:  storagetypes.VISIBILITY_TYPE_PRIVATE,
	})

	handleErr(err, "CreateObject")
	err = cli.PutObject(ctx, bucket, prefix+"obj001", int64(buffer.Len()),
		bytes.NewReader(buffer.Bytes()), types.PutObjectOptions{TxnHash: txnHash})
	handleErr(err, "PutObject")
	log.Printf("object: %s has been uploaded to SP\n", prefix+"obj001")

	time.Sleep(time.Second * 10)

	waitObjectSeal(cli, bucket, prefix+"obj001")

	// case2:  create object with tag
	// create and put object

	var tags storagetypes.ResourceTags
	tags.Tags = append(tags.Tags, storagetypes.ResourceTags_Tag{Key: "key1", Value: "value1"})
	tags.Tags = append(tags.Tags, storagetypes.ResourceTags_Tag{Key: "key2", Value: "value2"})
	txnHash, err = cli.CreateObject(ctx, bucket, prefix+"obj002", bytes.NewReader(buffer.Bytes()), types.CreateObjectOptions{
		ContentType: "text/plain",
		Visibility:  storagetypes.VISIBILITY_TYPE_PRIVATE,
		Tags:        &tags,
	})

	handleErr(err, "CreateObject")
	err = cli.PutObject(ctx, bucket, prefix+"obj002", int64(buffer.Len()),
		bytes.NewReader(buffer.Bytes()), types.PutObjectOptions{TxnHash: txnHash})
	handleErr(err, "PutObject")
	log.Printf("object: %s has been uploaded to SP\n", prefix+"obj001")

	time.Sleep(time.Second * 10)

	waitObjectSeal(cli, bucket, prefix+"obj002")

	// case3:  create bucket with empty tag

	var emptytags storagetypes.ResourceTags
	txnHash, err = cli.CreateObject(ctx, bucket, prefix+"obj003", bytes.NewReader(buffer.Bytes()), types.CreateObjectOptions{
		ContentType: "text/plain",
		Visibility:  storagetypes.VISIBILITY_TYPE_PRIVATE,
		Tags:        &emptytags,
	})

	handleErr(err, "CreateObject")
	err = cli.PutObject(ctx, bucket, prefix+"obj003", int64(buffer.Len()),
		bytes.NewReader(buffer.Bytes()), types.PutObjectOptions{TxnHash: txnHash})
	handleErr(err, "PutObject")
	log.Printf("object: %s has been uploaded to SP\n", prefix+"obj001")

	time.Sleep(time.Second * 10)

	waitObjectSeal(cli, bucket, prefix+"obj003")

	// case 4: set tags
	txnHash, err = cli.CreateObject(ctx, bucket, prefix+"obj004", bytes.NewReader(buffer.Bytes()), types.CreateObjectOptions{
		ContentType: "text/plain",
		Visibility:  storagetypes.VISIBILITY_TYPE_PRIVATE,
	})

	handleErr(err, "CreateObject")
	err = cli.PutObject(ctx, bucket, prefix+"obj004", int64(buffer.Len()),
		bytes.NewReader(buffer.Bytes()), types.PutObjectOptions{TxnHash: txnHash})
	handleErr(err, "PutObject")
	log.Printf("object: %s has been uploaded to SP\n", prefix+"obj001")

	time.Sleep(time.Second * 10)

	waitObjectSeal(cli, bucket, prefix+"obj004")
	// Set tag
	grn := greenfield_types.NewObjectGRN(bucket, prefix+"obj004")
	var setTags storagetypes.ResourceTags
	setTags.Tags = append(tags.Tags, storagetypes.ResourceTags_Tag{Key: "tag04", Value: "tag04_value"})
	tx, err := cli.SetTag(ctx, grn.String(), setTags, types.SetTagsOptions{})
	if err != nil {
		log.Println(err)
	} else {
		log.Println("set tag tx: " + tx)
	}
	//===========

	// list object
	objects, err := cli.ListObjects(ctx, bucket, types.ListObjectsOptions{
		true, "", "", "/", "", 10, endpoint, "",
	})
	if err != nil {
		log.Println(err)
	}
	log.Println("list objects result:")
	for _, obj := range objects.Objects {
		i := obj.ObjectInfo
		log.Printf("object:%+v", i)
	}
}

func Test_GroupTags(t *testing.T) {

	account, err := types.NewAccountFromPrivateKey("test", privateKey)
	log.Println("test account addr is " + account.GetAddress().String())
	if err != nil {
		log.Fatalf("New account from private key error, %v", err)
	}

	cli, err := client.New(chainId, rpcAddr, client.Option{
		DefaultAccount: account,
	})
	if err != nil {
		log.Fatalf("unable to new greenfield client, %v", err)
	}
	ctx := context.Background()

	log.Println(primarySP)
	prefix := "g001"

	// case1:  create bucket without tag
	// create group
	groupTx, err := cli.CreateGroup(ctx, prefix+"-01", types.CreateGroupOptions{})
	handleErr(err, "CreateGroup")
	_, err = cli.WaitForTx(ctx, groupTx)
	if err != nil {
		log.Fatalln("txn fail")
	}

	// case2:  create bucket with tag

	var tags storagetypes.ResourceTags
	tags.Tags = append(tags.Tags, storagetypes.ResourceTags_Tag{Key: "key1", Value: "value1"})
	tags.Tags = append(tags.Tags, storagetypes.ResourceTags_Tag{Key: "key2", Value: "value2"})
	groupTx, err = cli.CreateGroup(ctx, prefix+"-02", types.CreateGroupOptions{
		Tags: &tags,
	})
	handleErr(err, "CreateGroup")
	_, err = cli.WaitForTx(ctx, groupTx)
	if err != nil {
		log.Fatalln("txn fail")
	}

	// case3:  create bucket with empty tag

	var emptyTags storagetypes.ResourceTags
	groupTx, err = cli.CreateGroup(ctx, prefix+"-03", types.CreateGroupOptions{
		Tags: &emptyTags,
	})
	handleErr(err, "CreateGroup")
	_, err = cli.WaitForTx(ctx, groupTx)
	if err != nil {
		log.Fatalln("txn fail")
	}

	// case 4: set tags
	groupTx, err = cli.CreateGroup(ctx, prefix+"-04", types.CreateGroupOptions{})
	handleErr(err, "CreateGroup")
	_, err = cli.WaitForTx(ctx, groupTx)
	if err != nil {
		log.Fatalln("txn fail")
	}
	// Set tag
	grn := greenfield_types.NewGroupGRN(account.GetAddress(), prefix+"-04")
	var setTags storagetypes.ResourceTags
	setTags.Tags = append(setTags.Tags, storagetypes.ResourceTags_Tag{Key: "tag04", Value: "tag04_value"})
	tx, err := cli.SetTag(ctx, grn.String(), setTags, types.SetTagsOptions{})
	if err != nil {
		log.Println(err)
	} else {
		log.Println("set tag tx: " + tx)
	}
	//===========
}

func waitObjectSeal(cli client.IClient, bucketName, objectName string) {
	ctx := context.Background()
	// wait for the object to be sealed
	timeout := time.After(15 * time.Second)
	ticker := time.NewTicker(2 * time.Second)

	for {
		select {
		case <-timeout:
			err := errors.New("object not sealed after 15 seconds")
			handleErr(err, "HeadObject")
		case <-ticker.C:
			objectDetail, err := cli.HeadObject(ctx, bucketName, objectName)
			handleErr(err, "HeadObject")
			if objectDetail.ObjectInfo.GetObjectStatus().String() == "OBJECT_STATUS_SEALED" {
				ticker.Stop()
				fmt.Printf("put object %s successfully \n", objectName)
				return
			}
		}
	}
}

func handleErr(err error, funcName string) {
	if err != nil {
		log.Fatalln("fail to " + funcName + ": " + err.Error())
	}
}
