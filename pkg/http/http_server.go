package http

import (
	"encoding/json"
	log "github.com/sirupsen/logrus"
	"github.com/yiranone/chain-api-proxy/pkg/bean"
	cache2 "github.com/yiranone/chain-api-proxy/pkg/cache"
	"github.com/yiranone/chain-api-proxy/pkg/config"
	util2 "github.com/yiranone/chain-api-proxy/pkg/util"
	"io/ioutil"
	"net/http"
	"sync"
	"time"
)

func Handler(w http.ResponseWriter, r *http.Request, cfg *config.Config, cache *cache2.Cache,
	requestBlockNumberChain chan bean.GenericJSON,
	requestContexts map[string]map[string]*bean.RequestContext,
	requestContextsM sync.Mutex) {
	defer r.Body.Close()
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		log.Errorf("客户端解析body异常 err:%v", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	tid := util2.GenerateTraceID()
	log := log.WithFields(log.Fields{
		"source": "客户端",
		"tid":    tid,
	})

	var req bean.GenericJSON
	if err := json.Unmarshal(body, &req); err != nil {
		log.Errorf("客户端解析json异常body: %.200s err:%v", body, err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	startTime := time.Now()
	_, err = util2.ConvertToString(req["method"])
	if err != nil {
		log.Errorf("获取http请求的method异常了 %v, %v", req["method"], err)
		return
	}
	id, err := util2.ConvertToInt64(req["id"])
	if err != nil {
		log.Errorf("获取http请求的id异常了 %v, %v", req["id"], err)
		return
	}
	cacheKey := cache2.CreateCacheKey(req)
	log = log.WithField("clientReqId", id)
	log = log.WithField("cacheKey", util2.TruncateString(cacheKey, 128))
	log = log.WithField("blockNumberChainSize", len(requestBlockNumberChain))

	log.Printf("收到客户端请求 body: %.200s", body)

	// Check if response is in cache
	if cachedResponse, exists := cache.Get(cacheKey); exists {
		responseJSON, err := json.Marshal(cachedResponse)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		duration := time.Since(startTime).Seconds()
		log.Printf("***** 请求命中缓存，应答客户端: %s duration:%f, response: %.100s", cacheKey, duration, string(responseJSON))
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write(responseJSON)
		return
	}

	// Create RequestContext for waiting responses
	respChan := make(chan bean.GenericJSON, 1)
	ctx := &bean.RequestContext{
		CacheKey: cacheKey,
		Response: respChan,
		ID:       id,
		Tid:      tid,
	}

	requestContextsM.Lock()
	if _, exists := requestContexts[cacheKey]; !exists {
		requestContexts[cacheKey] = make(map[string]*bean.RequestContext)
	}
	requestContexts[cacheKey][tid] = ctx
	requestContextsM.Unlock()

	defer func() {
		requestContextsM.Lock()
		defer requestContextsM.Unlock()
		if ctxs, exists := requestContexts[cacheKey]; exists {
			delete(ctxs, tid)
			if len(ctxs) == 0 {
				delete(requestContexts, cacheKey)
			}
		}
	}()

	// Push the incoming request to the channel for further processing
	if cache.GetAndSetProcessing(cacheKey) {
		log.Printf(">>>> 客户端发送到chain %s id=%d requestBlockNumberChainSize%d", cacheKey, id, len(requestBlockNumberChain))
		req["source"] = "客户端"
		req["tid"] = tid
		req["clientReqId"] = id
		req["sendTime"] = time.Now()
		requestBlockNumberChain <- req
	}

	select {
	case res := <-ctx.Response:

		newRes, err := cache2.DeepCopy(res)
		if err != nil {
			log.Errorf("深度copy失败 err:%v", err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		newRes["id"] = ctx.ID
		responseJSON, err := json.Marshal(newRes)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		duration := time.Since(startTime).Seconds()

		log.Printf("***** 应答客户端，渠道触发 cacheKey=%s, ID:%d, 耗时:%.2fs response: %.100s", cacheKey, ctx.ID, duration, responseJSON)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write(responseJSON)
	case <-time.After(time.Duration(cfg.FrontHttpSeconds) * time.Second):
		duration := time.Since(startTime).Seconds()

		log.Printf("客户端请求超时了 for key: %s, ID:%d duration %f", cacheKey, ctx.ID, duration)
		if cachedResponse, exists := cache.Get(cacheKey); exists {
			//超时了判断缓存有没有
			responseJSON, err := json.Marshal(cachedResponse)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}

			duration := time.Since(startTime).Seconds()
			log.Printf("***** 客户端请求超时后命中缓存，应答客户端: %s 耗时:%.2fs, response: %.100s", cacheKey, duration, responseJSON)
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			w.Write(responseJSON)
			return
		}
		//http.Error(w, "Request timed out99 ", http.StatusGatewayTimeout)
		w.WriteHeader(http.StatusGatewayTimeout)
		w.Write([]byte("Request timed out"))
		return
	}
}
