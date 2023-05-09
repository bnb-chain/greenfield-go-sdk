package main

import "log"

// The config information is consistent with the testnet of greenfield

const (
	rpcAddr     = "https://gnfd.qa.bnbchain.world:443"
	chainId     = "greenfield_9000-1741"
	privateKey  = "1027c7d18143a02a7d0440b02000f7fa6e43dbd6da51d215fe0e300445be830e"
	objectSize  = 1000
	groupMember = "0x.."
	principal   = "0x.."
	toAddress   = "0x76d244CE05c3De4BbC6fDd7F56379B145709ade9"
)

func HandleErr(err error, funcName string) {
	if err != nil {
		log.Fatalln("fail to " + funcName + ": " + err.Error())
	}
}
