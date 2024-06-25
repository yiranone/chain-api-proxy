package util

import (
	"fmt"
	"math/rand"
	"sync/atomic"
)

var requestId = int32(1)

func GenerateTraceID() string {
	return fmt.Sprintf("%d", rand.Int63())
}

func NextRequestId() int32 {
	return atomic.AddInt32(&requestId, 1)
}
