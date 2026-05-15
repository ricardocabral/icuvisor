//go:build linux

package credstore

import (
	"errors"
	"strings"

	dbus "github.com/godbus/dbus/v5"
)

func isKeychainUnavailable(err error) bool {
	if err == nil {
		return false
	}
	var dbusErr dbus.Error
	if errors.As(err, &dbusErr) {
		switch dbusErr.Name {
		case "org.freedesktop.DBus.Error.ServiceUnknown",
			"org.freedesktop.DBus.Error.NameHasNoOwner",
			"org.freedesktop.DBus.Error.NoServer",
			"org.freedesktop.DBus.Error.FileNotFound",
			"org.freedesktop.Secret.Error.NoSuchObject":
			return true
		case "org.freedesktop.DBus.Error.AccessDenied",
			"org.freedesktop.Secret.Error.IsLocked",
			"org.freedesktop.Secret.Error.NoSession":
			return false
		}
	}

	message := strings.ToLower(err.Error())
	unavailableFragments := []string{
		"couldn't determine address of session bus",
		"cannot autolaunch dbus without x11 $display",
		"dbus-launch",
		"no such file or directory",
		"the name org.freedesktop.secrets was not provided",
		"service org.freedesktop.secrets",
		"object does not exist at path \"/org/freedesktop/secrets/aliases/default\"",
		"object does not exist at path '/org/freedesktop/secrets/aliases/default'",
		"path not found",
	}
	for _, fragment := range unavailableFragments {
		if strings.Contains(message, fragment) {
			return true
		}
	}
	return false
}
