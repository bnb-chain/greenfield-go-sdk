package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"time"

	"github.com/bnb-chain/greenfield-go-sdk/client"
)

// The config information is consistent with the testnet of greenfield
// You need to set the privateKey, bucketName, objectName and groupName to make the basic examples work well
const (
	rpcAddr                 = "https://gnfd-testnet-fullnode-tendermint-us.bnbchain.org:443"
	chainId                 = "greenfield_5600-1"
	crossChainDestBsChainId = 97
	privateKey              = "xx"
	objectSize              = 1000
	groupMember             = "0x.." // used for group examples
	principal               = "0x.." // used for permission examples
	bucketName              = "test-bucket"
	objectName              = "test-object"
	groupName               = "test-group"
	toAddress               = "0x.." // used for cross chain transfer
	httpsAddr               = ""
	paymentAddr             = ""
)

func handleErr(err error, funcName string) {
	if err != nil {
		log.Fatalln("fail to " + funcName + ": " + err.Error())
	}
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
