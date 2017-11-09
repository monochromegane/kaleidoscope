package kaleidoscope

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
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
