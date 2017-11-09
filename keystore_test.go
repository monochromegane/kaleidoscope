package kaleidoscope

import (
	"strings"
	"testing"

	ci "github.com/libp2p/go-libp2p-crypto"
)

func TestKeystoreEncryptAndDecrypt(t *testing.T) {
	keystore := testKeystore()
	expect := "Some value"

	enc, err := keystore.EncryptString(expect)
	if err != nil {
		t.Errorf("EncryptString should not return error, but %s", err)
	}
	if string(enc) == expect {
		t.Errorf("EncryptString should return encrypted value, but %s", string(enc))
	}

	plain, err := keystore.Decrypt(enc)
	if err != nil {
		t.Errorf("Decrypt should not return error, but %s", err)
	}
	if string(plain) != expect {
		t.Errorf("Decrypt should return decrypted value (%s), but %s", expect, string(plain))
	}
}

func TestKeystorePeerID(t *testing.T) {
	keystore := testKeystore()

	id, err := keystore.PeerID()
	if err != nil {
		t.Errorf("PeerID should not return error, but %s", err)
	}
	if !strings.HasPrefix(id, "Qm") {
		t.Errorf("PeerID should return multihash, but %s", id)
	}
}

func testKeystore() Keystore {
	priv, _, _ := ci.GenerateKeyPair(ci.RSA, 2048)
	return Keystore{priv: priv}
}
