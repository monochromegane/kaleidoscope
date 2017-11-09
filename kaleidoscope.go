package kaleidoscope

import (
	"io"
	"strconv"
	"strings"
	"time"
)

type Kaleidoscope struct {
	dbname string
	head   string
	client Client
}

func New() (Kaleidoscope, error) {
	client, err := NewClient()
	if err != nil {
		return Kaleidoscope{}, err
	}

	return Kaleidoscope{
		client: client,
	}, nil
}

func (k *Kaleidoscope) Create(dbname string) (string, error) {
	return k.set(dbname, EmptyDirMultiHash, "__database_name", dbname)
}

func (k *Kaleidoscope) Set(key, value string) (string, error) {
	return k.set(k.dbname, k.latest(), key, value)
}

func (k Kaleidoscope) Get(key string) ([]byte, []byte, error) {
	value, err := k.client.Cat(k.latest()+"/"+key+"/value", RequestOptions{})
	if err != nil {
		return []byte{}, []byte{}, err
	}
	return value[0:10], value[11:], nil
}

func (k *Kaleidoscope) set(dbname, root, key, value string) (string, error) {
	hash, err := k.client.Add("value", wrapWithMetadata(value),
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

func wrapWithMetadata(value string) io.Reader {
	return strings.NewReader(strconv.FormatInt(time.Now().Unix(), 10) + "," + value)
}
