package shared

import "os"

// LookupConfigValueWithFallback returns a config value of the given key if not found it uses the fallback value
func LookupConfigValueWithFallback(key, fallback string) (value string) {
	if value, found := os.LookupEnv(key); found {
		return value
	}
	return fallback
}

// LookupConfigValue returns a config value of the given key, found is false if key does not exist
func LookupConfigValue(key string) (value string, found bool) {
	return os.LookupEnv(key)
}
