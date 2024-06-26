package pull

import (
	"bytes"
	"encoding/json"
	"fmt"
	log "github.com/sirupsen/logrus"
	"github.com/yiranone/chain-api-proxy/pkg/bean"
	cache2 "github.com/yiranone/chain-api-proxy/pkg/cache"
	"github.com/yiranone/chain-api-proxy/pkg/config"
	url2 "github.com/yiranone/chain-api-proxy/pkg/url"
	"github.com/yiranone/chain-api-proxy/pkg/util"
	"io"
	"io/ioutil"
	"net/http"
	"runtime"
	"runtime/debug"
	"strconv"
	"strings"
	"sync"
	"time"
)

var latestBlockNumberChain = make(chan int, 100)

func CalcLatestBlockNumber(config *config.Config, urlManager *url2.URLManager, cache *cache2.Cache,
	requestBlockNumberChain chan bean.GenericJSON) {
	if config.LoopSeconds <= 0 {
		log.Infof("不启动calcLatestBlockNumber")
	}
	defer func() {
		if err := recover(); err != nil {
			log.Println("calcLatestBlockNumber error 异常退出了:", err)
			log.Printf("*************************************recover begin")
			log.Printf("Program crashed with error: %v\n", err)
			buf := make([]byte, 1<<16)
			stackSize := runtime.Stack(buf, true)
			log.Printf("Stack trace:\n%s\n", buf[:stackSize])
			log.Printf("Stack trace:\n%s\n", debug.Stack())
			log.Printf("*************************************recover end")
		}
	}()
	currentBlockNumberDec := 0
	for newBlockNumberDec := range latestBlockNumberChain {
		log := log.WithField("latestSequenceChainSize", len(latestBlockNumberChain))
		newBlockNumberHex := fmt.Sprintf("0x%x", newBlockNumberDec)
		log.Printf("收到最新区块通知%d,%s", newBlockNumberDec, newBlockNumberHex)
		if currentBlockNumberDec == 0 {
			currentBlockNumberDec = newBlockNumberDec
		}
		for currentBlockNumberDec <= newBlockNumberDec {
			currentBlockNumberHex := fmt.Sprintf("0x%x", currentBlockNumberDec)
			tid := util.GenerateTraceID()
			log := log.WithField("tid", tid).WithField("block", currentBlockNumberDec).WithField("blockHex", currentBlockNumberHex)
			log.Infof("开始处理区块信息 %d,%s", currentBlockNumberDec, currentBlockNumberHex)
			startSendChainTime := time.Now()
			var sendCost1, sendCost2, sendCost3 float64
			var lockCost1, lockCost2, lockCost3 float64
			//只有fantom 和 ethereum 有 trace_block
			if config.Chain == 250 || config.Chain == 1 {
				url, _ := urlManager.GetRandomURL(url2.JobRequestSpecialUrl)
				traceBlockjson := bean.GenericJSON{
					"method":   "trace_block",
					"params":   []interface{}{currentBlockNumberHex},
					"source":   "获取到区块触发1",
					"tid":      tid,
					"sendTime": time.Now(),
					"url":      url,
				}
				traceBlockjsoncachekey := cache2.CreateCacheKey(traceBlockjson)
				if cache.GetAndSetProcessing(traceBlockjsoncachekey) {
					lockCost1 = time.Since(startSendChainTime).Seconds()
					//log.Printf(">>>> 定时任务发送traceBlockjsoncachekey到channel %s", traceBlockjsoncachekey)
					requestBlockNumberChain <- traceBlockjson
					traceBlockjson := bean.GenericJSON{
						"method":   "trace_block",
						"params":   []interface{}{currentBlockNumberHex},
						"source":   "获取到区块触发2",
						"tid":      tid,
						"sendTime": time.Now(),
						"url":      url,
					}
					requestBlockNumberChain <- traceBlockjson
					sendCost1 = time.Since(startSendChainTime).Seconds()
				}
			}

			url, _ := urlManager.GetRandomURL(url2.JobRequestUrl)
			ethGetLogsjson := bean.GenericJSON{
				"method": "eth_getLogs",
				"params": []interface{}{
					map[string]interface{}{"fromBlock": currentBlockNumberHex, "toBlock": currentBlockNumberHex},
				},
				"source":   "获取到区块触发",
				"tid":      tid,
				"url":      url,
				"sendTime": time.Now(),
			}
			startSendChainTime = time.Now()
			ethGetLogsjsoncachekey := cache2.CreateCacheKey(ethGetLogsjson)
			if cache.GetAndSetProcessing(ethGetLogsjsoncachekey) {
				lockCost2 = time.Since(startSendChainTime).Seconds()
				//log.Printf(">>>> 定时任务发送ethGetLogsjsoncachekey到channel %s", ethGetLogsjsoncachekey)
				requestBlockNumberChain <- ethGetLogsjson
			}
			sendCost2 = time.Since(startSendChainTime).Seconds()

			getBlockByNumberjson := bean.GenericJSON{
				"method":   "eth_getBlockByNumber",
				"params":   []interface{}{currentBlockNumberHex, true},
				"source":   "获取到区块触发",
				"tid":      tid,
				"url":      url,
				"sendTime": time.Now(),
			}
			startSendChainTime = time.Now()
			getBlockByNumbercachekey := cache2.CreateCacheKey(getBlockByNumberjson)
			if cache.GetAndSetProcessing(getBlockByNumbercachekey) {
				lockCost3 = time.Since(startSendChainTime).Seconds()
				//log.Printf(">>>>来源%s   定时任务发送getBlockByNumbercachekey到channel %s", source, getBlockByNumbercachekey)
				requestBlockNumberChain <- getBlockByNumberjson
			}
			sendCost3 = time.Since(startSendChainTime).Seconds()

			sendChainTimeCost := time.Since(startSendChainTime).Seconds()
			if sendChainTimeCost > 1 {
				log.Errorf("发送区块到chain耗时超过1s:%f cost:%f,%f,%f lockCost:%f,%f,%f ", sendChainTimeCost, sendCost1, sendCost2, sendCost3, lockCost1, lockCost2, lockCost3)
			}
			currentBlockNumberDec++
		}
	}
}
func FixSchedulerFetchLatestBlock(cfg *config.Config,
	requestBlockNumberChain chan bean.GenericJSON) {
	if cfg.LoopSeconds <= 0 {
		log.Infof("不启动fixSchedulerFetchLatestBlock")
		return
	}
	log.Infof("启动fixSchedulerFetchLatestBlock ,抓取最新区块，间隔%ds抓取一次", cfg.LoopSeconds)
	defer func() {
		if err := recover(); err != nil {
			log.Println("fixSchedulerFetchLatestBlock error 异常退出了:", err)
			log.Printf("*************************************recover begin")
			log.Printf("Program crashed with error: %v\n", err)
			buf := make([]byte, 1<<16)
			stackSize := runtime.Stack(buf, true)
			log.Printf("Stack trace:\n%s\n", buf[:stackSize])
			log.Printf("Stack trace:\n%s\n", debug.Stack())
			log.Printf("*************************************recover end")
		}
	}()
	ticker := time.NewTicker(time.Duration(cfg.LoopSeconds) * time.Second)
	defer ticker.Stop()
	for {
		select {
		case t := <-ticker.C:
			// 在这里执行你想要的操作
			log.Infof("fixSchedulerFetchLatestBlock执行请求最新区块 当前requestBlockNumberChainSize=%d", len(requestBlockNumberChain))
			requestBlockNumberChain <- bean.GenericJSON{
				"method":   "eth_blockNumber",
				"t":        t,
				"source":   "定时触发",
				"tid":      util.GenerateTraceID(),
				"sendTime": time.Now(),
			}
			requestBlockNumberChain <- bean.GenericJSON{
				"method":   "eth_getBlockByNumber",
				"params":   []interface{}{"latest", true},
				"t":        t,
				"source":   "定时触发",
				"tid":      util.GenerateTraceID(),
				"sendTime": time.Now(),
			}
		}
	}
}

func GoTest(cfg *config.Config, requestBlockNumberChain chan bean.GenericJSON) {
	ticker := time.NewTicker(60 * time.Second)
	defer ticker.Stop()
	for {
		select {
		case t := <-ticker.C:
			requestBlockNumberChainSize := len(requestBlockNumberChain)
			log := log.WithField("blockNumberChainSize", requestBlockNumberChainSize)
			s := 0
			if requestBlockNumberChainSize == cfg.PullJobSize {
				for {
					if s >= cfg.PullJobSize/2 { // drop
						break
					}
					s++
					data, ok := <-requestBlockNumberChain
					if !ok {
						log.Infof("Channel closed, no more Data.")
						break
					}
					log.Infof("ChannelDataDrop: %v\n", data)
				}
			}
			log.Infof("检测blockNumberChainSize=%d , %v", len(requestBlockNumberChain), t)
		}
	}
}

func PollBlockByNumberAPI(index int, config *config.Config, urlManager *url2.URLManager, httpClient *http.Client,
	cache *cache2.Cache,
	requestBlockNumberChain chan bean.GenericJSON,
	requestContexts *map[string]map[string]*bean.RequestContext,
	requestContextsM *sync.Mutex) {
	defer func() {
		if err := recover(); err != nil {
			log.Println("pollBlockByNumberAPI error 异常退出了:", err)
			log.Printf("*************************************recover begin")
			log.Printf("Program crashed with error: %v\n", err)
			buf := make([]byte, 1<<16)
			stackSize := runtime.Stack(buf, true)
			log.Printf("Stack trace:\n%s\n", buf[:stackSize])
			log.Printf("Stack trace:\n%s\n", debug.Stack())
			log.Printf("*************************************recover end")
		}
	}()

	for {
		reqMap := <-requestBlockNumberChain
		requestBlockNumberChainSize := len(requestBlockNumberChain)
		log := log.WithField("chainSize", requestBlockNumberChainSize)
		log = log.WithField("index", index)

		id := util.NextRequestId()
		payload := bean.GenericJSON{
			"jsonrpc": "2.0",
			"id":      strconv.Itoa(int(id)),
			"method":  reqMap["method"],
		}
		// 检查 reqMap 中是否包含 "params"
		if params, ok := reqMap["params"]; ok {
			if params != "" {
				payload["params"] = params
			}
		}
		times := 0
		if times2, ok := reqMap["times"]; ok {
			times, _ = times2.(int)
		}
		reqMap["times"] = times + 1

		source := reqMap["source"]
		method := reqMap["method"].(string)
		url := reqMap["url"]

		log = log.WithField("source", source)
		tid, ok := reqMap["tid"]
		if ok {
			log = log.WithField("tid", tid)
		}
		reqId, ok := reqMap["id"]
		if ok {
			log = log.WithField("reqId", reqId)
		}
		clientReqId, ok := reqMap["clientReqId"]
		if ok {
			log = log.WithField("clientReqId", clientReqId)
		}
		sendTime, ok := reqMap["sendTime"].(time.Time)
		if ok {
			chainSendCostTime := time.Since(sendTime).Seconds()
			log = log.WithField("cst", fmt.Sprintf("%.2f", chainSendCostTime))
		}
		cacheKey := cache2.CreateCacheKey(payload)
		log = log.WithField("cacheKey", util.TruncateString(cacheKey, 128))

		if times > 4 {
			log.Errorf("times超过次数,放弃发送")
			cache.Delete(cacheKey)
			continue
		}

		var httpUrl string
		if url != nil && url != "" {
			httpUrl = url.(string)
		} else if util.Contains(config.SpecialMethodMap, method) {
			httpUrl2, err := urlManager.GetRandomURL(url2.ClientRequestSpecialUrl)
			if err != nil {
				log.Errorf("不应该发生1 GetRandomURL err: %v", err)
				continue
			} else {
				httpUrl = httpUrl2
			}
		} else {
			httpUrl2, err := urlManager.GetRandomURL(url2.ClientRequestUrl)
			if err != nil {
				log.Errorf("不应该发生2 GetRandomURL urlType:%s err: %v", url2.ClientRequestUrl, err)
				continue
			} else {
				httpUrl = httpUrl2
			}
		}
		if httpUrl == "" {
			log.Errorf("不应该发生3 httpUrl 没有获取到")
			continue
		}

		//log.Printf("发送获取eth_blockNumber请求 %s", payload)

		startTime := time.Now()
		payloadBytes, err := json.Marshal(payload)
		if err != nil {
			log.Errorf("将payload装换为json异常 json marshal err: %v", err)
			continue
		}
		req, err := http.NewRequest("POST", httpUrl, io.NopCloser(bytes.NewReader(payloadBytes)))
		if err != nil {
			log.Printf("Failed to create request: %v", err)
			cache.Delete(cacheKey)
			continue
		}
		req.Header.Set("Content-Type", "application/json")

		//log.Printf("来源%s Sending to URL: %s with body: %s", source, httpUrl, string(payloadBytes))

		// 设置GetBody方法
		req.GetBody = func() (io.ReadCloser, error) {
			return ioutil.NopCloser(bytes.NewBuffer(payloadBytes)), nil
		}

		startTimeHttp := time.Now()
		resp, err := httpClient.Do(req)
		if err != nil {
			if strings.Contains(err.Error(), "Client.Timeout") {
				urlManager.AddTimeoutCount(httpUrl)
			} else {
				urlManager.AddSendErrorCount(httpUrl)
			}
			//其他错误
			httpTimeCost := time.Since(startTimeHttp).Seconds()
			log.Printf("Failed to send request  %fs times=%d url=%s,request=%s err=%v", httpTimeCost, times, httpUrl, string(payloadBytes), err)
			sendAgainIfTimesBelow(times, log, httpUrl, cacheKey, payloadBytes, "", reqMap, cache, requestBlockNumberChain)
			continue
		}
		httpTimeCost := time.Since(startTimeHttp).Seconds()

		responseBodyBytes, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			httpTimeCost := time.Since(startTimeHttp).Seconds()
			log.Printf("Failed to read response %fs times=%d url=%s,request=%s err=%v", httpTimeCost, times, httpUrl, string(payloadBytes), err)
			resp.Body.Close()
			urlManager.AddReadErrorCount(httpUrl)
			sendAgainIfTimesBelow(times, log, httpUrl, cacheKey, payloadBytes, "", reqMap, cache, requestBlockNumberChain)
			continue
		}
		errClose := resp.Body.Close()
		if errClose != nil {
			log.Errorf("不应该发生的 Failed to close response body: %v", err)
			continue
		}

		responseBodyString := string(responseBodyBytes)
		if strings.Contains(responseBodyString, "cannot be found") {
			log.Printf("区块还没同步好 %s,%s", httpUrl, responseBodyString)
			urlManager.AddBlockNotFound(httpUrl)
			sendAgainIfTimesBelow(times, log, httpUrl, cacheKey, payloadBytes, responseBodyString, reqMap, cache, requestBlockNumberChain)
			continue
		}
		if strings.Contains(responseBodyString, "Out of requests") ||
			strings.Contains(responseBodyString, "daily quota exceeded") ||
			strings.Contains(responseBodyString, "monthly limit") ||
			strings.Contains(responseBodyString, "upgrade your plan") {
			log.Printf("没有资源了,再请求一次 %s,%s", httpUrl, responseBodyString)
			urlManager.SetInvalid(httpUrl, responseBodyString)
			sendAgainIfTimesBelow(times, log, httpUrl, cacheKey, payloadBytes, responseBodyString, reqMap, cache, requestBlockNumberChain)
			continue
		}

		var responsePayload bean.GenericJSON
		if err := json.Unmarshal(responseBodyBytes, &responsePayload); err != nil {
			duration := time.Since(startTime).Seconds()
			log.Printf("解析json异常 duration=%fs url=%s resp:%v,%v", duration, httpUrl, responseBodyString, err)
			sendAgainIfTimesBelow(times, log, httpUrl, cacheKey, payloadBytes, responseBodyString, reqMap, cache, requestBlockNumberChain)
			continue
		}

		if util.IsResultEmptyOrSizeZeroOrEmptyObject(responsePayload["result"]) {
			if responsePayload["error"] != nil {
				urlManager.AddResultErrorCount(httpUrl)
			}
			urlManager.AddResultNullCount(httpUrl)
			duration := time.Since(startTime).Seconds()
			if times < 2 {
				log.Printf("渠道应答数据result为空，再请求一次 耗时=%fs times=%d,url=%s payloadBytes=%v responseBodyBytes:%v", duration, times, httpUrl, string(payloadBytes), responseBodyString)
				sendAgainIfTimesBelow(times, log, httpUrl, cacheKey, payloadBytes, responseBodyString, reqMap, cache, requestBlockNumberChain)
			} else {
				cache.Delete(cacheKey)
				log.Printf("渠道应答数据result为空，超过请求次数，放弃 耗时=%fs times=%d,url=%s payloadBytes=%v responseBodyBytes:%v", duration, times, httpUrl, string(payloadBytes), responseBodyString)
				//为空也返回给客户端，要不客户端会超时
				notifyClientDone(log, cacheKey, responsePayload, requestContexts, requestContextsM)
			}
			continue
		} else {
			var blockNumber int64
			method := payload["method"]
			if method == "eth_blockNumber" {
				blockHex, blockHexOk := responsePayload["result"].(string)
				if blockHexOk {
					blockDec, err := strconv.ParseInt(blockHex, 0, 64)
					if err != nil {
						log.Printf("不应该发生 Failed to convert block number: %v", err)
						continue
					} else {
						blockNumber = blockDec
						if config.LoopSeconds > 0 {
							go func() {
								latestBlockNumberChain <- int(blockNumber)
							}()
						}
					}
				}
			}
			duration := time.Since(startTime).Seconds()
			log.Printf("渠道应答数据成功 url=%s duration=%fs httpTimeCost=%fs times=%d \n>>>>>请求 %s \n<<<<<应答: %.200s blockNumber=%d, ",
				httpUrl, duration, httpTimeCost, times, string(payloadBytes), responseBodyBytes, blockNumber)

			cache.Add(cacheKey, responsePayload)
			notifyClientDone(log, cacheKey, responsePayload, requestContexts, requestContextsM)
		}
	}

}

func sendAgainIfTimesBelow(times int, log *log.Entry, httpUrl string, cacheKey string, payloadBytes []byte,
	responseBodyString string, reqMap bean.GenericJSON, cache *cache2.Cache, requestBlockNumberChain chan bean.GenericJSON) {
	go func() {
		if times < 3 { // 2的话重新发一次，总共2次
			log.Printf("是否重新发送判断结果ok次数还够，再请求一次url=%s times=%d, payloadBytes=%v responseBodyBytes:%v", httpUrl, times, string(payloadBytes), responseBodyString)
			reqMap["sendTime"] = time.Now()
			reqMap["url"] = "" //设置为空，后面换一个url请求
			requestBlockNumberChain <- reqMap
		} else {
			cache.Delete(cacheKey)
		}
	}()
}

func notifyClientDone(log *log.Entry, cacheKey string, responsePayload bean.GenericJSON,
	requestContexts *map[string]map[string]*bean.RequestContext,
	requestContextsM *sync.Mutex) {
	requestContextsM.Lock()
	defer requestContextsM.Unlock()

	if ctxMap, exists := (*requestContexts)[cacheKey]; exists {
		for tid, ctx := range ctxMap {
			log.Printf("发送应答给客户端channel tid: %s", tid)
			ctx.Response <- responsePayload
			close(ctx.Response) // Close the channel to signal that no more responses will be sent
		}
		delete(*requestContexts, cacheKey) // Remove the entire cacheKey entry after notifying all clients
	}
}
