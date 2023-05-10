// httphandlers_test.go
package main

import (
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestHealthzHandler(t *testing.T) {
	req, err := http.NewRequest("GET", "/healthz", nil)
	if err != nil {
		t.Fatal(err)
	}

	// test HTTP status OK
	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(healthzHandler)
	handler.ServeHTTP(rr, req)
	if !assert.Equal(t, http.StatusOK, rr.Code) {
		t.Errorf("handler returned wrong status code: got %v want %v",
			rr.Code, http.StatusOK)
	}

	// test HTTP body
	expected := `{"alive": true}`
	if !assert.Equal(t, expected, rr.Body.String()) {
		t.Errorf("handler returned unexpected body: got %v want %v",
			rr.Body.String(), expected)
	}
}

func TestReadyzHandler(t *testing.T) {
	req, err := http.NewRequest("GET", "/readyz", nil)
	if err != nil {
		t.Fatal(err)
	}

	ready := &atomic.Value{}
	ready.Store(true)
	notReady := &atomic.Value{}
	notReady.Store(false)

	tests := []struct {
		name           string
		isReady        *atomic.Value
		expectedStatus int
		expectedBody   string
	}{
		{
			name:           "test_503",
			isReady:        notReady,
			expectedStatus: 503,
			expectedBody:   "Service Unavailable\n",
		},
		{
			name:           "test_200",
			isReady:        ready,
			expectedStatus: 200,
			expectedBody:   `{"ready": true}`,
		},
	}

	// test HTTP status OK
	for _, tt := range tests {
		rr := httptest.NewRecorder()
		handler := http.HandlerFunc(readyzHandler(tt.isReady))
		handler.ServeHTTP(rr, req)
		if !assert.Equal(t, tt.expectedStatus, rr.Code) {
			t.Errorf("handler returned wrong status code: got %v want %v",
				rr.Code, tt.expectedStatus)
		}
	}

	// test HTTP body
	for _, tt := range tests {
		rr := httptest.NewRecorder()
		handler := http.HandlerFunc(readyzHandler(tt.isReady))
		handler.ServeHTTP(rr, req)
		if !assert.Equal(t, tt.expectedBody, rr.Body.String()) {
			t.Errorf("handler returned unexpected body: got %v want %v",
				rr.Body.String(), tt.expectedBody)
		}
	}
}
