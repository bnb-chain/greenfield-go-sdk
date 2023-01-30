package ethsecp256k1

import (
	"encoding/hex"
	"github.com/evmos/ethermint/crypto/ethsecp256k1"
	ethHd "github.com/evmos/ethermint/crypto/hd"
)

func HexToEthSecp256k1PrivKey(hexString string) (*ethsecp256k1.PrivKey, error) {
	bz, err := hex.DecodeString(hexString)
	if err != nil {
		return nil, err
	}
	return ethHd.EthSecp256k1.Generate()(bz).(*ethsecp256k1.PrivKey), nil
}
