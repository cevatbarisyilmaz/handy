package handy_test

import (
	"bytes"
	"context"
	"crypto/rand"
	"crypto/rsa"
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"github.com/cevatbarisyilmaz/ara"
	"github.com/cevatbarisyilmaz/handy"
	"io/ioutil"
	"math/big"
	"net"
	"net/http"
	"testing"
	"time"
)

func getCert() (*tls.Certificate, error) {
	serialNumberLimit := new(big.Int).Lsh(big.NewInt(1), 128)
	serialNumber, err := rand.Int(rand.Reader, serialNumberLimit)
	if err != nil {
		return nil, err
	}
	template := &x509.Certificate{
		SerialNumber: serialNumber,
		Subject: pkix.Name{
			Organization: []string{"Stark Industries"},
		},
		NotBefore:             time.Now(),
		NotAfter:              time.Now().Add(time.Hour),
		KeyUsage:              x509.KeyUsageKeyEncipherment | x509.KeyUsageDigitalSignature,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
		BasicConstraintsValid: true,
	}
	key, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		return nil, err
	}
	derBytes, err := x509.CreateCertificate(rand.Reader, template, template, key.Public(), key)
	if err != nil {
		return nil, err
	}
	return &tls.Certificate{
		Certificate: [][]byte{derBytes},
		PrivateKey:  key,
	}, nil
}

func TestHTTPToHTTPSHandler_ServeHTTP(t *testing.T) {
	cert, err := getCert()
	if err != nil {
		t.Fatal(err)
	}
	httpServer := &http.Server{
		Addr:    "127.0.0.1:http",
		Handler: handy.NewHTTPToHTTPSHandler(),
	}
	go func() {
		_ = httpServer.ListenAndServe()
	}()
	defer func() {
		err := httpServer.Shutdown(context.Background())
		if err != nil {
			t.Fatal(err)
		}
	}()
	httpsServer := &http.Server{
		Addr: "127.0.0.1:https",
		TLSConfig: &tls.Config{
			Certificates: []tls.Certificate{*cert},
		},
		Handler: greeter{},
	}
	go func() {
		_ = httpsServer.ListenAndServeTLS("", "")
	}()
	defer func() {
		err := httpsServer.Shutdown(context.Background())
		if err != nil {
			t.Fatal(err)
		}
	}()
	client := http.Client{
		Transport: &http.Transport{
			DialContext: (&ara.Dialer{
				Resolver: loopbackResolver{},
			}).DialContext,
			TLSClientConfig: &tls.Config{
				InsecureSkipVerify: true,
			},
		},
	}
	res, err := client.Get("http://example.com")
	if err != nil {
		t.Fatal(err)
	}
	if res.StatusCode != http.StatusOK {
		t.Fatal("want", http.StatusOK, "got", res.StatusCode)
	}
	if res.TLS == nil {
		t.Fatal("no https")
	}
	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(body, greetingBytes) {
		t.Fatal("want", greetingString, "got", string(body))
	}
}

func TestHTTPToHTTPSWrapper_ServeHTTP(t *testing.T) {
	cert, err := getCert()
	if err != nil {
		t.Fatal(err)
	}
	server := &http.Server{
		Handler:   handy.NewHTTPToHTTPSWrapper(greeter{}),
		TLSConfig: &tls.Config{Certificates: []tls.Certificate{*cert}},
	}
	go func() {
		listener, err := net.Listen("tcp", "127.0.0.1:http")
		if err != nil {
			t.Fatal(err)
		}
		_ = server.Serve(listener)
	}()
	go func() {
		listener, err := net.Listen("tcp", "127.0.0.1:https")
		if err != nil {
			t.Fatal(err)
		}
		_ = server.ServeTLS(listener, "", "")
	}()
	defer func() {
		err := server.Shutdown(context.Background())
		if err != nil {
			t.Fatal(err)
		}
	}()
	client := http.Client{
		Transport: &http.Transport{
			DialContext: (&ara.Dialer{
				Resolver: loopbackResolver{},
			}).DialContext,
			TLSClientConfig: &tls.Config{
				InsecureSkipVerify: true,
			},
		},
	}
	res, err := client.Get("http://example.com")
	if err != nil {
		t.Fatal(err)
	}
	if res.StatusCode != http.StatusOK {
		t.Fatal("want", http.StatusOK, "got", res.StatusCode)
	}
	if res.TLS == nil {
		t.Fatal("no https")
	}
	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(body, greetingBytes) {
		t.Fatal("want", greetingString, "got", string(body))
	}
}
