package handy_test

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"github.com/cevatbarisyilmaz/handy"
	"io/ioutil"
	"net"
	"net/http"
	"strconv"
	"sync"
	"testing"
	"time"
)

const duration = time.Second * 4
const requestAmount = 16
const penetrationDuration = duration * 4
const sensitivity = 64 * time.Millisecond

type agent struct {
	ip net.IP
}

func (a *agent) penetrate(raddr string) error {
	client := &http.Client{
		Timeout: duration,
		Transport: &http.Transport{
			DialContext: (&net.Dialer{
				Timeout: duration,
				LocalAddr: &net.TCPAddr{
					IP:   a.ip,
					Port: 0,
				},
			}).DialContext,
		},
	}
	var successfulHits []time.Time
	start := time.Now()
	for {
		if time.Now().Sub(start) >= penetrationDuration {
			return nil
		}
		response, err := client.Get(raddr)
		if err != nil {
			return err
		}
		now := time.Now()
		if response.StatusCode == http.StatusOK {
			if len(successfulHits) < requestAmount {
				successfulHits = append(successfulHits, now)
			} else {
				interval := now.Sub(successfulHits[0])
				if interval+sensitivity < duration {
					fmt.Println(interval)
					return errors.New("too early successful hit by " + (duration - interval).String())
				} else if interval > duration+sensitivity {
					return errors.New("too late successful hit by " + (interval - duration).String())
				}
				successfulHits = append(successfulHits[1:], now)
			}
			body, err := ioutil.ReadAll(response.Body)
			if err != nil {
				return err
			}
			err = response.Body.Close()
			if err != nil {
				return err
			}
			if !bytes.Equal(body, greetingBytes) {
				return errors.New("want: " + greetingString + " got: " + string(body))
			}
		} else if response.StatusCode == http.StatusTooManyRequests {
			if len(successfulHits) < requestAmount {
				return errors.New("unexpected miss")
			} else {
				interval := now.Sub(successfulHits[0])
				if interval > duration+sensitivity {
					return errors.New("unexpected miss " + interval.String() + " passed since the corresponding successful hit")
				}
			}
		} else {
			return errors.New("unexpected status code" + strconv.Itoa(response.StatusCode))
		}
		time.Sleep(time.Millisecond * 16)
	}
}

func TestLimiterWrapper_ServeHTTP(t *testing.T) {
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	server := &http.Server{Handler: handy.NewLimiterWrapper(greeter{}, duration, requestAmount)}
	go func() {
		_ = server.Serve(listener)
	}()
	defer func() {
		err := server.Shutdown(context.Background())
		if err != nil {
			t.Fatal(err)
		}
	}()
	wg := &sync.WaitGroup{}
	errChan := make(chan error, 1)
	for i := byte(1); i < 10; i++ {
		wg.Add(1)
		go func(i byte) {
			a := agent{ip: net.IPv4(127, 0, 0, i)}
			if err := a.penetrate("http://" + listener.Addr().String()); err != nil {
				errChan <- err
			}
			wg.Done()
		}(i)
	}
	go func() {
		wg.Wait()
		errChan <- nil
	}()
	err = <-errChan
	if err != nil {
		t.Fatal(err)
	}
}
