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

func TestHostAuthWrapper_ServeHTTP(t *testing.T) {
	const host = "example.com"
	const fakeHost = "fakexample.com"
	tests := []*struct {
		Handler http.Handler
		Cases   map[string]bool
	}{
		{
			Handler: handy.NewHostAuthWrapper(greeter{}, host, nil),
			Cases: map[string]bool{
				"www." + host:     true,
				"en." + host:      true,
				host:              true,
				"www." + fakeHost: false,
				fakeHost:          false,
			},
		},
		{
			Handler: handy.NewHostAuthWrapper(greeter{}, host, []string{"www", "en"}),
			Cases: map[string]bool{
				"www." + host:     true,
				"en." + host:      true,
				"tr." + host:      false,
				host:              false,
				"www." + fakeHost: false,
				fakeHost:          false,
			},
		},
		{
			Handler: handy.NewHostAuthWrapper(greeter{}, host, []string{"www", "en", ""}),
			Cases: map[string]bool{
				"www." + host:     true,
				"en." + host:      true,
				"tr." + host:      false,
				host:              true,
				"www." + fakeHost: false,
				fakeHost:          false,
			},
		},
	}
	client := ara.NewClient(loopbackResolver{})
	for _, test := range tests {
		listener, err := net.Listen("tcp", "127.0.0.1:0")
		if err != nil {
			t.Fatal(err)
		}
		server := &http.Server{
			Handler: test.Handler,
		}
		go func() {
			_ = server.Serve(listener)
		}()
		for c, r := range test.Cases {
			res, err := client.Get("http://" + net.JoinHostPort(c, strconv.Itoa(listener.Addr().(*net.TCPAddr).Port)))
			if err != nil {
				t.Error(err)
				continue
			}
			if res.StatusCode == http.StatusMisdirectedRequest && r {
				t.Error("want", http.StatusOK, "got", http.StatusMisdirectedRequest)
			} else if !r && res.StatusCode != http.StatusMisdirectedRequest {
				t.Error("want", http.StatusMisdirectedRequest, "got", res.StatusCode)
			} else if res.StatusCode == http.StatusOK {
				body, err := ioutil.ReadAll(res.Body)
				if err != nil {
					t.Error(err)
					continue
				}
				if !bytes.Equal(body, greetingBytes) {
					t.Error("want", greetingString, "got", string(greetingBytes))
				}
			}
		}
		err = server.Shutdown(context.Background())
		if err != nil {
			t.Fatal(err)
		}
	}
}
