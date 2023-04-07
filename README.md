# Greenfield Go SDK


## Disclaimer
**The software and related documentation are under active development, all subject to potential future change without
notification and not ready for production use. The code and security audit have not been fully completed and not ready
for any bug bounty. We advise you to be careful and experiment on the network at your own risk. Stay safe out there.**

## Instruction
The Greenfield-GO-SDK provides a thin wrapper for interacting with greenfield in three ways:

1.Interact using GreenfieldClient client, you can perform queries on accounts, chain info, and broadcasting transactions.

2.Interact using GnfdClient, it integrates GreenfieldClient and the ability to access Storage provider, 
you can call storage functions like createObject,putObject and getObject to realize the basic operation of object storage.

3.Interact using TendermintClient client, you can perform low-level operations like executing ABCI queries and viewing network/consensus state.


### Requirement

Go version above 1.18

## Usage

### Importing

```go
import (
    "github.com/bnb-chain/greenfield-go-sdk" latest
)
```
### Replace dependencies
```go
replace (
    cosmossdk.io/math => github.com/bnb-chain/greenfield-cosmos-sdk/math v0.0.0-20230228075616-68ac309b432c
    github.com/cosmos/cosmos-sdk => github.com/bnb-chain/greenfield-cosmos-sdk v0.0.13
    github.com/gogo/protobuf => github.com/regen-network/protobuf v1.3.3-alpha.regen.1
    github.com/tendermint/tendermint => github.com/bnb-chain/greenfield-tendermint v0.0.3
)
```

### Key Manager

A Key Manager is needed to sign transaction messages or verify signatures. The Key Manager is an Identity Manager used 
to define who you are in Greenfield. It provides the following interface:

```go
type KeyManager interface {
    GetPrivKey() ctypes.PrivKey
    GetAddr() types.AccAddress
}
```

We provide three construction functions to generate the Key Manager:

```go
NewPrivateKeyManager(priKey string) (KeyManager, error)

NewMnemonicKeyManager(mnemonic string) (KeyManager, error)
```

- NewPrivateKeyManager: You should provide a Hex encoded string of your private key.
- NewMnemonicKeyManager: You should provide your mnemonic, usually a string of 24 words.

Examples:

From private key hex string:
```GO
privateKey := "9579fff0cab07a4379e845a890105004ba4c8276f8ad9d22082b2acbf02d884b"
keyManager, err := NewPrivateKeyManager(privateKey)
```

From mnemonic:
```Go
mnemonic := "dragon shy author wave swamp avoid lens hen please series heavy squeeze alley castle crazy action peasant green vague camp mirror amount person legal"
keyManager, _ := keys.NewMnemonicKeyManager(mnemonic)
```

### Use Greenfield Client

#### Initialize client without a key manager; use it for querying purposes only.

```go
client := NewGreenfieldClient("localhost:9090", "greenfield_9000-121")

query := banktypes.QueryBalanceRequest{
		Address: "0x76d244CE05c3De4BbC6fDd7F56379B145709ade9",
		Denom:   "BNB",
}
res, err := client.BankQueryClient.Balance(context.Background(), &query)  
```

#### Initialize client with a key manager to sign and send transactions

```go
keyManager, _ := keys.NewPrivateKeyManager("ab463aca3d2965233da3d1d6108aa521274c5ddc2369ff72970a52a451863fbf")
gnfdClient := NewGreenfieldClient("localhost:9090", 
	                            "greenfield_9000-121",
	                            WithKeyManager(km),
                                    WithGrpcDialOption(grpc.WithTransportCredentials(insecure.NewCredentials()))
)
```

#### Broadcast TX

A generic method `BroadcastTx` is provided to give you the flexibility to broadcast different types of transactions.

```go
BroadcastTx(msgs []sdk.Msg, txOpt *types.TxOption, opts ...grpc.CallOption) (*tx.BroadcastTxResponse, error)
```

`txOpt` is provided to customize your transaction. All fields are optional.
```go
type TxOption struct {
    Mode      *tx.BroadcastMode   // default to `sync` mode
    GasLimit  uint64 // default to use simulated gas 
    Memo      string
    FeeAmount sdk.Coins
    FeePayer  sdk.AccAddress
    Nonce     uint64
}
```
Example:

```go
payerAddr, _ := sdk.AccAddressFromHexUnsafe("0x76d244CE05c3De4BbC6fDd7F56379B145709ade9")
transfer := banktypes.NewMsgSend(km.GetAddr(), to, sdk.NewCoins(sdk.NewInt64Coin("BNB", 12)))
broadcastMode := tx.BroadcastMode_BROADCAST_MODE_ASYNC
txOpt := &types.TxOption{
    Mode       &broadcastMode
    GasLimit:  1000000,
    Memo:      "test",
    FeePayer:  payerAddr,
}
response, _ := gnfdClient.BroadcastTx([]sdk.Msg{transfer}, txOpt)
```

#### Simulate TX

For the purpose of simulating a tx and get the gas info, `SimulateTx` is provided.

```go
SimulateTx(msgs []sdk.Msg, txOpt *types.TxOption, opts ...grpc.CallOption) (*tx.SimulateResponse, error)
```

### Sign Tx

`SignTx` function signs the provided messages and returns raw bytes.

```go
SignTx(msgs []sdk.Msg, txOpt *types.TxOption) ([]byte, error)
```

### Get Nonce

The `GetNonce` function retrieves the nonce of an account.

```go
GetNonce() (uint64, error)
```

#### Support transaction types
Please refer to [msgTypes.go](./types/msg_types.go) to see all of the supported types of `sdk.Msg`.

### Use GnfdClient

The construct fuction need to pass chain grpc address, chain id and SP endpoint
```go 
client, err := NewGnfdClient(grpcAddr, chainId, endpoint, keyManager, false,
          WithKeyManager(keyManager),
          WithGrpcDialOption(grpc.WithTransportCredentials(insecure.NewCredentials())))
```

#### Call APIs and Send Requests 

1) create bucket 
```
opts := CreateBucketOptions{ChargedQuota: chargeQuota, Visibility: &storageTypes.VISIBILITY_TYPE_PRIVATE}
txnHash, err = client.CreateBucket(ctx, bucketName, primarySp, opts)
 
// head bucket
bucketInfo, err := client.HeadBucket(ctx, bucketName)

```

2) two stages of uploading including createObject and putObject

```
// (1) create object on chain
txnHash, err = client.CreateObject(ctx, bucketName, objectName,
           bytes.NewReader(buffer.Bytes()), CreateObjectOptions{})
            
object, err := s.gnfdClient.HeadObject(ctx, bucketName, objectName)

// (2) upload payload to SP  

fileReader, err := os.Open(filePath)
err = s.gnfdClient.PutObject(ctx, bucketName, objectName, txnHash, fileSize,
        fileReader, PutObjectOption{})

```

3) get Object

```
body, objectInfo, err := s.gnfdClient.GetObject(ctx, bucketName, objectName, sp.GetObjectOption{})
objectBytes, err := io.ReadAll(body)
```

### Use Tendermint RPC Client

```go
client := NewTendermintClient("http://0.0.0.0:26750")
abci, err := client.TmClient.ABCIInfo(context.Background())
```

There is an option with multiple providers available. Upon interaction with the blockchain, the provider with the highest
block height will be chosen.

```go
gnfdClients := NewGnfdCompositClients(
    []string{test.TEST_GRPC_ADDR, test.TEST_GRPC_ADDR2, test.TEST_GRPC_ADDR3},
    []string{test.TEST_RPC_ADDR, test.TEST_RPC_ADDR2, test.TEST_RPC_ADDR3},
    test.TEST_CHAIN_ID,
    WithGrpcDialOption(grpc.WithTransportCredentials(insecure.NewCredentials())))

client, err := gnfdClients.GetClient()
```