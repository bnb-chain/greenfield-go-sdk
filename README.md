# Greenfield Go SDK


## Disclaimer
**The software and related documentation are under active development, all subject to potential future change without
notification and not ready for production use. The code and security audit have not been fully completed and not ready
for any bug bounty. We advise you to be careful and experiment on the network at your own risk. Stay safe out there.**

## Instruction

The Greenfield-GO-SDK provides a thin wrapper for interacting with greenfield storage network. 

Rich SDKs is provided to operate Greenfield resources, send txn or query status, mainly divided into the following categories: 

(1) SDKs for operating resources, such as creating buckets, objects, groups, and uploading files for basic storage functions; 

(2) Various functions for setting and modifying metadata, such as updating bucket information and setting permissions; 

(3) Query functions to get the resources status and state;

(4) Payment-related functions to support payment; 

(5) Cross-chain related functions to achieve cross-chain transfer and mirror functions; 

(6) Account related functions to operate greenfield account.

### Requirement

Go version above 1.18

## Getting started
To get started working with the SDK setup your project for Go modules, and retrieve the SDK dependencies with `go get`.
This example shows how you can use the greenfield go SDK to interact with the greenfield storage network,

### Initialize Project

```sh
$ mkdir ~/hellogreenfield
$ cd ~/hellogreenfield
$ go mod init hellogreenfield
```

### Add SDK Dependencies

```sh
$ go get github.com/bnb-chain/greenfield-go-sdk
```

replace dependencies

```go.mod
cosmossdk.io/math => github.com/bnb-chain/greenfield-cosmos-sdk/math v0.0.0-20230228075616-68ac309b432c
github.com/cosmos/cosmos-sdk => github.com/bnb-chain/greenfield-cosmos-sdk v0.0.13
github.com/gogo/protobuf => github.com/regen-network/protobuf v1.3.3-alpha.regen.1
github.com/tendermint/tendermint => github.com/bnb-chain/greenfield-tendermint v0.0.3
```

### Initialize Client

The greenfield client requires the following parameters to connect to greenfield chain and storage providers.

| Parameter             | Description                                       |
|:----------------------|:--------------------------------------------------|
| rpcAddr               | the tendermit address of greenfield chain         |
| chainId               | the chain id of greenfield                        |
| client.Option  | All the options such as DefaultAccount and secure |

The DefaultAccount is need to set in options if you need send request to SP or send txn to Greenfield
```go
package main

import (
	"context"
	"log"

	"github.com/bnb-chain/greenfield-go-sdk/client"
	"github.com/bnb-chain/greenfield-go-sdk/types"
)

func main() {
	privateKey := "<Your own private key>"
	account, err := types.NewAccountFromPrivateKey("test", privateKey)
	if err != nil {
		log.Fatalf("New account from private key error, %v", err)
	}

	rpcAddr := "https://gnfd-testnet-fullnode-tendermint-us.bnbchain.org:443"
	chainId := "greenfield_5600-1"
	
	gnfdCLient, err := client.New(rpcAddr, chainId, client.Option{DefaultAccount: account})
	if err != nil {
		log.Fatalf("unable to new greenfield client, %v", err)
	}
}

```

###  Quick Start Example

The "examples" directory provides a wealth of examples to guide users in using the SDK's various features, including basic storage upload and download functionality, 
group functionality, permission functionality, as well as payment and cross-chain related functionality.

We recommend becoming familiar with the "storage.go" example first, as it includes the most basic operations such as creating a bucket, uploading files, downloading files, and accessing resource headers.

#### config example

You need to modify the variables in "common.go" under the "examples" directory to set the initialization information for the client, including "rpcAddr", "chainId", and "privateKey", etc. In addition, 
you also need to set basic parameters such as "bucket name" and "object name" to run the basic functionality of storage.

### run examples
The steps to run example are as follows
```
make examples
cd examples
./storage 
```
You can also directly execute "go run" to run a specific example. 
For example, execute "go run storage.go common.go" to run the relevant example for storage.
Please note that the "permission.go" example must be run after "storage.go" because resources such as objects need to be created first before setting permissions.

## Reference

- [Greenfield](https://github.com/bnb-chain/greenfield): the greenfield blockchain
- [Greenfield-Contract](https://github.com/bnb-chain/greenfield-contracts): the cross chain contract for Greenfield that deployed on BSC network.
- [Greenfield-Tendermint](https://github.com/bnb-chain/greenfield-tendermint): the consensus layer of Greenfield blockchain.
- [Greenfield-Storage-Provider](https://github.com/bnb-chain/greenfield-storage-provider): the storage service infrastructures provided by either organizations or individuals.
- [Greenfield-Relayer](https://github.com/bnb-chain/greenfield-relayer): the service that relay cross chain package to both chains.
- [Greenfield-Cmd](https://github.com/bnb-chain/greenfield-cmd): the most powerful command line to interact with Greenfield system.
- [Awesome Cosmos](https://github.com/cosmos/awesome-cosmos): Collection of Cosmos related resources which also fits Greenfield.