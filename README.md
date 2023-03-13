# Greenfield Go SDK

The `Greenfield-GO-SDK` provides a thin wrapper for interacting with `greenfield` in three ways:

1. Interact using `GreenfieldClient` client, you may perform querying accounts, chain info and broadcasting transaction.
2. Interact using `TendermintClient` client, you may perform low-level operations like executing ABCI queries, viewing network/consensus state.
3. Interact using `SPClient` client, you may perform  request for the service of storage provider like putObject, getObject

### Requirement

Go version above 1.19

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
    github.com/cosmos/cosmos-sdk => github.com/bnb-chain/greenfield-cosmos-sdk v0.0.9
    github.com/gogo/protobuf => github.com/regen-network/protobuf v1.3.3-alpha.regen.1
    github.com/tendermint/tendermint => github.com/bnb-chain/greenfield-tendermint v0.0.2
)
```

### Key Manager

Key Manager is needed to sign the transaction msg or verify signature. Key Manager is an Identity Manager to define who
you are in the greenfield. It provides following interface:

```go
type KeyManager interface {
    GetPrivKey() ctypes.PrivKey
    GetAddr() types.AccAddress
}
```

We provide three construct functions to generate the Key Manager:
```go
NewPrivateKeyManager(priKey string) (KeyManager, error)

NewMnemonicKeyManager(mnemonic string) (KeyManager, error)
```

- NewPrivateKeyManager. You should provide a Hex encoded string of your private key.
- NewMnemonicKeyManager. You should provide your mnemonic, usually is a string of 24 words.

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

#### Init client without key manager, you should use it for only querying purpose.

```go
client := NewGreenfieldClient("localhost:9090", "greenfield_9000-121")

query := banktypes.QueryBalanceRequest{
		Address: "0x76d244CE05c3De4BbC6fDd7F56379B145709ade9",
		Denom:   "BNB",
}
res, err := client.BankQueryClient.Balance(context.Background(), &query)  
```

#### Init client with key manager, for signing and sending tx

```go
keyManager, _ := keys.NewPrivateKeyManager("ab463aca3d2965233da3d1d6108aa521274c5ddc2369ff72970a52a451863fbf")
gnfdClient := NewGreenfieldClient("localhost:9090", 
	                            "greenfield_9000-121",
	                            WithKeyManager(km),
                                    WithGrpcDialOption(grpc.WithTransportCredentials(insecure.NewCredentials()))
)
```

#### Broadcast TX

A generic method `BroadcastTx` is provided to give you the flexibility to broadcast different types of transaction.
```go
BroadcastTx(msgs []sdk.Msg, txOpt *types.TxOption, opts ...grpc.CallOption) (*tx.BroadcastTxResponse, error)
```

`txOpt` is provided to customize your transaction. It is optional, and all fields are optional.
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

`SignTx` is provided which sign the `msgs` and returns raw bytes 

```go
SignTx(msgs []sdk.Msg, txOpt *types.TxOption) ([]byte, error)
```

### Get Nonce

Get the nonce of account
```go
GetNonce() (uint64, error)
```

#### Support transaction type
Please refer to [msgTypes.go](./types/msgTypes.go) to get all types of `sdk.Msg` supported 


### Use Tendermint RPC Client

```go
client := NewTendermintClient("http://0.0.0.0:26750")
abci, err := client.TmClient.ABCIInfo(context.Background())
```

There is an option which multiple providers are available, and by the time you interact with Blockchain, it will choose the
provider with the highest block height

```go
gnfdClients := NewGnfdCompositClients(
    []string{test.TEST_GRPC_ADDR, test.TEST_GRPC_ADDR2, test.TEST_GRPC_ADDR3},
    []string{test.TEST_RPC_ADDR, test.TEST_RPC_ADDR2, test.TEST_RPC_ADDR3},
    test.TEST_CHAIN_ID,
    WithGrpcDialOption(grpc.WithTransportCredentials(insecure.NewCredentials())))

client, err := gnfdClients.GetClient()
```

### Use Storage Provider Client

#### Auth Mechanism

SPclient support two auto type. The first type is to use the local signer.Sign method, which will call 
the private key of keyManager to sign the request message. The second type is to use the metamask wallet
to generate an authentication token

For the first type, you need to specify the SignType of AuthInfo as AuthV1 using the NewAuthInfo function, 
and pass it as a parameter to the API.

```
authInfo := NewAuthInfo(false, "")
err = client.GetApproval(context.Background(), bucketName, "", authInfo)
```

For the first type, you only need get the metaMask sign Token and specify the SignType of AuthInfo as AuthV2.
The metaMask sign Token can be designs like JWT token with expiration in it. It needs to be implemented externally
and pass it as a parameter to the API.

```
authInfo := NewAuthInfo(true, "this is metamask auto token")
err = client.GetApproval(context.Background(), bucketName, "", authInfo)
```

#### Init client

if SPclient use the AuthV1, the client need to init with key manager
```
// if client keep the private key in keyManager locally
client := NewSpClientWithKeyManager("http://0.0.0.0:26750", &spClient.Option{}, keyManager)
```
if SPclient use the AuthV2, the client can init without key manager
```
// If the client does not manage the private key locally and use local
client := NewSpClient("http://0.0.0.0:26750", &spClient.Option{})
```

#### call API and send request to storage provider

```go
fileReader, err := os.Open(filePath)

meta := spClient.ObjectMeta{
    ObjectSize:  length,
    ContentType: "application/octet-stream",
}

err = client.PutObject(ctx, bucketName, ObjectName, txnHash, fileReader, meta, NewAuthInfo(false, "")))
```