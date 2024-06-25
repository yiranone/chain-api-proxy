# web3 api forward/proxy (web3服务api的转发/代理)

### Chain-API-Proxy: A Handy Web3 Proxy

`chain-api-proxy` is an efficient Web3 proxy that can forward JSON API requests to backend API providers such as ankr.com and blastapi.io.

Your app => chain-api-proxy => Chain API providers => Blockchain

#### Why use chain-api-proxy?

1. **Reduce request numbers**: Aggregate identical requests from multiple clients and only request the backend service once.
2. **Cache requests**: Identical requests can be read directly from the cache and returned.
3. **Pre-fetch block data**: Use scheduled tasks to fetch data from backend channels in advance and store it in the cache.
4. **Support multiple backend channels**: Configure multiple backend URLs and automatically isolate them in case of failures.
5. **Written in Go**: Easy to deploy and maintain.

### Chain-API-Proxy：高效的Web3代理

`chain-api-proxy` 是一个高效的 Web3 代理，可以将 JSON API 请求转发到后端的 API 服务商，例如 ankr.com 和 blastapi.io。

您的应用程序 => chain-api-proxy => 链 API 服务商 => 区块链

#### 为什么使用 chain-api-proxy?

1. **减少请求数量**：汇聚多个客户端的相同请求，只向后端服务请求一次。
2. **缓存请求**：相同的请求可以从缓存中读取并直接返回。
3. **提前获取区块数据**：使用定时任务提前从后端渠道获取数据并放入缓存。
4. **支持多个后端渠道**：可以配置多个后端 URL，在请求失败时能自动隔离故障渠道。
5. **使用 Go 编写**：易于部署和维护。

# how to use (如何使用)

chain-api-proxy -config config/sample.yaml
