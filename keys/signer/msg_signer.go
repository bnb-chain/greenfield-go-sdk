package keys

import (
	"github.com/bnb-chain/greenfield-go-sdk/keys"
	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	ethcrypto "github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/crypto/secp256k1"
	"github.com/evmos/ethermint/crypto/ethsecp256k1"
)

// MsgSigner It defines a type to sign a message in the same way as MsgEthereumTx
type MsgSigner struct {
	keyManager keys.KeyManager
}

func NewMsgSigner(key keys.KeyManager) *MsgSigner {
	return &MsgSigner{
		keyManager: key,
	}
}

// Sign signs the message using the underlying private key
func (m MsgSigner) Sign(msg []byte) ([]byte, cryptotypes.PubKey, error) {
	sig, err := m.keyManager.Sign(msg)
	if err != nil {
		return nil, nil, err
	}

	return sig, m.keyManager.PubKey(), nil
}

// RecoverAddr recovers the sender address from msg and signature
func RecoverAddr(msg []byte, sig []byte) (sdk.AccAddress, ethsecp256k1.PubKey, error) {
	pubKeyByte, err := secp256k1.RecoverPubkey(msg, sig)
	if err != nil {
		return nil, ethsecp256k1.PubKey{}, err
	}
	pubKey, _ := ethcrypto.UnmarshalPubkey(pubKeyByte)
	pk := ethsecp256k1.PubKey{
		Key: ethcrypto.CompressPubkey(pubKey),
	}

	recoverAcc := sdk.AccAddress(pk.Address().Bytes())

	return recoverAcc, pk, nil
}
