package kaleidoscope

type Client struct {
	ipfs IPFS
}

func NewClient() (Client, error) {
	ipfs, err := NewLocalIPFS()
	if err != nil {
		return Client{}, err
	}
	return Client{
		ipfs: ipfs,
	}, nil
}
