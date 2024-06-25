package bean

import (
	"time"
)

type GenericJSON map[string]interface{}

type RequestContext struct {
	CacheKey string
	Response chan GenericJSON
	ID       int64
	Tid      string
	SendTime time.Time
}
