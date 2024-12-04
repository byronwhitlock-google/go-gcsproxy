package main

import (
	"os"
	"strconv"
)

func envConfigStringWithDefault(key string, defValue string) string {
	envVar := os.Getenv(key)
	if len(envVar) == 0 {
		return defValue
	}
	return envVar
}

func envConfigBoolWithDefault(key string, defValue bool) bool {
	envVar, boolError := strconv.ParseBool(os.Getenv(key))
	if boolError == nil {
		return envVar
	}
	return defValue
}

func envConfigIntWithDefault(key string, defValue int) int {
	envVar, intError := strconv.Atoi(os.Getenv(key))
	if intError == nil {
		return envVar
	}
	return defValue
}
