# Greenfield Go SDK

## Instruction

The Greenfield-GO-SDK provides a thin wrapper for interacting with greenfield storage network. 

Rich SDKs is provided to operate Greenfield resources or query status of resources.

### Requirement

Go version above 1.20

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
cosmossdk.io/api => github.com/bnb-chain/greenfield-cosmos-sdk/api v0.0.0-20230816082903-b48770f5e210
cosmossdk.io/math => github.com/bnb-chain/greenfield-cosmos-sdk/math v0.0.0-20230816082903-b48770f5e210
github.com/cometbft/cometbft => github.com/bnb-chain/greenfield-cometbft v1.1.0
github.com/cometbft/cometbft-db => github.com/bnb-chain/greenfield-cometbft-db v0.8.1-alpha.1
github.com/cosmos/cosmos-sdk => github.com/bnb-chain/greenfield-cosmos-sdk v1.1.0
github.com/cosmos/iavl => github.com/bnb-chain/greenfield-iavl v0.20.1
github.com/syndtr/goleveldb => github.com/syndtr/goleveldb v1.0.1-0.20210819022825-2ae1ddf74ef7
github.com/consensys/gnark-crypto => github.com/consensys/gnark-crypto v0.7.0
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
	
	gnfdCLient, err := client.New(chainId, rpcAddr, client.Option{DefaultAccount: account})
	if err != nil {
		log.Fatalf("unable to new greenfield client, %v", err)
	}
}

```

###  Quick Start Examples

The examples directory provides a wealth of examples to guide users in using the SDK's various features, including basic storage upload and download functions, 
group functions, permission functions, as well as payment and cross-chain related functions.

The **basic.go** includes the basic functions to fetch the blockchain info.

The **storage.go** includes the most storage functions such as creating a bucket, uploading files, downloading files, heading and deleting resource.

The **group.go** includes the group related functions such as creating a group and updating group member.

The **payment.go** includes the payment related functions to manage payment accounts.

The **permission.go** includes the permission related functions to manage resources(bucket, object, group) policy.

The **crosschain.go** includes the cross chain related functions to transfer or mirror resource to BSC.


#### Config Examples

You need to modify the variables in "common.go" under the "examples" directory to set the initialization information for the client, including "rpcAddr", "chainId", and "privateKey", etc. In addition, 
you also need to set basic parameters such as "bucket name" and "object name" to run the basic functionality of storage.

#### Run Examples
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