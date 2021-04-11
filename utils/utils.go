package utils

import (
	"fmt"
	"os"
	"strings"
)

// pass slice of strings for namespaces
func GetEnvAsSlice(name string, defaultVal []string, sep string) []string {
	valStr := GetDenyListEnv(name, "")
	if valStr == "" {
		return defaultVal
	}
	// split on ","
	val := strings.Split(valStr, sep)
	return val
}

// lookup DENY_LIST, default is nil
func GetDenyListEnv(key string, defaultVal string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return defaultVal
}

func ContainsString(slice []string, s string) bool {
	for _, item := range slice {
		if item == s {
			return true
		}
	}
	return false
}

func RemoveString(slice []string, s string) (result []string) {
	for _, item := range slice {
		if item == s {
			continue
		}
		result = append(result, item)
	}
	return
}

func GetWatchNamespace() (string, error) {
	ns, found := os.LookupEnv("WATCH_NAMESPACE")
	if !found {
		return "", fmt.Errorf("%s must be set", "WATCH_NAMESPACE")
	}
	return ns, nil
}
