// Copyright (c) 2012-2013 Jason McVetta.  This is Free Software, released
// under the terms of the GPL v3.  See http://www.gnu.org/copyleft/gpl.html for
// details.  Resist intellectual serfdom - the ownership of ideas is akin to
// slavery.

package napping

import (
	"encoding/json"
	"github.com/bmizerany/assert"
	"io/ioutil"
	"log"
	"net/http"
	"net/http/httptest"
	"testing"
)

func init() {
	log.SetFlags(log.Ltime | log.Lshortfile)
}

func TestInvalidUrl(t *testing.T) {
	//
	//  Missing protocol scheme - url.Parse should fail
	//

	url := "://foobar.com"
	_, err := Get(url, nil, nil, nil)
	assert.NotEqual(t, nil, err)
	//
	// Unsupported protocol scheme - HttpClient.Do should fail
	//
	url = "foo://bar.com"
	_, err = Get(url, nil, nil, nil)
	assert.NotEqual(t, nil, err)
}

type structType struct {
	Foo int
	Bar string
}

type errorStruct struct {
	Status  int
	Message string
}

var (
	fooParams = Params{"foo": "bar"}
	barParams = Params{"bar": "baz"}
	fooStruct = structType{
		Foo: 111,
		Bar: "foo",
	}
	barStruct = structType{
		Foo: 222,
		Bar: "bar",
	}
)

func TestGet(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(HandleGet))
	defer srv.Close()
	//
	// Good request
	//
	url := "http://" + srv.Listener.Addr().String()
	p := fooParams
	res := structType{}
	resp, err := Get(url, &p, &res, nil)
	if err != nil {
		t.Error(err)
	}
	assert.Equal(t, 200, resp.Status())
	assert.Equal(t, res, barStruct)
	//
	// Bad request
	//
	url = "http://" + srv.Listener.Addr().String()
	p = Params{"bad": "value"}
	e := errorStruct{}
	opts := Opts{
		ExpectedStatus: 200,
	}
	resp, err = Get(url, &p, nil, &opts)
	if err != UnexpectedStatus {
		t.Error(err)
	}
	assert.Equal(t, 500, resp.Status())
	expected := errorStruct{
		Message: "Bad query params: bad=value",
		Status:  500,
	}
	resp.Unmarshall(&e)
	assert.Equal(t, e, expected)
}

func TestPost(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(HandlePost))
	defer srv.Close()
	s := Session{}
	s.Log = true
	url := "http://" + srv.Listener.Addr().String()
	payload := fooStruct
	res := structType{}
	resp, err := s.Post(url, &payload, &res, nil)
	if err != nil {
		t.Error(err)
	}
	assert.Equal(t, 200, resp.Status())
	assert.Equal(t, res, barStruct)
}

func TestPostUnmarshallable(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(HandlePost))
	defer srv.Close()
	type ft func()
	var f ft
	url := "http://" + srv.Listener.Addr().String()
	res := structType{}
	payload := f
	_, err := Post(url, &payload, &res, nil)
	assert.NotEqual(t, nil, err)
	_, ok := err.(*json.UnsupportedTypeError)
	if !ok {
		t.Log(err)
		t.Error("Expected json.UnsupportedTypeError")
	}
}

func TestPut(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(HandlePut))
	defer srv.Close()
	url := "http://" + srv.Listener.Addr().String()
	res := structType{}
	resp, err := Put(url, &fooStruct, &res, nil)
	if err != nil {
		t.Error(err)
	}
	assert.Equal(t, resp.Status(), 200)
	// Server should return NO data
	assert.Equal(t, resp.RawText(), "")
}

func JsonError(w http.ResponseWriter, msg string, code int) {
	e := errorStruct{
		Status:  code,
		Message: msg,
	}
	blob, err := json.Marshal(e)
	if err != nil {
		http.Error(w, msg, code)
		return
	}
	http.Error(w, string(blob), code)
}

func HandleGet(w http.ResponseWriter, req *http.Request) {
	u := req.URL
	q := u.Query()
	for k, _ := range fooParams {
		if fooParams[k] != q.Get(k) {
			msg := "Bad query params: " + u.Query().Encode()
			JsonError(w, msg, http.StatusInternalServerError)
			return
		}
	}
	//
	// Generate response
	//
	blob, err := json.Marshal(barStruct)
	if err != nil {
		JsonError(w, err.Error(), http.StatusInternalServerError)
		return
	}
	req.Header.Add("content-type", "application/json")
	w.Write(blob)
}

func HandlePost(w http.ResponseWriter, req *http.Request) {
	//
	// Parse Payload
	//
	if req.ContentLength <= 0 {
		msg := "Content-Length must be greater than 0."
		JsonError(w, msg, http.StatusLengthRequired)
		return
	}
	body, err := ioutil.ReadAll(req.Body)
	if err != nil {
		JsonError(w, err.Error(), http.StatusInternalServerError)
		return
	}
	var s structType
	err = json.Unmarshal(body, &s)
	if err != nil {
		JsonError(w, err.Error(), http.StatusBadRequest)
		return
	}
	if s != fooStruct {
		msg := "Bad request body"
		JsonError(w, msg, http.StatusBadRequest)
		return
	}
	//
	// Compose Response
	//
	blob, err := json.Marshal(barStruct)
	if err != nil {
		JsonError(w, err.Error(), http.StatusInternalServerError)
		return
	}
	req.Header.Add("content-type", "application/json")
	w.Write(blob)
}

func HandlePut(w http.ResponseWriter, req *http.Request) {
	//
	// Parse Payload
	//
	if req.ContentLength <= 0 {
		msg := "Content-Length must be greater than 0."
		JsonError(w, msg, http.StatusLengthRequired)
		return
	}
	body, err := ioutil.ReadAll(req.Body)
	if err != nil {
		JsonError(w, err.Error(), http.StatusInternalServerError)
		return
	}
	var s structType
	err = json.Unmarshal(body, &s)
	if err != nil {
		JsonError(w, err.Error(), http.StatusBadRequest)
		return
	}
	if s != fooStruct {
		msg := "Bad request body"
		JsonError(w, msg, http.StatusBadRequest)
		return
	}
	return
}
