package kaleidoscope

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"mime/multipart"
	"net/textproto"
	"net/url"
)

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

type Object struct {
	Hash string
}

type IPNS struct {
	Name  string
	Value string
	Path  string
}

type Message struct {
	Data string
}

func (c Client) Add(name string, r io.Reader, opts RequestOptions) (string, error) {
	mp, contentType, err := multiPartFromReader(name, r)
	if err != nil {
		return "", err
	}

	req := NewRequest(c.ipfs.url, "add", opts)
	req.Body = &mp
	req.Headers["Content-Type"] = contentType

	resp, err := req.Send(c.ipfs.client)
	if err != nil {
		return "", err
	}
	defer resp.Close()

	if resp.Error != nil {
		return "", resp.Error
	}

	dec := json.NewDecoder(resp.Output)
	var final string
	for {
		var out Object
		err = dec.Decode(&out)
		if err != nil {
			if err == io.EOF {
				break
			}
			return "", err
		}
		final = out.Hash
	}

	if final == "" {
		return "", errors.New("no results received")
	}

	return final, nil
}

func (c Client) ObjectPatchAddLink(root, name, ref string, opts RequestOptions) (string, error) {
	req := NewRequest(c.ipfs.url, "object/patch/add-link", opts, root, name, ref)
	resp, err := req.Send(c.ipfs.client)
	if err != nil {
		return "", err
	}
	defer resp.Close()

	if resp.Error != nil {
		return "", resp.Error
	}

	var out Object
	err = json.NewDecoder(resp.Output).Decode(&out)
	if err != nil {
		return "", err
	}

	return out.Hash, nil
}

func (c Client) ObjectPatchRmLink(root, name string, opts RequestOptions) (string, error) {
	req := NewRequest(c.ipfs.url, "object/patch/rm-link", opts, root, name)
	resp, err := req.Send(c.ipfs.client)
	if err != nil {
		return "", err
	}
	defer resp.Close()

	if resp.Error != nil {
		return "", resp.Error
	}

	var out Object
	err = json.NewDecoder(resp.Output).Decode(&out)
	if err != nil {
		return "", err
	}

	return out.Hash, nil
}

func (c Client) Cat(path string, opts RequestOptions) ([]byte, error) {
	req := NewRequest(c.ipfs.url, "cat", opts, path)
	resp, err := req.Send(c.ipfs.client)
	if err != nil {
		return []byte{}, err
	}
	defer resp.Close()

	if resp.Error != nil {
		return []byte{}, resp.Error
	}

	return ioutil.ReadAll(resp.Output)
}

func (c Client) KeyGen(name string, opts RequestOptions) error {
	req := NewRequest(c.ipfs.url, "key/gen", opts, name)
	resp, err := req.Send(c.ipfs.client)
	if err != nil {
		return err
	}
	defer resp.Close()

	if resp.Error != nil {
		return resp.Error
	}
	return nil
}

func (c Client) NamePublish(hash string, opts RequestOptions) (string, string, error) {
	req := NewRequest(c.ipfs.url, "name/publish", opts, hash)
	resp, err := req.Send(c.ipfs.client)
	if err != nil {
		return "", "", err
	}
	defer resp.Close()

	if resp.Error != nil {
		return "", "", err
	}

	var out IPNS
	err = json.NewDecoder(resp.Output).Decode(&out)
	if err != nil {
		return "", "", err
	}
	return out.Name, out.Value, nil
}

func (c Client) NameResolve(name string, opts RequestOptions) (string, error) {
	req := NewRequest(c.ipfs.url, "name/resolve", opts, name)
	resp, err := req.Send(c.ipfs.client)
	if err != nil {
		return "", err
	}
	defer resp.Close()

	if resp.Error != nil {
		return "", err
	}

	var out IPNS
	err = json.NewDecoder(resp.Output).Decode(&out)
	if err != nil {
		return "", err
	}
	return out.Path, nil
}

func (c Client) PubSubPub(topic, payload string, opts RequestOptions) error {
	req := NewRequest(c.ipfs.url, "pubsub/pub", opts, topic, payload)
	resp, err := req.Send(c.ipfs.client)
	if err != nil {
		return err
	}
	defer resp.Close()

	if resp.Error != nil {
		return err
	}
	return nil
}

func (c Client) PubSubSub(topic string, opts RequestOptions) (Stream, error) {
	req := NewRequest(c.ipfs.url, "pubsub/sub", opts, topic)
	resp, err := req.Send(c.ipfs.client)
	if err != nil {
		return Stream{}, err
	}

	if resp.Error != nil {
		return Stream{}, err
	}

	stream := Stream{
		Data: make(chan string),
		src:  resp.Output,
	}
	go stream.read()
	return stream, nil
}

func multiPartFromReader(name string, r io.Reader) (bytes.Buffer, string, error) {
	var b bytes.Buffer
	w := multipart.NewWriter(&b)

	h := make(textproto.MIMEHeader)
	h.Set("Content-Disposition",
		fmt.Sprintf(`form-data; name="file"; filename="%s"`, url.QueryEscape(name)))
	h.Set("Content-Type", "application/octet-stream")

	pw, err := w.CreatePart(h)
	if err != nil {
		return b, "", err
	}

	_, err = io.Copy(pw, r)
	if err != nil {
		return b, "", err
	}

	w.Close()

	return b, w.FormDataContentType(), nil
}
