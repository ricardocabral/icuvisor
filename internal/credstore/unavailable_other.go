//go:build !linux

package credstore

func isKeychainUnavailable(error) bool {
	return false
}
