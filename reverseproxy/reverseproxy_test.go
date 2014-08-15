package reverseproxy_test

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"testing"

	. "github.com/mix3/go-tenco/reverseproxy"
	"github.com/mix3/go-tenco/storage"
	"github.com/mix3/go-tenco/storage/sqlite"
	"github.com/stretchr/testify/assert"
)

func TestReverseProxy(t *testing.T) {
	strage, err := sqlite.New(":memory:")
	if err != nil {
		t.Fatal(err)
	}
	strage.Set("hoge", "http://localhost:8081")
	strage.Set("fuga", "http://localhost:8082")
	handler := New("example.com", strage, nil)
	go func() {
		mux := http.NewServeMux()
		mux.HandleFunc("/", func(w http.ResponseWriter, req *http.Request) {
			fmt.Fprint(w, "hoge")
		})
		s := http.Server{Addr: ":8081", Handler: mux}
		t.Fatal(s.ListenAndServe())
	}()
	go func() {
		mux := http.NewServeMux()
		mux.HandleFunc("/", func(w http.ResponseWriter, req *http.Request) {
			fmt.Fprint(w, "fuga")
		})
		s := http.Server{Addr: ":8082", Handler: mux}
		t.Fatal(s.ListenAndServe())
	}()
	go func() {
		mux := http.NewServeMux()
		mux.HandleFunc("/", func(w http.ResponseWriter, req *http.Request) {
			fmt.Fprint(w, "piyo")
		})
		s := http.Server{Addr: ":8083", Handler: mux}
		t.Fatal(s.ListenAndServe())
	}()
	go func() {
		s := http.Server{Addr: ":8080", Handler: handler}
		t.Fatal(s.ListenAndServe())
	}()

	checkRequest(t, "hoge.example.com", "http://localhost:8080", "hoge")
	checkRequest(t, "hoge.example.com", "http://localhost:8080/test", "hoge")
	checkRequest(t, "fuga.example.com", "http://localhost:8080", "fuga")
	checkRequest(t, "piyo.example.com", "http://localhost:8080", "404 page not found\n")
	strage.Set("piyo", "http://localhost:8083")
	checkRequest(t, "piyo.example.com", "http://localhost:8080", "piyo")
	strage.Delete("hoge")
	checkRequest(t, "hoge.example.com", "http://localhost:8080", "404 page not found\n")
}

func TestHandler(t *testing.T) {
	strage, err := sqlite.New(":memory:")
	if err != nil {
		t.Fatal(err)
	}
	strage.Set("hoge", "http://localhost:8081")
	rootHandler := func(w http.ResponseWriter, req *http.Request, s storage.Storager) {
		v, _ := s.Get("hoge")
		fmt.Fprint(w, v)
	}
	handler := New("example.com", strage, rootHandler)
	go func() {
		s := http.Server{Addr: ":8084", Handler: handler}
		t.Fatal(s.ListenAndServe())
	}()

	checkRequest(t, "example.com", "http://localhost:8084", "http://localhost:8081")
}

func checkRequest(t *testing.T, subdomain, url, exp string) {
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		t.Fatal(err)
	}
	req.Host = subdomain
	c := http.Client{}
	res, err := c.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	defer res.Body.Close()
	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, exp, string(body))
}
