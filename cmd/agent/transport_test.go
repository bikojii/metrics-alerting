package main

import (
	"net/http"
	"net/http/httptest"
)

type handlerTransport struct{ handler http.Handler }

func (tr handlerTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	w := httptest.NewRecorder()
	tr.handler.ServeHTTP(w, req)
	if err := req.Context().Err(); err != nil {
		return nil, err
	}
	response := w.Result()
	response.Request = req
	return response, nil
}
