package main

import (
	"flag"
	log "github.com/sirupsen/logrus"
	"github.com/yiranone/chain-api-proxy/pkg/bean"
	cache2 "github.com/yiranone/chain-api-proxy/pkg/cache"
	"github.com/yiranone/chain-api-proxy/pkg/config"
	http2 "github.com/yiranone/chain-api-proxy/pkg/http"
	log2 "github.com/yiranone/chain-api-proxy/pkg/log"
	"github.com/yiranone/chain-api-proxy/pkg/pull"
	url2 "github.com/yiranone/chain-api-proxy/pkg/url"
	"gopkg.in/natefinch/lumberjack.v2"
	"net"
	"net/http"
	"runtime"
	"runtime/debug"
	"strconv"
	"sync"
	"time"
)

var (
	cache                   = cache2.NewCache()
	requestBlockNumberChain = make(chan bean.GenericJSON, 100) // Channel for incoming requests
	requestContexts         = make(map[string]map[string]*bean.RequestContext)
	requestContextsM        sync.Mutex
)

func main() {

	defer func() {
		if err := recover(); err != nil {
			log.Println("main error 异常退出了:", err)
			log.Printf("*************************************recover begin")
			log.Printf("Program crashed with error: %v\n", err)
			buf := make([]byte, 1<<16)
			stackSize := runtime.Stack(buf, true)
			log.Printf("Stack trace:\n%s\n", buf[:stackSize])
			log.Printf("Stack trace:\n%s\n", debug.Stack())
			log.Printf("*************************************recover end")
		}
	}()

	// 定义命令行参数
	configPath := flag.String("config", "config.yaml", "Path to the configuration file")
	flag.Parse()

	// 读取配置文件
	cfg, err := config.LoadConfig(*configPath)
	if err != nil {
		log.Fatalf("Failed to load config file: %v", err)
	}
	logPath := cfg.LogPath
	if logPath != "" && logPath != "-" {
		log.Printf("设置日志文件保存位置%v", logPath)
		mutWr := &lumberjack.Logger{
			Filename:   logPath + "/log.txt",
			MaxSize:    400, // megabytes
			MaxBackups: 2,
			MaxAge:     3,     //days
			Compress:   false, // disabled by default
		}
		log.SetOutput(mutWr)
	}
	log.SetFormatter(&log2.MyLogFormatter{})
	log := log.WithField("source", "系统启动")
	// 使用解析后的配置
	for _, url := range cfg.ClientRequestUrl {
		log.Println("ClientRequestUrl:", url)
	}
	for _, url := range cfg.ClientRequestSpecialUrl {
		log.Println("ClientRequestSpecialUrl:", url)
	}
	for _, url := range cfg.JobRequestUrl {
		log.Println("JobRequestUrl:", url)
	}
	for _, url := range cfg.JobRequestSpecialUrl {
		log.Println("JobRequestSpecialUrl:", url)
	}

	urlManager := url2.NewURLManager(cfg.ClientRequestUrl, cfg.ClientRequestSpecialUrl, cfg.JobRequestUrl, cfg.JobRequestSpecialUrl)
	urlManager.SetAllValid()
	go urlManager.StartResetScheduler()

	go cache.CleanupExpiredItems(cfg.CacheSeconds)

	log.Infof("后端httpclient设置全局超时时间%ds", cfg.BackendHttpSeconds)
	transport := &http.Transport{
		IdleConnTimeout:     5 * time.Minute,
		MaxIdleConns:        20,
		MaxIdleConnsPerHost: 10,
		MaxConnsPerHost:     80,
		DialContext: (&net.Dialer{
			Timeout:   7 * time.Second,
			KeepAlive: 1 * time.Minute,
		}).DialContext,
	}
	httpClient := &http.Client{
		Timeout:   time.Duration(cfg.BackendHttpSeconds) * time.Second, // 设置全局超时时间
		Transport: transport,
	}

	log.Infof("启动拉取JOB，数量%d", cfg.PullJobSize)
	for i := 0; i < cfg.PullJobSize; i++ {
		go pull.PollBlockByNumberAPI(i, cfg, urlManager, httpClient, cache, requestBlockNumberChain, &requestContexts, &requestContextsM)
	}
	go pull.FixSchedulerFetchLatestBlock(cfg, requestBlockNumberChain)
	go pull.CalcLatestBlockNumber(cfg, urlManager, cache, requestBlockNumberChain)
	go pull.GoTest(cfg, requestBlockNumberChain)

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		http2.Handler(w, r, cfg, cache, requestBlockNumberChain, &requestContexts, &requestContextsM)
	})
	http.HandleFunc("/urlStats", func(w http.ResponseWriter, r *http.Request) {
		http2.HandlerUrlManagerStatusRequest(w, r, urlManager)
	})
	http.HandleFunc("/cache", func(w http.ResponseWriter, r *http.Request) {
		http2.HandlerCacheStatusRequest(w, r, cache)
	})

	log.Printf("Listening on port %d...", cfg.Port)
	log.Fatal(http.ListenAndServe(":"+strconv.Itoa(cfg.Port), nil))
}
