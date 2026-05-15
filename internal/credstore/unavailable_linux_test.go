//go:build linux

package credstore

import (
	"context"
	"errors"
	"testing"

	dbus "github.com/godbus/dbus/v5"
)

func TestIsKeychainUnavailableOnLinux(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		err  error
		want bool
	}{
		{name: "nil", err: nil, want: false},
		{name: "no session bus", err: errors.New("dbus: couldn't determine address of session bus"), want: true},
		{name: "autolaunch", err: errors.New("Cannot autolaunch D-Bus without X11 $DISPLAY"), want: true},
		{name: "secret service unknown", err: dbus.Error{Name: "org.freedesktop.DBus.Error.ServiceUnknown"}, want: true},
		{name: "default collection missing", err: errors.New("Object does not exist at path \"/org/freedesktop/secrets/aliases/default\""), want: true},
		{name: "path not found", err: errors.New("path not found"), want: true},
		{name: "access denied stays visible", err: dbus.Error{Name: "org.freedesktop.DBus.Error.AccessDenied"}, want: false},
		{name: "locked stays visible", err: dbus.Error{Name: "org.freedesktop.Secret.Error.IsLocked"}, want: false},
		{name: "unexpected stays visible", err: errors.New("malformed D-Bus response"), want: false},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			if got := isKeychainUnavailable(tc.err); got != tc.want {
				t.Fatalf("isKeychainUnavailable(%v) = %v, want %v", tc.err, got, tc.want)
			}
		})
	}
}

func TestGetMapsUnavailableKeychainToNotFoundOnLinux(t *testing.T) {
	t.Parallel()

	store := keyringStore{backend: &fakeKeyringBackend{getErr: errors.New("dbus: couldn't determine address of session bus")}}
	_, err := store.Get(context.Background(), IntervalsAPIKeyAccount)
	if !errors.Is(err, ErrNotFound) {
		t.Fatalf("Get() error = %v, want ErrNotFound", err)
	}
}
