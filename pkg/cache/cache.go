package cache

import (
	"encoding/json"
	log "github.com/sirupsen/logrus"
	"github.com/yiranone/chain-api-proxy/pkg/bean"
	"sync"
	"time"
)

type CacheItem struct {
	Value      bean.GenericJSON
	Stage      int // 0初始化， 1 处理中 2 已完成
	StoredTime time.Time
}

// Cache stores the response cacheKey to prevent duplicate requests
type Cache struct {
	Data  map[string]CacheItem
	Mutex sync.RWMutex
}

// NewCache creates a new Cache instance
func NewCache() *Cache {
	cache := &Cache{
		Data: make(map[string]CacheItem),
	}
	return cache
}

func (c *Cache) GetAndSetProcessing(key string) bool {
	c.Mutex.Lock()
	defer c.Mutex.Unlock()
	item, exists := c.Data[key]
	if !exists {
		c.Data[key] = CacheItem{Stage: 1}
		return true
	}
	// 如果 key 存在并且没有被处理
	if item.Stage == 0 {
		c.Data[key] = CacheItem{Stage: 1}
		return true
	}
	return false
}
func (c *Cache) Delete(key string) {
	c.Mutex.Lock()
	defer c.Mutex.Unlock()
	if c.Data[key].Stage != 2 { //已经拿到数据了不删除
		delete(c.Data, key)
	}
}

// Add adds a response to the cache with a timestamp
func (c *Cache) Add(key string, value bean.GenericJSON) {
	c.Mutex.Lock()
	defer c.Mutex.Unlock()
	c.Data[key] = CacheItem{
		Value:      value,
		Stage:      2,
		StoredTime: time.Now(),
	}
}

// DeepCopy 进行深度复制
func DeepCopy(src bean.GenericJSON) (bean.GenericJSON, error) {
	var dst bean.GenericJSON
	bytes, err := json.Marshal(src)
	if err != nil {
		return nil, err
	}
	err = json.Unmarshal(bytes, &dst)
	if err != nil {
		return nil, err
	}
	return dst, nil
}

// Get retrieves a response from the cache
func (c *Cache) Get(key string) (bean.GenericJSON, bool) {
	c.Mutex.RLock()
	defer c.Mutex.RUnlock()
	item, exists := c.Data[key]
	if !exists {
		return nil, false
	}
	if item.Stage != 2 {
		return nil, false
	}
	//if item.Value != nil {
	//	return item.Value, true
	//}
	return item.Value, true
}

func (c *Cache) CleanupExpiredItems(seconds int) {
	log.Infof("启动CleanupExpiredItems 缓存保留时间%ds", seconds)
	for {
		time.Sleep(20 * time.Second) // Check for expired items every minute
		c.Mutex.Lock()
		log.Infof("启动CleanupExpiredItems 缓存保留时间%ds", seconds)
		for key, item := range c.Data {
			if time.Since(item.StoredTime) > time.Duration(seconds)*time.Second {
				log.Infof("清理缓存 key=%s StoreTime:%s", key, item.StoredTime.Format(time.DateTime))
				delete(c.Data, key)
			}
		}
		c.Mutex.Unlock()
	}
}

func CreateCacheKey(payload bean.GenericJSON) string {
	key := payload["method"].(string)
	if params, ok := payload["params"].([]interface{}); ok && len(params) > 0 {
		paramsJSON, _ := json.Marshal(params)
		key += "_" + string(paramsJSON)
	}
	return key
}
