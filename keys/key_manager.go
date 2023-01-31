package keys

import (
	"encoding/hex"
	"encoding/json"
	"fmt"
	"github.com/cosmos/cosmos-sdk/crypto/hd"
	ctypes "github.com/cosmos/cosmos-sdk/crypto/types"
	"github.com/evmos/ethermint/crypto/ethsecp256k1"
	ethHd "github.com/evmos/ethermint/crypto/hd"

	"io/ioutil"
	"strings"

	"github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/go-bip39"
)

const (
	defaultBIP39Passphrase = ""
	FullPath               = "m/44'/60'/0'/0/0"
)

type KeyManager interface {
	Sign(b []byte) ([]byte, error)
	GetPrivKey() ctypes.PrivKey
	GetAddr() types.AccAddress
}

type keyManager struct {
	privKey  ctypes.PrivKey
	addr     types.AccAddress
	mnemonic string
}

func NewPrivateKeyManager(priKey string) (KeyManager, error) {
	k := keyManager{}
	err := k.recoveryFromPrivateKey(priKey)
	return &k, err
}

func NewMnemonicKeyManager(mnemonic string) (KeyManager, error) {
	k := keyManager{}
	err := k.recoveryFromMnemonic(mnemonic, FullPath)
	return &k, err
}

func NewKeyStoreKeyManager(file string, auth string) (KeyManager, error) {
	k := keyManager{}
	err := k.recoveryFromKeyStore(file, auth)
	return &k, err
}

func (m *keyManager) recoveryFromPrivateKey(privateKey string) error {
	priBytes, err := hex.DecodeString(privateKey)
	if err != nil {
		return err
	}

	if len(priBytes) != 32 {
		return fmt.Errorf("Len of Keybytes is not equal to 32 ")
	}
	var keyBytesArray [32]byte
	copy(keyBytesArray[:], priBytes[:32])
	priKey := ethHd.EthSecp256k1.Generate()(keyBytesArray[:]).(*ethsecp256k1.PrivKey)
	addr := types.AccAddress(priKey.PubKey().Address())
	m.addr = addr
	m.privKey = priKey
	return nil
}

func (m *keyManager) recoveryFromMnemonic(mnemonic, keyPath string) error {
	words := strings.Split(mnemonic, " ")
	if len(words) != 12 && len(words) != 24 {
		return fmt.Errorf("mnemonic length should either be 12 or 24")
	}
	seed, err := bip39.NewSeedWithErrorChecking(mnemonic, defaultBIP39Passphrase)
	if err != nil {
		return err
	}
	// create master key and derive first key:
	masterPriv, ch := hd.ComputeMastersFromSeed(seed)
	derivedPriv, err := hd.DerivePrivateKeyForPath(masterPriv, ch, keyPath)
	if err != nil {
		return err
	}
	priKey := ethHd.EthSecp256k1.Generate()(derivedPriv[:]).(*ethsecp256k1.PrivKey)
	addr := types.AccAddress(priKey.PubKey().Address())
	if err != nil {
		return err
	}
	m.addr = addr
	m.privKey = priKey
	m.mnemonic = mnemonic
	return nil
}

func (m *keyManager) recoveryFromKeyStore(keystoreFile string, auth string) error {
	if auth == "" {
		return fmt.Errorf("Password is missing ")
	}
	keyJson, err := ioutil.ReadFile(keystoreFile)
	if err != nil {
		return err
	}
	var encryptedKey EncryptedKeyJSON
	err = json.Unmarshal(keyJson, &encryptedKey)
	if err != nil {
		return err
	}
	keyBytes, err := decryptKey(&encryptedKey, auth)
	if err != nil {
		return err
	}
	if len(keyBytes) != 32 {
		return fmt.Errorf("Len of Keybytes is not equal to 32 ")
	}
	var keyBytesArray [32]byte
	copy(keyBytesArray[:], keyBytes[:32])
	priKey := ethHd.EthSecp256k1.Generate()(keyBytesArray[:]).(*ethsecp256k1.PrivKey)
	addr := types.AccAddress(priKey.PubKey().Address())
	m.addr = addr
	m.privKey = priKey
	return nil
}

func (m *keyManager) Sign(signBytes []byte) ([]byte, error) {
	return m.privKey.Sign(signBytes)
}

func (m *keyManager) GetPrivKey() ctypes.PrivKey {
	return m.privKey
}

func (m *keyManager) GetAddr() types.AccAddress {
	return m.addr
}
