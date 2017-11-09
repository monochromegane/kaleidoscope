package kaleidoscope

import (
	"io/ioutil"
	"net/http"
	"os"
	"path"
	"strings"

	homedir "github.com/mitchellh/go-homedir"
	ma "github.com/multiformats/go-multiaddr"
	manet "github.com/multiformats/go-multiaddr-net"
)

const (
	DefaultPathName     = ".ipfs"
	DefaultPathRoot     = "~/" + DefaultPathName
	DefaultApiFile      = "api"
	DefaultKeystoreRoot = "keystore"
	EnvDir              = "IPFS_PATH"
	EmptyDirMultiHash   = "QmUNLLsPACCz1vLxQVkXqqLX5R1X345qqfHbsf67hvA3Nn"
)

type IPFS struct {
	url    string
	client http.Client
}

func NewLocalIPFS() (IPFS, error) {
	baseDir := os.Getenv(EnvDir)
	if baseDir == "" {
		baseDir = DefaultPathRoot
	}

	baseDir, err := homedir.Expand(baseDir)
	if err != nil {
		return IPFS{}, err
	}

	apiFile := path.Join(baseDir, DefaultApiFile)

	if _, err := os.Stat(apiFile); err != nil {
		return IPFS{}, err
	}

	api, err := ioutil.ReadFile(apiFile)
	if err != nil {
		return IPFS{}, err
	}

	return NewIPFS(strings.TrimSpace(string(api)))
}

func NewIPFS(url string) (IPFS, error) {
	if a, err := ma.NewMultiaddr(url); err == nil {
		_, host, err := manet.DialArgs(a)
		if err == nil {
			url = host
		}
	}
	c := http.Client{
		Transport: &http.Transport{
			DisableKeepAlives: true,
		},
	}

	return IPFS{
		url:    url,
		client: c,
	}, nil
}
