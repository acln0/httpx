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

package httpx_test

import (
	"net/http"
	"net/url"
	"testing"

	"acln.ro/httpx"
)

func TestShift(t *testing.T) {
	tests := []struct {
		path string
		seg  string
		rest string
	}{
		{"", "", ""},
		{"/", "", "/"},
		{"/xyz", "xyz", ""},
		{"/xyz/", "xyz", "/"},
		{"/abc/xyz", "abc", "/xyz"},
		{"/abc/xyz/", "abc", "/xyz/"},
		{"/abc/xyz/t", "abc", "/xyz/t"},
	}
	for _, tt := range tests {
		u := &url.URL{
			Scheme: "http",
			Host:   "example.com",
			Path:   tt.path,
		}
		req, err := http.NewRequest("", u.String(), nil)
		if err != nil {
			t.Fatalf("%q: %v", tt.path, err)
		}
		seg := httpx.Shift(req)
		rest := req.URL.Path
		if seg != tt.seg {
			t.Errorf("shift(%q): seg == %q, want %q",
				tt.path, seg, tt.seg)
		}

		if rest != tt.rest {
			t.Errorf("shift(%q): req.URL.Path == %q, want %q",
				tt.path, rest, tt.rest)
		}
	}
}

func TestContext(t *testing.T) {
	path := "/abc/xyz"
	u := &url.URL{
		Scheme: "http",
		Host:   "example.com",
		Path:   path,
	}
	req, err := http.NewRequest("", u.String(), nil)
	if err != nil {
		t.Fatal(err)
	}

	if p := httpx.Path(req); p != "" {
		t.Fatalf("Path: got %q on request with no path key", p)
	}

	req = httpx.WithPath(req)
	seg := httpx.Shift(req)
	if seg != "abc" {
		t.Fatalf("shift(%q): seg == %q, want %q", path, seg, "abc")
	}

	old := req
	new := httpx.WithPath(old)
	if old != new {
		t.Fatalf("different requests with path key present")
	}

	if p := httpx.Path(req); p != path {
		t.Fatalf("Path after Shift returned %q, want %q", p, path)
	}
}
