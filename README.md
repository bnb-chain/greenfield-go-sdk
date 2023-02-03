# Greenfield Go SDK

The `Greenfield-GO-SDK` provides a thin wrapper for interacting with `greenfield` in two ways:

1. Interact using `gnfd-tendermint` RPC client, you may perform low-level operations like executing ABCI queries, viewing network/consensus state.
2. Interact using `gnfd-cosmos-sdk` GRPC clients, this includes querying accounts, chain info and broadcasting transaction.

### Requirement

Go version above 1.19

## Usage

### Importing

```go
import (
    "github.com/bnb-chain/gnfd-go-sdk" latest
)
```

### Key Manager

Key Manager is needed to sign the transaction msg or verify signature. Key Manager is an Identity Manager to define who
you are in the greenfield. It provides following interface:

```go
type KeyManager interface {

    Sign(signByte []byte) ([]byte, error)
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

### Use GRPC Client

#### Init client without key manager, you should use it for only querying purpose.

```go
client := NewGreenfieldClient("localhost:9090", "greenfield_9000-121")
```

#### Init client with key manager, for signing and sending tx

```go
keyManager, _ := keys.NewPrivateKeyManager("ab463aca3d2965233da3d1d6108aa521274c5ddc2369ff72970a52a451863fbf")
client := NewGreenfieldClientWithKeyManager("localhost:9090", "greenfield_9000-121", keyManager)
```

#### Broadcast TX

A generic method `BroadcastTx` is provided to give you the flexibility to broadcast different types of transaction.
```go
BroadcastTx(msgs []sdk.Msg, txOpt *types.TxOption, opts ...grpc.CallOption) (*tx.BroadcastTxResponse, error)
```

`txOpt` is provided to customize your transaction. It is optional, and all fields are optional.
```go
type TxOption struct {
    Async     bool   // default to `sync` mode
    GasLimit  uint64 // default to use simulated gas 
    Memo      string
    FeeAmount sdk.Coins
    FeePayer  sdk.AccAddress
}
```
Example:

```go
payerAddr, _ := sdk.AccAddressFromHexUnsafe("0x76d244CE05c3De4BbC6fDd7F56379B145709ade9")
txOpt := &types.TxOption{
    Async:     true,
    GasLimit:  1000000,
    Memo:      "test",
    FeeAmount: sdk.Coins{{"bnb", sdk.NewInt(1)}},
    FeePayer:  payerAddr,
}
response, _ := gnfdCli.BroadcastTx([]sdk.Msg{transfer}, txOpt)
```
before using it, you need to construct the appropriate type of `Msg`, refer to `gnfd-cosmos-sdk` for msg types supported