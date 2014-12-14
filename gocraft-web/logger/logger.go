package web

import (
	log "github.com/Sirupsen/logrus"
	"github.com/gocraft/web"

	"crypto/rand"
	"encoding/hex"
	"net"
	"strconv"
	"time"
)

func mustRandRead(buf []byte) {
	// TODO: i bet this is slow as shit.
	_, err := rand.Read(buf)
	if err != nil {
		panic(err)
	}
}

func reqId() string {
	id := make([]byte, 8)
	mustRandRead(id)
	return hex.EncodeToString(id)
}

//
// Structured Logging middleware for gocraft/web,
// using logrus.  Sets "Request-Id" header to a unique
// value per-request, and trusts x-forwarded-for headers
// if present.
//
func Middleware(rw web.ResponseWriter, req *web.Request, next web.NextMiddlewareFunc) {

	startTime := time.Now()

	id := reqId()
	req.Header.Set("Request-Id", id)
	rw.Header().Set("Request-Id", id)

	ip := req.Header.Get("x-forwarded-for")
	if ip == "" {
		ip, _, _ = net.SplitHostPort(req.RemoteAddr)
	}

	ua := req.Header.Get("user-agent")

	log.WithFields(log.Fields{
		"method":     req.Method,
		"path":       req.URL.Path,
		"ip":         ip,
		"user-agent": ua,
		"id":         id,
	}).Info("request started")

	next(rw, req)

	duration := time.Since(startTime).Nanoseconds()
	var durationUnits string
	switch {
	case duration > 2000000:
		durationUnits = "ms"
		duration /= 1000000
	case duration > 1000:
		durationUnits = "μs"
		duration /= 1000
	default:
		durationUnits = "ns"
	}

	dstr := strconv.FormatInt(duration, 10) + durationUnits

	log.WithFields(log.Fields{
		"method":     req.Method,
		"status":     rw.StatusCode(),
		"path":       req.URL.Path,
		"ip":         ip,
		"user-agent": ua,
		"id":         id,
		"duration":   dstr,
	}).Info("request completed")
}
