package keys

import (
	"encoding/hex"
	"encoding/json"
	"fmt"
	keyring99 "github.com/99designs/keyring"
	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/crypto/keyring"
	"github.com/cosmos/cosmos-sdk/crypto/keys/secp256k1"
	"io"
	"io/ioutil"

	"github.com/cosmos/cosmos-sdk/types"
	"github.com/tendermint/tendermint/crypto"
)

type KeyManager interface {
	Sign(msg tx.StdSignMsg) ([]byte, error)
	GetPrivKey() crypto.PrivKey
	GetAddr() ctypes.AccAddress
}

type keyManager struct {
	privKey  crypto.PrivKey
	addr     types.AccAddress
	mnemonic string
}

func NewKeyStoreKeyManager(file string, auth string) (KeyManager, error) {
	k := keyManager{}
	err := k.recoveryFromKeyStore(file, auth)
	return &k, err
}

func NewPrivateKeyManager(priKey string) (KeyManager, error) {
	k := keyManager{}
	err := k.recoveryFromPrivateKey(priKey)
	return &k, err
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
	priKey := secp256k1.PrivKeySecp256k1(keyBytesArray)
	addr := ctypes.AccAddress(priKey.PubKey().Address())
	m.addr = addr
	m.privKey = priKey
	return nil
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
	priKey := secp256k1.PrivKey(keyBytesArray)
	addr := ctypes.AccAddress(priKey.PubKey().Address())
	m.addr = addr
	m.privKey = priKey
	return nil
}

func GenPrivateKeyFromSecret(secret []byte) *secp256k1.PrivKey {
	return secp256k1.GenPrivKeyFromSecret(secret)
}

func NewInMemory(cdc codec.Codec, opts ...keyring.Option) keyring.Keyring {
	return keyring.NewInMemory(cdc, opts...)
}

func NewInMemoryWithKeyring(kr keyring99.Keyring, cdc codec.Codec, opts ...keyring.Option) keyring.Keyring {
	return keyring.NewInMemoryWithKeyring(kr, cdc, opts...)
}

func NewKeyring(appName, backend, rootDir string, userInput io.Reader, cdc codec.Codec, opts ...keyring.Option) (keyring.Keyring, error) {
	return keyring.New(appName, backend, rootDir, userInput, cdc, opts...)
}

//
//func (m *keyManager) Sign(msg tx.StdSignMsg) ([]byte, error) {
//	sig, err := m.makeSignature(msg)
//	if err != nil {
//		return nil, err
//	}
//	newTx := tx.NewStdTx(msg.Msgs, []tx.StdSignature{sig}, msg.Memo, msg.Source, msg.Data)
//	bz, err := tx.Cdc.MarshalBinaryLengthPrefixed(&newTx)
//	if err != nil {
//		return nil, err
//	}
//	return bz, nil
//}

func (m *keyManager) GetPrivKey() crypto.PrivKey {
	return m.privKey
}

func (m *keyManager) GetAddr() ctypes.AccAddress {
	return m.addr
}

//
//func (m *keyManager) makeSignature(msg tx.StdSignMsg) (sig tx.StdSignature, err error) {
//	if err != nil {
//		return
//	}
//	sigBytes, err := m.privKey.Sign(msg.Bytes())
//	if err != nil {
//		return
//	}
//	return tx.StdSignature{
//		AccountNumber: msg.AccountNumber,
//		Sequence:      msg.Sequence,
//		PubKey:        m.privKey.PubKey(),
//		Signature:     sigBytes,
//	}, nil
//}

//
//func GetInscriptionPrivateKey(cfg *config.InscriptionConfig) *ethsecp256k1.PrivKey {
//	var privateKey string
//	if cfg.KeyType == config.KeyTypeAWSPrivateKey {
//		result, err := config.GetSecret(cfg.AWSSecretName, cfg.AWSRegion)
//		if err != nil {
//			panic(err)
//		}
//		type AwsPrivateKey struct {
//			PrivateKey string `json:"private_key"`
//		}
//		var awsPrivateKey AwsPrivateKey
//		err = json.Unmarshal([]byte(result), &awsPrivateKey)
//		if err != nil {
//			panic(err)
//		}
//		privateKey = awsPrivateKey.PrivateKey
//	} else {
//		privateKey = cfg.PrivateKey
//	}
//	privKey, err := HexToEthSecp256k1PrivKey(privateKey)
//	if err != nil {
//		panic(err)
//	}
//	return privKey
//}
