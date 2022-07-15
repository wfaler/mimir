// SPDX-License-Identifier: AGPL-3.0-only

package push

import (
	"github.com/grafana/mimir/pkg/mimirpb"
)

type supplierFunc func() (*mimirpb.WriteRequest, func(), error)

// Request represents a push request. It allows lazy body reading from the underlying http request
// and adding cleanups that should be done after the request is completed.
type Request struct {
	cleanups []func()

	getRequest supplierFunc

	supplied bool
	request  *mimirpb.WriteRequest
	err      error
}

func newRequest(p supplierFunc) *Request {
	return &Request{
		cleanups:   make([]func(), 0, 10),
		getRequest: p,
	}
}

func NewParsedRequest(r *mimirpb.WriteRequest) *Request {
	return newRequest(func() (*mimirpb.WriteRequest, func(), error) {
		return r, nil, nil
	})
}

// WriteRequest returns request from supplier function. Function is only called once,
// and subsequent calls to WriteRequest return the same value.
func (r *Request) WriteRequest() (*mimirpb.WriteRequest, error) {
	if !r.supplied {
		var cleanup func()
		r.request, cleanup, r.err = r.getRequest()
		if cleanup != nil {
			r.AddCleanup(cleanup)
		}
		r.supplied = true
	}
	return r.request, r.err
}

// AddCleanup adds a function that will be called once CleanUp is called.
func (r *Request) AddCleanup(f func()) {
	r.cleanups = append(r.cleanups, f)
}

// CleanUp calls all added cleanups.
func (r *Request) CleanUp() {
	for _, f := range r.cleanups {
		f()
	}
}
