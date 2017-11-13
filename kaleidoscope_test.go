package kaleidoscope

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestKaleidoScopeCreateAndSave(t *testing.T) {
	dbname := "dbname"

	expectForAdd := "QmSomeLinkHash"
	resForAdd := `{"Name":"file.txt","Hash":"QmSomeObjectHash","Size":"13"}
{"Name":"","Hash":"%s","Size":"67"}`

	expectForAddLink := "QmSomeNewLinkHash"
	resForAddLink := `{"Hash":"%s"}`

	ipfs := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/api/v0/key/gen" {
			fmt.Fprintln(w, `{"Name":"some_key","Id":"QmSomePeerID"}`)
		} else if r.URL.Path == "/api/v0/add" {
			fmt.Fprintln(w, fmt.Sprintf(resForAdd, expectForAdd))
		} else if r.URL.Path == "/api/v0/object/patch/add-link" {
			fmt.Fprintln(w, fmt.Sprintf(resForAddLink, expectForAddLink))
		} else if r.URL.Path == "/api/v0/name/publish" {
			fmt.Fprintln(w, `{"Name":"QmSomeName","Value":"/ipfs/QmSomeValue"}`)
		}
	}))
	defer ipfs.Close()

	kes := testKaleidoScope(ipfs.URL)
	kes.keystore.persistence = false
	dbhash, err := kes.Create(dbname, 2048)

	if err != nil {
		t.Errorf("Create should not return error, but %s", err)
	}
	if dbhash != expectForAddLink {
		t.Errorf("Create should return database's hash (%s), but %s", expectForAddLink, dbhash)
	}

	if kes.dbname != dbname {
		t.Errorf("Create should set current dbname (%s), but %s", dbname, kes.dbname)
	}

	if kes.head != expectForAddLink {
		t.Errorf("Create should set current hash (%s), but %s", expectForAddLink, kes.head)
	}

	err = kes.Save()
	if err != nil {
		t.Errorf("Save should not return error, but %s", err)
	}
}

func TestKaleidoScopeGet(t *testing.T) {
	expect := wrapWithMetadata("Some value")
	keystore := NewKeyStore()
	keystore.persistence = false
	keystore.Load("dummy")
	bs, _ := keystore.EncryptString(expect)

	// expect := string(bs)

	ipfs := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write(bs)
	}))
	defer ipfs.Close()

	kes := testKaleidoScope(ipfs.URL)
	kes.keystore = keystore
	kes.Use("dummy")
	meta, value, err := kes.Get("some_key")

	if err != nil {
		t.Errorf("Get should not return error, but %s", err)
	}
	expects := strings.Split(expect, ",")[:2]
	expectMeta, expectValue := expects[0], expects[1]

	if string(meta) != expectMeta {
		t.Errorf("Get should return meta (%s), but %s", expectMeta, string(meta))
	}
	if string(value) != expectValue {
		t.Errorf("Get should return value (%s), but %s", expectValue, string(value))
	}
}

func testKaleidoScope(url string) Kaleidoscope {
	kes, _ := New()
	kes.client = testClient(url)
	return kes
}
