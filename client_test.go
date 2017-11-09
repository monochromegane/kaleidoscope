package kaleidoscope

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestClientAdd(t *testing.T) {
	expect := "QmSomeObjectHash"
	res := `{"Name":"file.txt","Hash":"%s","Size":"13"}`

	ipfs := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintln(w, fmt.Sprintf(res, expect))
	}))
	defer ipfs.Close()

	client := testClient(ipfs.URL)
	hash, err := client.Add("file.txt", strings.NewReader("some value"), RequestOptions{})

	if err != nil {
		t.Errorf("Add should not return error, but %s", err)
	}
	if hash != expect {
		t.Errorf("Add should return file's hash (%s), but %s", expect, hash)
	}
}

func TestClientAddWrapWithDirectory(t *testing.T) {
	expect := "QmSomeLinkHash"
	res := `{"Name":"file.txt","Hash":"QmSomeObjectHash","Size":"13"}
{"Name":"","Hash":"%s","Size":"67"}`

	ipfs := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintln(w, fmt.Sprintf(res, expect))
	}))
	defer ipfs.Close()

	client := testClient(ipfs.URL)
	hash, err := client.Add("file.txt", strings.NewReader("some value"),
		RequestOptions{"wrap-with-directory": "true"})

	if err != nil {
		t.Errorf("Add should not return error, but %s", err)
	}
	if hash != expect {
		t.Errorf("Add should return wrapped directory's hash (%s), but %s", expect, hash)
	}
}

func TestClientObjectPatchAddLink(t *testing.T) {
	expect := "QmSomeLinkHash"
	res := `{"Hash":"%s"}`

	ipfs := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintln(w, fmt.Sprintf(res, expect))
	}))
	defer ipfs.Close()

	client := testClient(ipfs.URL)
	hash, err := client.ObjectPatchAddLink("QmSomeRootHash", "new_file.txt", "QmSomeObjectHash", RequestOptions{})

	if err != nil {
		t.Errorf("ObjectPatchAddLink should not return error, but %s", err)
	}
	if hash != expect {
		t.Errorf("ObjectPatchAddLink should return new link's hash (%s), but %s", expect, hash)
	}
}

func testClient(url string) Client {
	ipfs, _ := NewIPFS(url)
	return Client{
		ipfs: ipfs,
	}
}
