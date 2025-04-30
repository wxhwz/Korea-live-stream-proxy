package main

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/patrickmn/go-cache"
)

var EnableTSMemCache = false

//var EnableSingleFlight = false

// 全局内存缓存实例
var memCache = cache.New(10*time.Second, 5*time.Second)

// // 全局 singleflight.Group 实例
// var requestGroup singleflight.Group

// 记录键
func (y *Wavve) addKey(key string) {
	y.tsSFKeysMu.Lock()
	y.tsSFKeys[key] = time.Now()
	y.tsSFKeysMu.Unlock()
}

// 清理过期键（例如超过 5 分钟）
func (y *Wavve) cleanupSFKeys() {
	ticker := time.NewTicker(1 * time.Minute)
	for range ticker.C {
		y.tsSFKeysMu.Lock()
		for key, created := range y.tsSFKeys {
			if time.Since(created) > 5*time.Minute {
				y.flight.Forget(key)
				delete(y.tsSFKeys, key)
			}
		}
		y.tsSFKeysMu.Unlock()
	}
}

// 生成缓存键（基于 URL 和 Range 头部）
func generateTSKey(tsUrl, rangeHeader string) string {
	// h := sha256.New()
	// h.Write([]byte(tsUrl + rangeHeader))
	// return hex.EncodeToString(h.Sum(nil))
	return tsUrl + "|" + rangeHeader
}

// func (y *Wavve) HandleTsRequestCache(c *gin.Context, tsUrl, cookie string) {
// 	tsUrlBytes, err := base64.URLEncoding.DecodeString(tsUrl)
// 	if err != nil {
// 		LogError("Invalid tsUrl", err)
// 		c.String(http.StatusBadRequest, "Invalid tsUrl")
// 		return
// 	}
// 	tsUrl = string(tsUrlBytes)
// 	cookieBytes, err := base64.URLEncoding.DecodeString(cookie)
// 	if err != nil {
// 		LogError("Invalid cookie", err)
// 		c.String(http.StatusBadRequest, "Invalid cookie")
// 		return
// 	}
// 	cookie = string(cookieBytes)

// 	// 解析 URL
// 	parsedURL, err := url.Parse(tsUrl)
// 	if err != nil {
// 		LogError("Invalid URL: ", err)
// 		c.String(http.StatusBadRequest, "Invalid URL")
// 		return
// 	}

// 	// 检查白名单
// 	if _, exist := ProxyUrlWhiteList[parsedURL.Host]; !exist {
// 		LogError("Unknown TS host: ", tsUrl)
// 		c.String(http.StatusNotFound, "Unknown TS host")
// 		return
// 	}

// 	// 构造请求头，验证并复制客户端的 Range 头部
// 	requestHeader := map[string]string{
// 		"Cookie":          cookie,
// 		"User-Agent":      "poopV2",
// 		"Accept-Encoding": "identity", // 禁用 gzip 压缩
// 	}
// 	rangeHeader := c.GetHeader("Range")
// 	if rangeHeader != "" {
// 		if !strings.HasPrefix(rangeHeader, "bytes=") || strings.Contains(rangeHeader, "..") {
// 			LogError("Invalid Range header: ", rangeHeader)
// 			c.String(http.StatusBadRequest, "Invalid Range header")
// 			return
// 		}
// 		requestHeader["Range"] = rangeHeader
// 	}

// 	// 生成缓存键
// 	cacheKey := generateTSCacheKey(tsUrl, rangeHeader)

// 	// 检查内存缓存
// 	if cachedData, found := memCache.Get(cacheKey); found {
// 		data := cachedData.([]byte)
// 		c.Header("Content-Type", "video/mp2t")
// 		c.Header("X-Cache", "HIT-MEMORY")
// 		c.Status(http.StatusOK)
// 		c.Writer.Write(data)
// 		return
// 	}

// 	// 发送请求
// 	resp, err := MRequestTS(tsUrl, "GET", requestHeader)
// 	if err != nil {
// 		LogError("Failed to fetch TS: ", err)
// 		c.String(http.StatusBadGateway, "Failed to fetch TS")
// 		return
// 	}
// 	defer resp.Body.Close()
// 	if resp.StatusCode != http.StatusOK {
// 		LogError("Unexpected status code: ", resp.Status)
// 		c.String(http.StatusBadGateway, "Failed to fetch TS")
// 		return
// 	}

// 	// 设置响应头
// 	for key, values := range resp.Header {
// 		for _, value := range values {
// 			c.Header(key, value)
// 		}
// 	}
// 	c.Header("Content-Type", "video/mp2t")
// 	c.Header("X-Cache", "MISS")

// 	// 设置状态码
// 	c.Status(resp.StatusCode)

// 	// 流式传输数据，处理连接中断
// 	reader := resp.Body
// 	ctx := c.Request.Context()
// 	writer := c.Writer

// 	// 确保 writer 支持 Flush
// 	flusher, canFlush := writer.(http.Flusher)

// 	// 使用缓冲区逐步传输数据
// 	var cacheBuffer []byte
// 	buf := make([]byte, 64*1024) // 64KB 缓冲区
// 	for {
// 		select {
// 		case <-ctx.Done():
// 			LogError("Client connection closed: ", ctx.Err())
// 			return
// 		default:
// 			nr, er := reader.Read(buf)
// 			if nr > 0 {
// 				nw, ew := writer.Write(buf[:nr])
// 				if ew != nil {
// 					LogError("Write error: ", ew)
// 					return
// 				}
// 				if nr != nw {
// 					LogError("Short write")
// 					return
// 				}
// 				// 收集数据用于缓存
// 				cacheBuffer = append(cacheBuffer, buf[:nr]...)
// 				// 定期 Flush 减少缓冲延迟
// 				if canFlush {
// 					flusher.Flush()
// 				}
// 			}
// 			if er != nil {
// 				if er != io.EOF {
// 					LogError("Read error: ", er)
// 				}
// 				// 缓存数据到内存
// 				if len(cacheBuffer) > 0 {
// 					memCache.Set(cacheKey, cacheBuffer, cache.DefaultExpiration)
// 				}
// 				return
// 			}
// 		}
// 	}

// }
func (y *Wavve) HandleTsRequest(c *gin.Context, tsUrl, cookie string) {
	tsUrlBytes, err := base64.URLEncoding.DecodeString(tsUrl)
	if err != nil {
		LogError("Invalid tsUrl", err)
		c.String(http.StatusBadRequest, "Invalid tsUrl")
		return
	}
	tsUrl = string(tsUrlBytes)
	cookieBytes, err := base64.URLEncoding.DecodeString(cookie)
	if err != nil {
		LogError("Invalid cookie", err)
		c.String(http.StatusBadRequest, "Invalid cookie")
		return
	}
	cookie = string(cookieBytes)

	// 解析 URL
	parsedURL, err := url.Parse(tsUrl)
	if err != nil {
		LogError("Invalid URL: ", err)
		c.String(http.StatusBadRequest, "Invalid URL")
		return
	}

	// 检查白名单
	if _, exist := ProxyUrlWhiteList[parsedURL.Host]; !exist {
		LogError("Unknown TS host: ", tsUrl)
		c.String(http.StatusNotFound, "Unknown TS host")
		return
	}

	// 构造请求头，验证并复制客户端的 Range 头部
	requestHeader := map[string]string{
		"Cookie":          cookie,
		"User-Agent":      "poopV2",
		"Accept-Encoding": "identity", // 禁用 gzip 压缩
	}
	rangeHeader := c.GetHeader("Range")
	if rangeHeader != "" {
		if !strings.HasPrefix(rangeHeader, "bytes=") || strings.Contains(rangeHeader, "..") {
			LogError("Invalid Range header: ", rangeHeader)
			c.String(http.StatusBadRequest, "Invalid Range header")
			return
		}
		requestHeader["Range"] = rangeHeader
	}

	// 发送请求
	resp, err := MRequestTS(tsUrl, "GET", requestHeader)
	if err != nil {
		LogError("Failed to fetch TS: ", err)
		c.String(http.StatusBadGateway, "Failed to fetch TS")
		return
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		LogError("Unexpected status code: ", resp.Status)
		c.String(http.StatusBadGateway, "Failed to fetch TS")
		return
	}

	// 设置响应头
	for key, values := range resp.Header {
		for _, value := range values {
			c.Header(key, value)
		}
	}
	c.Header("Content-Type", "video/mp2t")
	c.Header("X-Cache", "MISS")

	// 设置状态码
	c.Status(resp.StatusCode)

	// 流式传输数据，处理连接中断
	reader := resp.Body
	ctx := c.Request.Context()
	writer := c.Writer

	// 确保 writer 支持 Flush
	flusher, canFlush := writer.(http.Flusher)

	// 使用缓冲区逐步传输数据
	buf := make([]byte, 64*1024) // 64KB 缓冲区
	for {
		select {
		case <-ctx.Done():
			LogError("Client connection closed: ", ctx.Err())
			return
		default:
			nr, er := reader.Read(buf)
			if nr > 0 {
				nw, ew := writer.Write(buf[:nr])
				if ew != nil {
					LogError("Write error: ", ew)
					return
				}
				if nr != nw {
					LogError("Short write")
					return
				}
				// 定期 Flush 减少缓冲延迟
				if canFlush {
					flusher.Flush()
				}
			}
			if er != nil {
				if er != io.EOF {
					LogError("Read error: ", er)
				}
				return
			}
		}
	}

}

func (y *Wavve) HandleTsRequestCacheSF(c *gin.Context, tsUrl, cookie string) {
	tsUrlBytes, err := base64.URLEncoding.DecodeString(tsUrl)
	if err != nil {
		LogError("Invalid tsUrl", err)
		c.String(http.StatusBadRequest, "Invalid tsUrl")
		return
	}
	tsUrl = string(tsUrlBytes)
	cookieBytes, err := base64.URLEncoding.DecodeString(cookie)
	if err != nil {
		LogError("Invalid cookie", err)
		c.String(http.StatusBadRequest, "Invalid cookie")
		return
	}
	cookie = string(cookieBytes)

	// 解析 URL
	parsedURL, err := url.Parse(tsUrl)
	if err != nil {
		LogError("Invalid URL: ", err)
		c.String(http.StatusBadRequest, "Invalid URL")
		return
	}

	// 检查白名单
	if _, exist := ProxyUrlWhiteList[parsedURL.Host]; !exist {
		LogError("Unknown TS host: ", tsUrl)
		c.String(http.StatusNotFound, "Unknown TS host")
		return
	}

	// 构造请求头，验证并复制客户端的 Range 头部
	requestHeader := map[string]string{
		"Cookie":          cookie,
		"User-Agent":      "poopV2",
		"Accept-Encoding": "identity", // 禁用 gzip 压缩
	}
	// rangeHeader := c.GetHeader("Range")
	// if rangeHeader != "" {
	// 	if !strings.HasPrefix(rangeHeader, "bytes=") || strings.Contains(rangeHeader, "..") {
	// 		LogError("Invalid Range header: ", rangeHeader)
	// 		c.String(http.StatusBadRequest, "Invalid Range header")
	// 		return
	// 	}
	// 	requestHeader["Range"] = rangeHeader
	// }

	// // 生成缓存键
	// cacheKey := generateTSKey(tsUrl, rangeHeader)
	cacheKey := tsUrl

	// 检查内存缓存
	if cachedData, found := memCache.Get(cacheKey); found {
		data := cachedData.([]byte)
		c.Header("Content-Type", "video/mp2t")
		c.Header("X-Cache", "HIT-MEMORY")
		c.Status(http.StatusOK)
		c.Writer.Write(data)
		return
	}
	y.addKey(cacheKey) // 记录键
	// 使用 singleflight 避免重复请求
	v, err, _ := y.flight.Do(cacheKey, func() (interface{}, error) {
		// 发送请求
		resp, err := MRequestTS(tsUrl, "GET", requestHeader)
		if err != nil {
			return nil, err
		}
		defer resp.Body.Close()
		if resp.StatusCode != http.StatusOK {
			return nil, fmt.Errorf("unexpected status code: %s", resp.Status)
		}

		// 读取响应体
		var cacheBuffer []byte
		buf := make([]byte, 64*1024) // 64KB 缓冲区
		for {
			nr, er := resp.Body.Read(buf)
			if nr > 0 {
				cacheBuffer = append(cacheBuffer, buf[:nr]...)
			}
			if er != nil {
				if er != io.EOF {
					return nil, er
				}
				break
			}
		}

		// 缓存数据到内存
		if len(cacheBuffer) > 0 {
			memCache.Set(cacheKey, cacheBuffer, cache.DefaultExpiration)
		}

		return cacheBuffer, nil
	})

	if err != nil {
		LogError("Failed to fetch TS: ", err)
		c.String(http.StatusBadGateway, "Failed to fetch TS")
		return
	}

	// 设置响应头
	c.Header("Content-Type", "video/mp2t")
	c.Header("X-Cache", "MISS")
	c.Status(http.StatusOK)

	// 写入响应数据
	data := v.([]byte)
	c.Writer.Write(data)
}

// HandleTsRequest 修改后的函数
func (y *Wavve) HandleTsRequestSF(c *gin.Context, tsUrl, cookie string) {
	// 解码 tsUrl 和 cookie（保持不变）
	tsUrlBytes, err := base64.URLEncoding.DecodeString(tsUrl)
	if err != nil {
		LogError("Invalid tsUrl", err)
		c.String(http.StatusBadRequest, "Invalid tsUrl")
		return
	}
	tsUrl = string(tsUrlBytes)
	cookieBytes, err := base64.URLEncoding.DecodeString(cookie)
	if err != nil {
		LogError("Invalid cookie", err)
		c.String(http.StatusBadRequest, "Invalid cookie")
		return
	}
	cookie = string(cookieBytes)

	// 解析 URL 和白名单检查（保持不变）
	parsedURL, err := url.Parse(tsUrl)
	if err != nil {
		LogError("Invalid URL: ", err)
		c.String(http.StatusBadRequest, "Invalid URL")
		return
	}
	if _, exist := ProxyUrlWhiteList[parsedURL.Host]; !exist {
		LogError("Unknown TS host: ", tsUrl)
		c.String(http.StatusNotFound, "Unknown TS host")
		return
	}

	// 构造请求头（保持不变）
	requestHeader := map[string]string{
		"Cookie":          cookie,
		"User-Agent":      "poopV2",
		"Accept-Encoding": "identity",
	}
	// rangeHeader := c.GetHeader("Range")
	// if rangeHeader != "" {
	// 	if !strings.HasPrefix(rangeHeader, "bytes=") || strings.Contains(rangeHeader, "..") {
	// 		LogError("Invalid Range header: ", rangeHeader)
	// 		c.String(http.StatusBadRequest, "Invalid Range header")
	// 		return
	// 	}
	// 	requestHeader["Range"] = rangeHeader
	// }

	// // 使用 singleflight 包装网络请求
	// // 以 tsUrl 和 rangeHeader 作为 key，确保相同请求被合并
	// key := tsUrl + "|" + rangeHeader
	key := tsUrl
	y.addKey(key) // 记录键
	result, err, shared := y.flight.Do(key, func() (interface{}, error) {
		// 发送请求
		resp, err := MRequestTS(tsUrl, "GET", requestHeader)
		if err != nil {
			return nil, err
		}
		if resp.StatusCode != http.StatusOK {
			resp.Body.Close()
			return nil, fmt.Errorf("unexpected status code: %s", resp.Status)
		}

		// 读取响应数据到内存（注意：需要考虑大文件的情况）
		data, err := io.ReadAll(resp.Body)
		resp.Body.Close()
		if err != nil {
			return nil, err
		}

		return &tsResponse{
			Data:   data,
			Header: resp.Header,
		}, nil
	})
	if err != nil {
		LogError("Failed to fetch TS: ", err)
		c.String(http.StatusBadGateway, "Failed to fetch TS")
		return
	}

	// 提取 singleflight 返回的结果
	respData := result.(*tsResponse)

	// 设置响应头
	for key, values := range respData.Header {
		for _, value := range values {
			c.Header(key, value)
		}
	}
	c.Header("Content-Type", "video/mp2t")
	c.Header("X-Cache", "MISS")
	if shared {
		c.Header("X-Cache", "HIT") // 标记为 singleflight 缓存命中
	}

	// 设置状态码
	c.Status(http.StatusOK)

	// 流式传输数据
	ctx := c.Request.Context()
	writer := c.Writer
	flusher, canFlush := writer.(http.Flusher)

	// 使用 bytes.NewReader 将内存数据流式传输
	reader := bytes.NewReader(respData.Data)
	buf := make([]byte, 64*1024)
	for {
		select {
		case <-ctx.Done():
			LogError("Client connection closed: ", ctx.Err())
			return
		default:
			nr, er := reader.Read(buf)
			if nr > 0 {
				nw, ew := writer.Write(buf[:nr])
				if ew != nil {
					LogError("Write error: ", ew)
					return
				}
				if nr != nw {
					LogError("Short write")
					return
				}
				if canFlush {
					flusher.Flush()
				}
			}
			if er != nil {
				if er != io.EOF {
					LogError("Read error: ", er)
				}
				return
			}
		}
	}
}

// tsResponse 用于存储 singleflight 返回的响应数据
type tsResponse struct {
	Data   []byte
	Header http.Header
}
