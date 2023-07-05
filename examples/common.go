package main

import (
	"log"
)

// The config information is consistent with the testnet of greenfield
// You need to set the privateKey, bucketName, objectName and groupName to make the basic examples work well
const (
	rpcAddr     = "https://gnfd-testnet-fullnode-tendermint-us.bnbchain.org:443"
	chainId     = "greenfield_5600-1"
	privateKey  = "4490e6024ace246eafa46b3b6d6bda1df443475f77c17b197c230cf7f244941b"
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
