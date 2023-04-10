# Greenfield Go SDK


## Disclaimer
**The software and related documentation are under active development, all subject to potential future change without
notification and not ready for production use. The code and security audit have not been fully completed and not ready
for any bug bounty. We advise you to be careful and experiment on the network at your own risk. Stay safe out there.**

## Instruction

The Greenfield-GO-SDK provides a thin wrapper for interacting with greenfield storage network. 

### Requirement

Go version above 1.18

## Getting started
To get started working with the SDK setup your project for Go modules, and retrieve the SDK dependencies with `go get`.
This example shows how you can use the greenfield go SDK to interact with the greenfield storage network,

###### Initialize Project

```sh
$ mkdir ~/hellogreenfield
$ cd ~/hellogreenfield
$ go mod init hellogreenfield
```

###### Add SDK Dependencies

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

###### Code

In your preferred editor add the following content to `main.go`

```go
package main

import (
	"context"
	"log"

	"github.com/bnb-chain/greenfield-go-sdk/client"
	"github.com/bnb-chain/greenfield-go-sdk/types"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func main() {
	// Generate an account from private key
	privateKey := "9579fff0cab07a4379e845a890105004ba4c8276f8ad9d22082b2acbf02d884b"
	account, err := types.NewAccountFromPrivateKey("test", privateKey)
	if err != nil {
		log.Fatalf("New account from private key error, %v", err)
	}
	
	cli, err := client.New("greenfield_9000-121", "localhost:9090", account, &client.Option{GrpcDialOption: grpc.WithTransportCredentials(insecure.NewCredentials())})
	if err != nil {
		log.Fatalf("unable to new greenfield client, %v", err)
	}
	ctx := context.Background()
	nodeInfo, versionInfo, err := cli.GetNodeInfo(ctx)
	if err != nil {
		log.Fatalf("unable to get node info, %v", err)
	}
	log.Printf("nodeInfo moniker: %s, go version: %s", nodeInfo.Moniker, versionInfo.GoVersion)
	latestBlock, err := cli.GetLatestBlock(ctx)
	if err != nil {
		log.Fatalf("unable to get latest block, %v", err)
	}
	log.Printf("latestBlock header: %s", latestBlock.Header.String())

	heightBefore := latestBlock.Header.Height
	log.Printf("Wait for block height: %d", heightBefore)
	err = cli.WaitForBlockHeight(ctx, heightBefore+10)
	if err != nil {
		log.Fatalf("unable to wait for block height, %v", err)
	}
	height, err := cli.GetLatestBlockHeight(ctx)
	if err != nil {
		log.Fatalf("unable to get latest block height, %v", err)
	}

	log.Printf("Current block height: %d", height)
}

```

###### Compile and Execute

```sh
$ go run .
2023/04/06 19:49:28 nodeInfo moniker: validator0, version: go version go1.19.4 darwin/arm64
2023/04/06 19:49:28 latestBlock header: version:<block:11 > chain_id:"greenfield_9000-121" height:111000 time:<seconds:1680781768 nanos:251685000 > last_block_id:<hash:"\035\240\267\r\212\271x\024< \317\212\212\354\230\305\202\004\337y\260\334Qb\020C\315\032J&\236\213" part_set_header:<total:1 hash:"\222N\324\331X@\327E\010\035\206\253\204 \377\242Yw\0278\022!2j\370}u]lx\214>" > > last_commit_hash:"\022=!\tBf\301\344.\241\235fx\324\n\334\374V\313@ae\313\364\030\260\374\267\256t:\026" data_hash:"\343\260\304B\230\374\034\024\232\373\364\310\231o\271$'\256A\344d\233\223L\244\225\231\033xR\270U" validators_hash:"\377cq\253\322\311\231\313\024\325\272\2266\023\017>\307\343\365\213\326\337\271\301\r0\320A\374Ed\266" next_validators_hash:"\377cq\253\322\311\231\313\024\325\272\2266\023\017>\307\343\365\213\326\337\271\301\r0\320A\374Ed\266" consensus_hash:")M\217\275\013\224\267g\247\353\251\204\017)\2325\206\332\177\346\265\336\255;~\354\272\031<@\017\223" app_hash:"\372&\252\033[a'^\313=TJU\305\017N\225cr|#T|\335\213\356=\241\330O\006x" last_results_hash:"\343\260\304B\230\374\034\024\232\373\364\310\231o\271$'\256A\344d\233\223L\244\225\231\033xR\270U" evidence_hash:"\343\260\304B\230\374\034\024\232\373\364\310\231o\271$'\256A\344d\233\223L\244\225\231\033xR\270U" proposer_address:"0x21562548eAd41732C4614e09152624d0A50d9593" 
2023/04/06 19:49:28 Wait for block height: 111000
2023/04/06 19:49:34 Current block height: 111011
```


## Reference

- [Greenfield](https://github.com/bnb-chain/greenfield): the greenfield blockchain
- [Greenfield-Contract](https://github.com/bnb-chain/greenfield-contracts): the cross chain contract for Greenfield that deployed on BSC network.
- [Greenfield-Tendermint](https://github.com/bnb-chain/greenfield-tendermint): the consensus layer of Greenfield blockchain.
- [Greenfield-Storage-Provider](https://github.com/bnb-chain/greenfield-storage-provider): the storage service infrastructures provided by either organizations or individuals.
- [Greenfield-Relayer](https://github.com/bnb-chain/greenfield-relayer): the service that relay cross chain package to both chains.
- [Greenfield-Cmd](https://github.com/bnb-chain/greenfield-cmd): the most powerful command line to interact with Greenfield system.
- [Awesome Cosmos](https://github.com/cosmos/awesome-cosmos): Collection of Cosmos related resources which also fits Greenfield.