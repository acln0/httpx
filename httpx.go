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

const pathKey key = 0

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
