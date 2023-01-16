package inscription_go_sdk

import (
	"github.com/cosmos/cosmos-sdk/crypto/keyring"
)

func main() {
	haha := keyring.BackendFile
	println(haha)
}
