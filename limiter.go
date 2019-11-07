package handy

import (
	"log"
	"math"
	"net"
	"net/http"
	"strconv"
	"sync"
	"time"
)

type record struct {
	history []time.Time
	mu      *sync.Mutex
}

func newRecord() *record {
	return &record{history: []time.Time{}, mu: &sync.Mutex{}}
}

type limiterWrapper struct {
	oldRecords    map[[16]byte]*record
	recentRecords map[[16]byte]*record
	mu            *sync.Mutex
	limit         *struct {
		D time.Duration
		R int
	}
	wrapped          http.Handler
	goroutineWorking bool
}

// NewLimiterWrapper returns a handler that will drop the
// request if exceeds the given (request amount)/time limit,
// otherwise it will invoke the wrapped handler.
func NewLimiterWrapper(wrapped http.Handler, duration time.Duration, requestAmount int) *limiterWrapper {
	return &limiterWrapper{
		oldRecords:    map[[16]byte]*record{},
		recentRecords: map[[16]byte]*record{},
		mu:            &sync.Mutex{},
		limit: &struct {
			D time.Duration
			R int
		}{D: duration, R: requestAmount},
		wrapped:          wrapped,
		goroutineWorking: false,
	}
}

func (h *limiterWrapper) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	addr, err := net.ResolveTCPAddr("tcp", r.RemoteAddr)
	if err != nil {
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		log.Println("limiterWrapper: could not resolve address:", r.RemoteAddr)
		return
	}
	re := h.getRecord(addr.IP)
	re.mu.Lock()
	now := time.Now()
	if len(re.history) == h.limit.R {
		t := re.history[0].Add(h.limit.D).Sub(now)
		if t <= 0 {
			re.history = append(re.history[1:], now)
		} else {
			w.Header().Add("Retry-After", strconv.Itoa(int(math.Ceil(t.Seconds()))))
			http.Error(w, http.StatusText(http.StatusTooManyRequests), http.StatusTooManyRequests)
			re.mu.Unlock()
			return
		}
	} else {
		re.history = append(re.history, now)
	}
	re.mu.Unlock()
	h.wrapped.ServeHTTP(w, r)
}

func (h *limiterWrapper) getRecord(ip net.IP) *record {
	var ipByte [16]byte
	copy(ipByte[:], ip.To16())
	h.mu.Lock()
	defer h.mu.Unlock()
	r := h.recentRecords[ipByte]
	if r == nil {
		r = h.oldRecords[ipByte]
		if r == nil {
			r = newRecord()
		}
		h.recentRecords[ipByte] = r
	}
	if !h.goroutineWorking {
		h.goroutineWorking = true
		go func() {
			for {
				time.Sleep(h.limit.D)
				h.mu.Lock()
				h.oldRecords = h.recentRecords
				h.recentRecords = make(map[[16]byte]*record)
				if len(h.oldRecords) == 0 {
					h.goroutineWorking = false
					h.mu.Unlock()
					return
				}
				h.mu.Unlock()

			}
		}()
	}
	return r
}
