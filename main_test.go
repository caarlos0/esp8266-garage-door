package main

import (
	"fmt"
	"testing"

	"github.com/caarlos0/env/v11"
)

func TestPairingPINConfig(t *testing.T) {
	for _, tt := range []struct {
		name string
		env  map[string]string
		fail bool
	}{
		{"missing", map[string]string{}, true},
		{"empty", map[string]string{"HOMEKIT_PIN": ""}, true},
		{"library default", map[string]string{"HOMEKIT_PIN": "00102003"}, true},
		{"insecure", map[string]string{"HOMEKIT_PIN": "11111111"}, true},
		{"short", map[string]string{"HOMEKIT_PIN": "13579"}, true},
		{"letters", map[string]string{"HOMEKIT_PIN": "abcdefgh"}, true},
		{"valid", map[string]string{"HOMEKIT_PIN": "59318426"}, false},
	} {
		t.Run(tt.name, func(t *testing.T) {
			cfg, err := env.ParseAsWithOptions[Config](env.Options{Environment: tt.env})
			if err == nil {
				err = cfg.validate()
			}
			if (err != nil) != tt.fail {
				t.Fatalf("got error %v, want failure %v", err, tt.fail)
			}
		})
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
