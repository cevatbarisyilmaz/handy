package handy_test

import (
	"bytes"
	"context"
	"github.com/cevatbarisyilmaz/ara"
	"github.com/cevatbarisyilmaz/handy"
	"io/ioutil"
	"net"
	"net/http"
	"strconv"
	"testing"
)

func TestNakedToWWWHandler_ServeHTTP(t *testing.T) {
	const host = "example.com"
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	sm := http.NewServeMux()
	sm.Handle(host+"/", handy.NewNakedToWWWHandler())
	sm.Handle("www."+host+"/", greeter{})
	server := &http.Server{
		Handler: sm,
	}
	go func() {
		_ = server.Serve(listener)
	}()
	defer func() {
		err := server.Shutdown(context.Background())
		if err != nil {
			t.Fatal(err)
		}
	}()
	response, err := ara.NewClient(loopbackResolver{}).Get("http://" + net.JoinHostPort(host, strconv.Itoa(listener.Addr().(*net.TCPAddr).Port)))
	if err != nil {
		t.Fatal(response)
	}
	if response.StatusCode != http.StatusOK {
		t.Fatal("want", http.StatusOK, "got", response.StatusCode)
	}
	want := net.JoinHostPort("www."+host, strconv.Itoa(listener.Addr().(*net.TCPAddr).Port))
	if response.Request.URL.Host != want {
		t.Fatal("want", want, "got", response.Request.URL.Host)
	}
	body, err := ioutil.ReadAll(response.Body)
	if !bytes.Equal(body, greetingBytes) {
		t.Fatal("want", greetingString, "got", string(body))
	}
}

func TestNakedToWWWWrapper_ServeHTTP(t *testing.T) {
	const host = "example.com"
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	server := &http.Server{
		Handler: handy.NewNakedToWWWWrapper(greeter{}),
	}
	go func() {
		_ = server.Serve(listener)
	}()
	defer func() {
		err := server.Shutdown(context.Background())
		if err != nil {
			t.Fatal(err)
		}
	}()
	response, err := ara.NewClient(loopbackResolver{}).Get("http://" + net.JoinHostPort(host, strconv.Itoa(listener.Addr().(*net.TCPAddr).Port)))
	if err != nil {
		t.Fatal(response)
	}
	if response.StatusCode != http.StatusOK {
		t.Fatal("want", http.StatusOK, "got", response.StatusCode)
	}
	want := net.JoinHostPort("www."+host, strconv.Itoa(listener.Addr().(*net.TCPAddr).Port))
	if response.Request.URL.Host != want {
		t.Fatal("want", want, "got", response.Request.URL.Host)
	}
	body, err := ioutil.ReadAll(response.Body)
	if !bytes.Equal(body, greetingBytes) {
		t.Fatal("want", greetingString, "got", string(body))
	}
}
