package kaleidoscope

import (
	"bytes"
	"encoding/json"
	"strconv"
	"strings"
	"time"
)

type Kaleidoscope struct {
	dbname   string
	head     string
	client   Client
	keystore Keystore
	stream   Stream
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
	err := k.keystore.Load(dbname)
	if err != nil {
		return err
	}
	ipns, err := k.keystore.PeerID()
	if err != nil {
		return err
	}
	head, err := k.client.NameResolve(ipns, RequestOptions{"nocache": "true"})
	if err != nil {
		return err
	}
	err = k.keystore.Load(dbname)
	if err != nil {
		return err
	}
	k.use(dbname, head)
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

func (k Kaleidoscope) Save() error {
	_, _, err := k.client.NamePublish(k.head, RequestOptions{"key": k.dbname})
	return err
}

func (k *Kaleidoscope) StartSync() error {
	stream, err := k.client.PubSubSub(k.dbname, RequestOptions{"discover": "true"})
	if err != nil {
		return err
	}
	k.stream = stream
	go func() {
		for data := range k.stream.Data {
			var ope Operation
			err := json.NewDecoder(strings.NewReader(data)).Decode(&ope)
			if err != nil {
				continue
			}
			if ope.Database != k.dbname {
				continue
			}
			if strings.ToLower(ope.Type) == "set" {
				k.setHash(k.dbname, k.latest(), ope.Key, ope.Hash, false)
			}
		}
	}()
	return nil
}

func (k *Kaleidoscope) StopSync() {
	k.stream.Close()
}

type Operation struct {
	Type     string
	Database string
	Key      string
	Hash     string
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
	return k.setHash(dbname, root, key, hash, true)
}

func (k *Kaleidoscope) setHash(dbname, root, key, hash string, pub bool) (string, error) {
	dbhash, err := k.client.ObjectPatchAddLink(root, key, hash, RequestOptions{})
	if err != nil {
		return "", err
	}

	k.use(dbname, dbhash)
	if pub && k.stream.IsRunning() {
		ope := Operation{
			Type:     "set",
			Database: k.dbname,
			Key:      key,
			Hash:     hash,
		}
		json, err := json.Marshal(ope)
		if err != nil {
			return "", err
		}
		k.client.PubSubPub(k.dbname, string(json), RequestOptions{})
	}
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
