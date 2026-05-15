//go:build !linux

package credstore

import (
	"errors"
	"testing"
)

func TestIsKeychainUnavailableIsFalseOffLinux(t *testing.T) {
	t.Parallel()

	if isKeychainUnavailable(errors.New("dbus: couldn't determine address of session bus")) {
		t.Fatal("isKeychainUnavailable() = true off Linux, want false")
	}
}
