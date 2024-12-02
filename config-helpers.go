package main

import (
	"os"
	"strconv"
)

func getStringEnvVar(key string) (string, bool) {
	envVar := os.Getenv(key)
	if len(envVar) == 0 {
		return "", false
	}
	return envVar, true
}

func getBoolEnvVar(key string) (bool, bool) {
	envVar, boolError := strconv.ParseBool(os.Getenv(key))
	if boolError == nil {
		return envVar, true
	}
	return false, false
}

func getIntEnvVar(key string) (int, bool) {
	envVar, intError := strconv.Atoi(os.Getenv(key))
	if intError == nil {
		return envVar, true
	}
	return 0, false
}
