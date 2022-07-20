package push

import (
	"context"
	"net/http"
	"sync"

	"github.com/grafana/mimir/pkg/mimirpb"
)

type Request struct {
	Cleanups []func()
	Context  context.Context

	httpReq                      *http.Request
	maxMessageSize               int
	allowSkipLabelNameValidation bool
	parser                       ParserFunc
	parseRequestOnce             sync.Once
	parsedRequest                *mimirpb.WriteRequest
	parseError                   error
}

func (r *Request) parseRequest() {
	bufHolder := bufferPool.Get().(*bufHolder)
	var req mimirpb.PreallocWriteRequest
	buf, err := r.parser(r.Context, r.httpReq, r.maxMessageSize, bufHolder.buf, &req)
	if err != nil {
		//level.Error(logger).Log("err", err.Error())
		//http.Error(w, err.Error(), http.StatusBadRequest)
		bufferPool.Put(bufHolder)
		r.parseError = err
		return
	}
	// If decoding allocated a bigger buffer, put that one back in the pool.
	if len(buf) > len(bufHolder.buf) {
		bufHolder.buf = buf
	}

	r.Cleanups = append(r.Cleanups, func() {
		mimirpb.ReuseSlice(req.Timeseries)
		bufferPool.Put(bufHolder)
	})

	if r.allowSkipLabelNameValidation {
		req.SkipLabelNameValidation = req.SkipLabelNameValidation && r.httpReq.Header.Get(SkipLabelNameValidationHeader) == "true"
	} else {
		req.SkipLabelNameValidation = false
	}

	if req.Source == 0 {
		req.Source = mimirpb.API
	}

	r.parsedRequest = &req.WriteRequest
}

func (r *Request) WriteRequest() (*mimirpb.WriteRequest, error) {
	r.parseRequestOnce.Do(r.parseRequest)
	return r.parsedRequest, r.parseError
}

func (r *Request) CleanUp() {
	for _, f := range r.Cleanups {
		f()
	}
}

func NewParsedRequest(r *mimirpb.WriteRequest) *Request {
	req := &Request{parsedRequest: r}
	req.parseRequestOnce.Do(func() {}) // no need to parse anything, we have the request
	return req
}
