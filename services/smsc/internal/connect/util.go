package connect

import "time"

func MillisToDuration(millis int64) time.Duration {
	return time.Duration(millis * int64(time.Nanosecond))
}
