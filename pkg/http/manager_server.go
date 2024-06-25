package http

import (
	"encoding/json"
	cache2 "github.com/yiranone/chain-api-proxy/pkg/cache"
	url2 "github.com/yiranone/chain-api-proxy/pkg/url"
	"net/http"
)

func HandlerUrlManagerStatusRequest(w http.ResponseWriter, r *http.Request, m *url2.URLManager) {
	statuses, err := m.GetAllURLStatus()
	if err != nil {
		http.Error(w, "Failed to get URL statuses", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(statuses)
}

func HandlerCacheStatusRequest(w http.ResponseWriter, r *http.Request, cache *cache2.Cache) {
	cacheKey := r.URL.Query().Get("cacheKey")
	if cacheKey != "" {
		cacheData, ok := cache.Get(cacheKey)
		if ok {
			w.Header().Set("Content-Type", "application/json")
			responseJSON, err := json.Marshal(cacheData)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			w.Write(responseJSON)
		} else {
			w.Write([]byte("缓存不存在"))
		}
	}
}
