package middleware

import (
	"net"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/nika/soccer-manager-api/pkg/auth"
	"github.com/nika/soccer-manager-api/pkg/response"
)

const (
	throttleMaxPerWindow = 10
	throttleWindow       = time.Minute
	evictThreshold       = 2048
)

type throttleBucket struct {
	window int64 // unix timestamp truncated to throttleWindow
	count  int
}

var (
	throttleMu      sync.Mutex
	throttleBuckets = make(map[string]*throttleBucket)
)

func clientIP(r *http.Request) string {
	if xff := strings.TrimSpace(r.Header.Get("X-Forwarded-For")); xff != "" {
		if i := strings.IndexByte(xff, ','); i >= 0 {
			xff = strings.TrimSpace(xff[:i])
		}
		return xff
	}
	if xr := strings.TrimSpace(r.Header.Get("X-Real-IP")); xr != "" {
		return xr
	}
	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		return r.RemoteAddr
	}
	return host
}

func throttleKey(r *http.Request) string {
	if id := auth.UserIDFromContext(r.Context()); id != 0 {
		return "u:" + strconv.FormatInt(id, 10)
	}
	return "ip:" + clientIP(r)
}

func throttleEvictStale(current int64) {
	if len(throttleBuckets) < evictThreshold {
		return
	}
	for k, b := range throttleBuckets {
		if current-b.window > 1 {
			delete(throttleBuckets, k)
		}
	}
}

// Throttle allows at most 10 requests per minute per key. The key is the
// authenticated user ID when present in context (place after JWT middleware),
// otherwise the client IP.
func Throttle(next http.Handler) http.Handler {
	windowSecs := int64(throttleWindow / time.Second)
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		key := throttleKey(r)
		now := time.Now().Unix()
		win := now / windowSecs

		throttleMu.Lock()
		b, ok := throttleBuckets[key]
		if !ok {
			b = &throttleBucket{}
			throttleBuckets[key] = b
		}
		if b.window != win {
			b.window = win
			b.count = 0
		}
		if b.count >= throttleMaxPerWindow {
			throttleMu.Unlock()
			w.Header().Set("Retry-After", "60")
			response.Error(w, http.StatusTooManyRequests, "rate limit exceeded")
			return
		}
		b.count++
		throttleEvictStale(win)
		throttleMu.Unlock()

		next.ServeHTTP(w, r)
	})
}
