package kaleidoscope

type Kaleidoscope struct {
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
