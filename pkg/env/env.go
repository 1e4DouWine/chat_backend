package env

import (
	"os"
)

const (
	envKey      = "ENV"
	prodEnv     = "prod"
	testEnv     = "test"
	localDevEnv = ""
)

func GetEnv() string {
	return os.Getenv(envKey)
}

func IsProd() bool {
	return GetEnv() == prodEnv
}

func IsTest() bool {
	return GetEnv() == testEnv
}

func IsLocalDev() bool {
	return GetEnv() == localDevEnv
}
