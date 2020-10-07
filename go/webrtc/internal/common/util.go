package common

import (
	"strconv"
	"time"
)

func UniquePeerID() string {
	return strconv.Itoa(int(time.Now().UnixNano() / 1e6))
}

