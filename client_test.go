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
	hash, err := client.ObjectPatchAddLink("QmSomeRootHash", "new_file.txt",
		"QmSomeObjectHash", RequestOptions{})

	if err != nil {
		t.Errorf("ObjectPatchAddLink should not return error, but %s", err)
	}
	if hash != expect {
		t.Errorf("ObjectPatchAddLink should return new link's hash (%s), but %s", expect, hash)
	}
}

func TestClientCat(t *testing.T) {
	expect := "Some value"

	ipfs := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, expect)
	}))
	defer ipfs.Close()

	client := testClient(ipfs.URL)
	value, err := client.Cat("QmSomeObjectHash", RequestOptions{})

	if err != nil {
		t.Errorf("Cat should not return error, but %s", err)
	}
	if string(value) != expect {
		t.Errorf("Cat should return value (%s), but %s", expect, string(value))
	}
}

func TestClientKeyGen(t *testing.T) {
	ipfs := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, `{"Name":"some_key","Id":"QmSomePeerID"}`)
	}))
	defer ipfs.Close()

	client := testClient(ipfs.URL)
	err := client.KeyGen("some_key", RequestOptions{})

	if err != nil {
		t.Errorf("KeyGen should not return error, but %s", err)
	}
}

func TestClientNamePublish(t *testing.T) {
	expect := "QmSomeObjectHash"
	expectPath := "/ipfs/" + expect

	ipfs := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, fmt.Sprintf(`{"Name":"QmSomeName","Value":"/ipfs/%s"}`, expect))
	}))
	defer ipfs.Close()

	client := testClient(ipfs.URL)
	_, value, err := client.NamePublish(expect, RequestOptions{})

	if err != nil {
		t.Errorf("NamePublish should not return error, but %s", err)
	}
	if value != expectPath {
		t.Errorf("NamePublish should return value (%s), but %s", expectPath, value)
	}
}

func testClient(url string) Client {
	ipfs, _ := NewIPFS(url)
	return Client{
		ipfs: ipfs,
	}
}
