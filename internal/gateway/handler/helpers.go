package handler

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
)

func NewTestRequest(method, url string, body interface{}) *http.Request {
	var req *http.Request
	if body != nil {
		jsonBody, _ := json.Marshal(body)
		req = httptest.NewRequest(method, url, strings.NewReader(string(jsonBody)))
	} else {
		req = httptest.NewRequest(method, url, nil)
	}
	req.Header.Set("Content-Type", "application/json")
	return req
}

func ContextWithUser(r *http.Request, userID string) *http.Request {
	return r.WithContext(context.WithValue(r.Context(), "userID", userID))
}
