package kaleidoscope

import (
	"bytes"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestKaleidoScopeCreate(t *testing.T) {
	dbname := "dbname"

	expectForAdd := "QmSomeLinkHash"
	resForAdd := `{"Name":"file.txt","Hash":"QmSomeObjectHash","Size":"13"}
{"Name":"","Hash":"%s","Size":"67"}`

	expectForAddLink := "QmSomeNewLinkHash"
	resForAddLink := `{"Hash":"%s"}`

	ipfs := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/api/v0/add" {
			fmt.Fprintln(w, fmt.Sprintf(resForAdd, expectForAdd))
		} else if r.URL.Path == "/api/v0/object/patch/add-link" {
			fmt.Fprintln(w, fmt.Sprintf(resForAddLink, expectForAddLink))
		}
	}))
	defer ipfs.Close()

	kes := testKaleidoScope(ipfs.URL)
	dbhash, err := kes.Create(dbname)

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
}

func TestKaleidoScopeGet(t *testing.T) {
	buf := new(bytes.Buffer)
	buf.ReadFrom(wrapWithMetadata("Some value"))
	expect := buf.String()

	ipfs := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, expect)
	}))
	defer ipfs.Close()

	kes := testKaleidoScope(ipfs.URL)
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
