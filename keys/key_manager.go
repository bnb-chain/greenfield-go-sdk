package keys

import "github.com/bnb-chain/greenfield/sdk/keys"

type (
	KeyManager = keys.KeyManager
)

var (
	NewPrivateKeyManager  = keys.NewPrivateKeyManager
	NewMnemonicKeyManager = keys.NewMnemonicKeyManager
)
