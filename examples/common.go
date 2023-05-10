package main

import (
	"log"
)

// The config information is consistent with the testnet of greenfield
// You need to set the privateKey, bucketName, objectName and groupName to make the basic examples work well
const (
	rpcAddr     = "https://gnfd.qa.bnbchain.world:443"
	chainId     = "greenfield_9000-1741"
	privateKey  = "xx"
	objectSize  = 1000
	groupMember = "0x.." // used for group examples
	principal   = "0x.." // used for permission examples
	bucketName  = "test-bucket"
	objectName  = "test-object"
	groupName   = "test-group"
	toAddress   = "0x.." // used for cross chain transfer
)

func handleErr(err error, funcName string) {
	if err != nil {
		log.Fatalln("fail to " + funcName + ": " + err.Error())
	}
}
