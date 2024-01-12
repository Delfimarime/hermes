package common

import "time"

func MillisToDuration(millis int64) time.Duration {
	return time.Duration(millis * int64(time.Nanosecond))
}

func ToStrPointer(value string) *string {
	return &value
}

func ToIntPointer(value int) *int {
	return &value
}
