package main

import "log"

// The config information is consistent with the testnet of greenfield

const (
	rpcAddr     = "gnfd-testnet-fullnode-cosmos-us.bnbchain.org:9090"
	chainId     = "greenfield_5600-1"
	privateKey  = "xxx"
	objectSize  = 1000
	groupMember = "0x.."
	principal   = "0x.."
)

func HandleErr(err error, funcName string) {
	if err != nil {
		log.Fatalln("fail to " + funcName + ": " + err.Error())
	}
}
