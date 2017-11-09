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

func (k *Kaleidoscope) set(dbname, root, key, value string) (string, error) {
	hash, err := k.client.Add("value", k.wrapWithMetadata(value),
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

func (k Kaleidoscope) wrapWithMetadata(value string) io.Reader {
	return strings.NewReader(strconv.FormatInt(time.Now().Unix(), 10) + "," + value)
}
