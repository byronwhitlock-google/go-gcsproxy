package main

import (
	"os"
	"strconv"
)

func setStringEnvVar(key string, defValue *string) int {
	envVar := os.Getenv(key)
	if len(envVar) == 0 {
		return 0
	}
	*defValue = envVar
	return 0
}

func setBoolEnvVar(key string, defValue *bool) int {
	envVar, boolError := strconv.ParseBool(os.Getenv(key))
	if boolError == nil {
		*defValue = envVar
		return 0
	}
	return 0
}

func setIntEnvVar(key string, defValue *int) int {
	envVar, intError := strconv.Atoi(os.Getenv(key))
	if intError == nil {
		*defValue = envVar
		return 0
	}
	return 0
}
