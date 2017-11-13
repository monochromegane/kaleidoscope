package kaleidoscope

import (
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"path/filepath"

	ci "github.com/libp2p/go-libp2p-crypto"
	peer "github.com/libp2p/go-libp2p-peer"
	homedir "github.com/mitchellh/go-homedir"
)

type Keystore struct {
	priv        ci.PrivKey
	keypair     string
	persistence bool
}

func NewKeyStore() Keystore {
	return Keystore{persistence: true}
}

func (k *Keystore) Load(keypair string) error {
	if k.keypair == keypair {
		return nil
	}
	if k.persistence {
		return k.loadFromFile(keypair)
	} else {
		// Use only for testing.
		priv, _, err := ci.GenerateKeyPair(ci.RSA, 2048)
		if err != nil {
			return err
		}
		k.priv = priv
		k.keypair = keypair
	}
	return nil
}

func (k Keystore) EncryptString(plain string) ([]byte, error) {
	return k.Encrypt([]byte(plain))
}

func (k Keystore) Encrypt(plain []byte) ([]byte, error) {
	pub, _, err := k.key()
	if err != nil {
		return []byte{}, err
	}
	return pub.Encrypt(plain)
}

func (k Keystore) Decrypt(enc []byte) ([]byte, error) {
	_, priv, err := k.key()
	if err != nil {
		return []byte{}, err
	}
	return priv.Decrypt(enc)
}

func (k Keystore) PeerID() (string, error) {
	id, err := peer.IDFromPublicKey(k.priv.GetPublic())
	if err != nil {
		return "", err
	}
	return id.Pretty(), err
}

func (k Keystore) key() (*ci.RsaPublicKey, *ci.RsaPrivateKey, error) {
	rsaPriv, ok := k.priv.(*ci.RsaPrivateKey)
	if !ok {
		return nil, nil, fmt.Errorf("Invalid key type. Supports only RSA.")
	}
	rsaPub := rsaPriv.GetPublic().(*ci.RsaPublicKey)
	return rsaPub, rsaPriv, nil
}

func (k *Keystore) loadFromFile(keypair string) error {
	baseDir := os.Getenv(EnvDir)
	if baseDir == "" {
		baseDir = DefaultPathRoot
	}

	baseDir, err := homedir.Expand(baseDir)
	if err != nil {
		return err
	}

	keystore := path.Join(baseDir, DefaultKeystoreRoot)
	data, err := ioutil.ReadFile(filepath.Join(keystore, keypair))
	if err != nil {
		return err
	}

	priv, err := ci.UnmarshalPrivateKey(data)
	if err != nil {
		return err
	}

	k.priv = priv
	k.keypair = keypair
	return nil
}

// func NewKeyStore(keypair string) (Keystore, error) {
// 	baseDir := os.Getenv(EnvDir)
// 	if baseDir == "" {
// 		baseDir = DefaultPathRoot
// 	}
//
// 	baseDir, err := homedir.Expand(baseDir)
// 	if err != nil {
// 		return Keystore{}, err
// 	}
//
// 	keystore := path.Join(baseDir, DefaultKeystoreRoot)
// 	data, err := ioutil.ReadFile(filepath.Join(keystore, keypair))
// 	if err != nil {
// 		return Keystore{}, err
// 	}
//
// 	priv, err := ci.UnmarshalPrivateKey(data)
// 	if err != nil {
// 		return Keystore{}, err
// 	}
//
// 	return Keystore{
// 		priv: priv,
// 	}, nil
// }
