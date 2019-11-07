package handy_test

import (
	"context"
	"log"
	"net/http"
)

const greetingString = "Hi there!"

var greetingBytes = []byte(greetingString)

type greeter struct {
}

func (g greeter) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	_, err := w.Write(greetingBytes)
	if err != nil {
		log.Fatal(err)
	}
}

type loopbackResolver struct {
}

func (r loopbackResolver) LookupHost(ctx context.Context, host string) ([]string, error) {
	return []string{"127.0.0.1"}, nil
}
