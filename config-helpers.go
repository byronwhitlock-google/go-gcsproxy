package main

import (
	"os"
	"strconv"
	"strings"

	log "github.com/sirupsen/logrus"
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

// Parsing the "*:global-key" or "bucket/path:project/key,bucket2:key2" but the global key overrides all the other keys
func bucketKeyMappings(bucketKeyMapString string) map[string]string {

	if bucketKeyMapString==""{
		log.Debug("No Bucket Key Mapping given")
		return nil
	}

	bucketKeyMap := make(map[string]string)
	bucketKeys := strings.Split(bucketKeyMapString, ",")
	for i := 0; i < len(bucketKeys); i++ {
		
		bucketKeyArray := strings.Split(bucketKeys[i], ":")
		bucketKeyMap[bucketKeyArray[0]]=bucketKeyArray[1]
	}
	
	log.Debugf("BucketkeyMapping: %v", bucketKeyMap)
	return bucketKeyMap

}
