package main

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/brutella/hap"
	"github.com/brutella/hap/accessory"
)

func TestHTTPControlRemoved(t *testing.T) {
	a := accessory.NewGarageDoorOpener(accessory.Info{Name: "test"})
	server, err := newServer(hap.NewFsStore(t.TempDir()), a)
	if err != nil {
		t.Fatal(err)
	}
	handler := server.ServeMux().(http.Handler)
	for _, path := range []string{"/open", "/close"} {
		for _, method := range []string{http.MethodGet, http.MethodPost} {
			t.Run(method+path, func(t *testing.T) {
				response := httptest.NewRecorder()
				handler.ServeHTTP(response, httptest.NewRequest(method, path, nil))
				if response.Code != http.StatusNotFound {
					t.Fatalf("got status %d, want 404: %s", response.Code, response.Body)
				}
			})
		}
	}
	response := httptest.NewRecorder()
	handler.ServeHTTP(response, httptest.NewRequest(http.MethodGet, "/", nil))
	if response.Code != http.StatusOK {
		t.Fatalf("got status %d, want 200: %s", response.Code, response.Body)
	}
	for _, path := range []string{"/open", "/close"} {
		if strings.Contains(response.Body.String(), `"`+path+`"`) {
			t.Errorf("status page still links to removed control %s", path)
		}
	}
}

func TestAdjustState(t *testing.T) {
	for _, tt := range []struct {
		normally string
		state    string
		expected string
	}{
		{normallyClosed, stateOpen, stateOpen},
		{normallyClosed, stateClosed, stateClosed},
		{normallyOpen, stateOpen, stateClosed},
		{normallyOpen, stateClosed, stateOpen},
	} {
		t.Run(fmt.Sprintf("normally=%s state=%s", tt.normally, tt.state), func(t *testing.T) {
			got := adjustState(tt.state, tt.normally)
			if got != tt.expected {
				t.Errorf("got %q, want %q", got, tt.expected)
			}
		})
	}
}
