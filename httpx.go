// Copyright 2019 Andrei Tudor Călin
//
// Permission to use, copy, modify, and/or distribute this software for any
// purpose with or without fee is hereby granted, provided that the above
// copyright notice and this permission notice appear in all copies.
//
// THE SOFTWARE IS PROVIDED "AS IS" AND THE AUTHOR DISCLAIMS ALL WARRANTIES
// WITH REGARD TO THIS SOFTWARE INCLUDING ALL IMPLIED WARRANTIES OF
// MERCHANTABILITY AND FITNESS. IN NO EVENT SHALL THE AUTHOR BE LIABLE FOR
// ANY SPECIAL, DIRECT, INDIRECT, OR CONSEQUENTIAL DAMAGES OR ANY DAMAGES
// WHATSOEVER RESULTING FROM LOSS OF USE, DATA OR PROFITS, WHETHER IN AN
// ACTION OF CONTRACT, NEGLIGENCE OR OTHER TORTIOUS ACTION, ARISING OUT OF
// OR IN CONNECTION WITH THE USE OR PERFORMANCE OF THIS SOFTWARE.

// Package httpx provides convenience extensions to net/http, enabling a
// certain style of web programming which resembles the implementation
// of recursive descent parsers.
package httpx

import (
	"context"
	"net/http"
	"strings"
	"time"

	"acln.ro/log"
	"github.com/felixge/httpsnoop"
)

// Shift shifts req.URL.Path forward by one segment, and returns the segment,
// if any. req.URL.Path must either be empty, or have a "/" prefix.
//
// For the paths "" and "/", Shift is a no-op and returns "".
//
// For the path "/abc", Shift sets req.URL.Path to "" and returns "abc".
//
// For the path "/abc/anything", Shift sets req.URL.Path to "/anything",
// and returns "abc".
func Shift(req *http.Request) string {
	seg, rest := shift(req.URL.Path)
	req.URL.Path = rest
	return seg
}

func shift(path string) (seg string, rest string) {
	if path == "" || path == "/" {
		return "", path
	}

	path = path[1:]
	idx := strings.IndexByte(path, '/')
	if idx != -1 {
		seg = path[:idx]
		rest = path[idx:]
		return seg, rest
	} else {
		return path, ""
	}
}

type key int

const (
	pathKey      key = 0
	requestIDKey key = 1
)

// WithPath stores req.URL.Path in the context associated with req, and
// returns the new *http.Request, with the updated context.
//
// The value can later be retrieved by calling Path on the request.
//
// If the request context stores a path already, WithPath is a no-op
// and returns req.
func WithPath(req *http.Request) *http.Request {
	ctx := req.Context()
	if ctx.Value(pathKey) != nil {
		return req
	}
	vctx := context.WithValue(ctx, pathKey, req.URL.Path)
	return req.WithContext(vctx)
}

// Path returns the original URL.Path associated with req. If the context
// associated with req does not store a path, Path returns the empty string.
func Path(req *http.Request) string {
	val := req.Context().Value(pathKey)
	if val == nil {
		return ""
	}
	return val.(string)
}

// WithRequestID assigns an identifier to an HTTP request, if one is not
// assigned already.
func WithRequestID(req *http.Request, id string) *http.Request {
	ctx := req.Context()
	val := ctx.Value(requestIDKey)
	if val != nil {
		return req
	}
	return req.WithContext(context.WithValue(ctx, requestIDKey, id))
}

// RequestID returns the identifier associated with the request.
func RequestID(req *http.Request) string {
	val := req.Context().Value(requestIDKey)
	if val == nil {
		return ""
	}
	return val.(string)
}

// RequestLogger returns a logger scoped to the specified request. The logger
// records the "method", "path", "remote_addr" and "user_agent" keys. If present,
// it also records the "request_id" key.
func RequestLogger(base *log.Logger, req *http.Request) *log.Logger {
	kv := log.KV{
		"method":      req.Method,
		"path":        Path(req),
		"remote_addr": req.RemoteAddr,
	}
	if ua := req.UserAgent(); ua != "" {
		kv["user_agent"] = ua
	}
	if id := RequestID(req); id != "" {
		kv["request_id"] = id
	}
	return base.WithKV(kv)
}

// ServeInstrumented instruments w, wraps h, and calls the wrapped handler
// with the instrumented http.ResponseWriter and the specified *http.Request.
// It returns a summary of the request.
func ServeInstrumented(h http.Handler, w http.ResponseWriter, req *http.Request) Summary {
	m := httpsnoop.CaptureMetrics(h, w, req)
	return Summary{
		Status:   m.Code,
		Duration: m.Duration,
		Written:  m.Written,
	}
}

// Summary is a summary of an HTTP server response.
type Summary struct {
	// Status is the first HTTP status code written, or http.StatusOK
	// if no status was written explicitly.
	Status int

	// Duration measures the duration of the request.
	Duration time.Duration

	// Written typically counts the number of bytes written to the HTTP
	// response body.
	Written int64
}

// KV returns key-value pairs representing the Summary, suitable for logging
// using a acln.ro/log.Logger. The "status", "duration" and "written" keys
// are used.
func (s Summary) KV() log.KV {
	return log.KV{
		"status":   s.Status,
		"duration": s.Duration,
		"written":  s.Written,
	}
}
