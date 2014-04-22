package response

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-martini/martini"
	"github.com/martini-contrib/encoder"
)

type user struct {
	Name   string `json:"name"`
	Friend string `json:"friend,omitempty"`
}

func setupMartini() *martini.ClassicMartini {
	m := martini.Classic()

	m.Use(func(c martini.Context, w http.ResponseWriter) {
		c.MapTo(encoder.JsonEncoder{}, (*encoder.Encoder)(nil))
		w.Header().Set("Content-Type", "application/json; charset=utf-8")
	})
	m.Use(NewEncoder())

	return m
}

func decodeResponseBody(buf *bytes.Buffer) (error, interface{}) {
	var usr user
	err := json.Unmarshal(buf.Bytes(), &usr)
	if err != nil {
		return err, nil
	}
	return nil, usr
}

func setupHttpTester(method, path string) (*httptest.ResponseRecorder, *http.Request) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(method, path, nil)
	return w, r
}

func TestSimpleHandler(t *testing.T) {
	m := setupMartini()

	m.Get("/user", func() user {
		return user{Name: "John"}
	})

	w, r := setupHttpTester("GET", "/user")
	m.ServeHTTP(w, r)

	resp := w.Body
	err, usrInt := decodeResponseBody(resp)
	if err != nil {
		t.Error(err)
	}

	usr := usrInt.(user)
	if usr.Name != "John" {
		t.Fatal("Expected John, Got:", usr.Name)
	}
}

func TestMultiCallingHandler(t *testing.T) {
	m := setupMartini()

	m.Get("/user", userHandler)

	m.Get("/friend", friendHandler)

	w, r := setupHttpTester("GET", "/friend")
	m.ServeHTTP(w, r)

	resp := w.Body
	err, friendInt := decodeResponseBody(resp)
	if err != nil {
		t.Error(err)
	}

	friend := friendInt.(user)
	if friend.Name != "Zach" {
		t.Fatal("Expected name Zach, Got:", friend.Name)
	}

	w, r = setupHttpTester("GET", "/user")
	m.ServeHTTP(w, r)

	resp = w.Body
	err, usrInt := decodeResponseBody(resp)
	if err != nil {
		t.Error(err)
	}

	usr := usrInt.(user)
	if usr.Name != "John" {
		t.Fatal("Expected name John, Got:", usr.Name)
	}
	if usr.Friend != "Zach" {
		t.Fatal("Expected friend name Zach, Got:", usr.Friend)
	}
}

func Test404NotFound(t *testing.T) {
	m := setupMartini()

	w, r := setupHttpTester("GET", "/friend")
	m.ServeHTTP(w, r)

	resp, err := ioutil.ReadAll(w.Body)
	if err != nil {
		t.Fail()
	} else if string(resp) != "404 page not found\n" {
		t.Fatalf("Expected %v, Got: %v", []byte("404 page not found"), resp)
	}
}

func userHandler() user {
	friend := friendHandler()
	return user{Name: "John", Friend: friend.Name}
}

func friendHandler() user {
	return user{Name: "Zach"}
}
