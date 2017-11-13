package kaleidoscope

import (
	"bytes"
	"strconv"
	"time"
)

type Kaleidoscope struct {
	dbname   string
	head     string
	client   Client
	keystore Keystore
}

func New() (Kaleidoscope, error) {
	client, err := NewClient()
	if err != nil {
		return Kaleidoscope{}, err
	}
	return Kaleidoscope{
		client:   client,
		keystore: NewKeyStore(),
	}, nil
}

func (k *Kaleidoscope) Create(dbname string, size int) (string, error) {
	err := k.client.KeyGen(dbname, RequestOptions{
		"type": "rsa",
		"size": strconv.Itoa(size),
	})
	if err != nil {
		return "", err
	}
	err = k.keystore.Load(dbname)
	if err != nil {
		return "", err
	}

	return k.set(dbname, EmptyDirMultiHash, "__database_name", dbname)
}

func (k *Kaleidoscope) Use(dbname string) error {
	// get head

	// load keypair
	err := k.keystore.Load(dbname)
	if err != nil {
		return err
	}
	k.use(dbname, "")
	return nil
}

func (k *Kaleidoscope) Set(key, value string) (string, error) {
	return k.set(k.dbname, k.latest(), key, value)
}

func (k Kaleidoscope) Get(key string) ([]byte, []byte, error) {
	enc, err := k.client.Cat(k.latest()+"/"+key+"/value", RequestOptions{})
	if err != nil {
		return []byte{}, []byte{}, err
	}
	plain, err := k.keystore.Decrypt(enc)
	if err != nil {
		return []byte{}, []byte{}, err
	}
	return plain[0:10], plain[11:], nil
}

func (k *Kaleidoscope) set(dbname, root, key, value string) (string, error) {
	enc, err := k.keystore.EncryptString(wrapWithMetadata(value))
	if err != nil {
		return "", err
	}
	hash, err := k.client.Add("value", bytes.NewReader(enc),
		RequestOptions{"wrap-with-directory": "true"})
	if err != nil {
		return "", err
	}

	dbhash, err := k.client.ObjectPatchAddLink(root, key, hash, RequestOptions{})
	if err != nil {
		return "", err
	}

	k.use(dbname, dbhash)
	return dbhash, err
}

func (k *Kaleidoscope) use(dbname, head string) {
	// TODO: Use mutex
	k.dbname = dbname
	k.head = head
}

func (k Kaleidoscope) latest() string {
	// TODO: Use mutex
	return k.head
}

func wrapWithMetadata(value string) string {
	return strconv.FormatInt(time.Now().Unix(), 10) + "," + value
}

// func wrapWithMetadata(value string) io.Reader {
// 	return strings.NewReader(strconv.FormatInt(time.Now().Unix(), 10) + "," + value)
// }
