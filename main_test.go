package main

import (
	"fmt"
	"testing"
)

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
