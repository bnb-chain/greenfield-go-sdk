package keys

import (
	"fmt"
	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/crypto"
	"github.com/cosmos/cosmos-sdk/crypto/hd"
	"github.com/cosmos/cosmos-sdk/crypto/keyring"
	"github.com/cosmos/cosmos-sdk/crypto/types"
)

type keystore struct {
	db      keyring.Keyring
	cdc     codec.Codec
	backend string
	options keyring.Options
}

func newKeystore(kr keyring.Keyring, cdc codec.Codec, backend string, opts ...keyring.Option) keystore {
	// Default options for keybase, these can be overwritten using the
	// Option function
	options := keyring.Options{
		SupportedAlgos:       keyring.SigningAlgoList{hd.Secp256k1},
		SupportedAlgosLedger: keyring.SigningAlgoList{hd.Secp256k1},
	}

	for _, optionFn := range opts {
		optionFn(&options)
	}

	return keystore{
		db:      kr,
		cdc:     cdc,
		backend: backend,
		options: options,
	}
}

func (ks keystore) ImportPrivKey(uid, armor, passphrase string) error {
	if k, err := ks.Key(uid); err == nil {
		if uid == k.Name {
			return fmt.Errorf("cannot overwrite key: %s", uid)
		}
	}

	privKey, _, err := crypto.UnarmorDecryptPrivKey(armor, passphrase)
	if err != nil {
		return errors.Wrap(err, "failed to decrypt private key")
	}

	_, err = ks.writeLocalKey(uid, privKey)
	if err != nil {
		return err
	}

	return nil
}

func (ks keystore) ImportPubKey(uid string, armor string) error {
	if _, err := ks.Key(uid); err == nil {
		return fmt.Errorf("cannot overwrite key: %s", uid)
	}

	pubBytes, _, err := crypto.UnarmorPubKeyBytes(armor)
	if err != nil {
		return err
	}

	var pubKey types.PubKey
	if err := ks.cdc.UnmarshalInterface(pubBytes, &pubKey); err != nil {
		return err
	}

	_, err = ks.writeOfflineKey(uid, pubKey)
	if err != nil {
		return err
	}

	return nil
}

func (ks keystore) Sign(uid string, msg []byte) ([]byte, types.PubKey, error) {
	k, err := ks.Key(uid)
	if err != nil {
		return nil, nil, err
	}

	switch {
	case k.GetLocal() != nil:
		priv, err := extractPrivKeyFromLocal(k.GetLocal())
		if err != nil {
			return nil, nil, err
		}

		sig, err := priv.Sign(msg)
		if err != nil {
			return nil, nil, err
		}

		return sig, priv.PubKey(), nil

	case k.GetLedger() != nil:
		return SignWithLedger(k, msg)

		// multi or offline record
	default:
		pub, err := k.GetPubKey()
		if err != nil {
			return nil, nil, err
		}

		return nil, pub, errors.New("cannot sign with offline keys")
	}
}

func (ks keystore) SignByAddress(address sdk.Address, msg []byte) ([]byte, types.PubKey, error) {
	k, err := ks.KeyByAddress(address)
	if err != nil {
		return nil, nil, err
	}

	return ks.Sign(k.Name, msg)
}
