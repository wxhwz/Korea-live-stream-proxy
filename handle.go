package main

import (
	"encoding/base64"
	"fmt"
	"net/http"
	"net/url"
	"path"
	"strings"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"golang.org/x/sync/singleflight"
)

// m3u8CacheItem 存储缓存的 m3u8 内容和过期时间
type m3u8CacheItem struct {
	content   string
	expiresAt time.Time
}
type Wavve struct {
	flight                   singleflight.Group
	tsSFKeys                 map[string]time.Time // 记录 singleflight 键和创建时间
	tsSFKeysMu               sync.Mutex
	m3u8Cache                map[string]m3u8CacheItem // 缓存 m3u8 内容
	m3u8CacheMu              sync.RWMutex
	m3u8CacheTime            time.Duration
	m3u8CacheCleanUpInterval time.Duration
}

// 初始化
func NewWavve() *Wavve {
	w := &Wavve{
		tsSFKeys:                 make(map[string]time.Time),
		m3u8Cache:                make(map[string]m3u8CacheItem),
		m3u8CacheTime:            2 * time.Second,
		m3u8CacheCleanUpInterval: 1 * time.Second,
	}
	go w.cleanupSFKeys()
	go w.cleanupM3u8Cache()
	return w
}

// 清理过期缓存
func (y *Wavve) cleanupM3u8Cache() {
	ticker := time.NewTicker(y.m3u8CacheCleanUpInterval)
	for range ticker.C {
		y.m3u8CacheMu.Lock()
		for key, item := range y.m3u8Cache {
			if time.Now().After(item.expiresAt) {
				delete(y.m3u8Cache, key)
			}
		}
		y.m3u8CacheMu.Unlock()
	}
}

// HandleMainRequest 处理主请求
func (y *Wavve) HandleMainRequest(c *gin.Context, contentID string) {
	// 检查缓存
	y.m3u8CacheMu.RLock()
	if item, ok := y.m3u8Cache[contentID]; ok && time.Now().Before(item.expiresAt) {
		y.m3u8CacheMu.RUnlock()
		c.Header("Content-Type", "application/vnd.apple.mpegurl")
		c.String(http.StatusOK, item.content)
		return
	}
	y.m3u8CacheMu.RUnlock()

	// 使用 singleflight 避免重复请求
	key := "m3u8_" + contentID
	y.addKey(key)
	result, err, _ := y.flight.Do(key, func() (interface{}, error) {
		var (
			playUrl     string
			cookie      string
			found       bool
			isPreview   bool
			previewTime int64
		)

		// 检查频道缓存
		playUrl, cookie, found = GetWavveChannelCache(contentID)
		if !found {
			playUrl, cookie, isPreview, previewTime = WavveGetLive(contentID)
			if playUrl == "" {
				return nil, fmt.Errorf("playUrl null")
			}
			finalUrl, _, err := HandleM3u8Raw(playUrl, cookie, "url")
			if err != nil {
				return nil, fmt.Errorf("handle m3u8 raw: %w", err)
			}
			// 设置频道缓存
			if isPreview {
				SetWavveChannelCache(contentID, finalUrl, cookie, previewTime)
			} else {
				SetWavveChannelCache(contentID, finalUrl, cookie, 60*60)
			}
			playUrl = finalUrl
		}

		// 获取并替换 m3u8 数据
		m3u8Content := RequestAndReplaceM3u8Data(playUrl, cookie, "http://"+c.Request.Host+c.Request.URL.Path)
		if m3u8Content == "" {
			return nil, fmt.Errorf("m3u8Content null")
		}

		// 存入缓存
		y.m3u8CacheMu.Lock()
		y.m3u8Cache[contentID] = m3u8CacheItem{
			content:   m3u8Content,
			expiresAt: time.Now().Add(y.m3u8CacheTime),
		}
		y.m3u8CacheMu.Unlock()

		return m3u8Content, nil
	})

	if err != nil {
		LogError(err)
		c.String(http.StatusNotFound, err.Error())
		return
	}

	// 返回结果
	m3u8Content := result.(string)
	c.Header("Content-Type", "application/vnd.apple.mpegurl")
	c.String(http.StatusOK, m3u8Content)
}

func HandleM3u8Raw(m3u8Url, cookie, returnType string) (string, string, error) {

	_, respBody, err := MRequest(m3u8Url, "GET", nil,
		map[string]string{
			"Cookie":          cookie,
			"Accept-Encoding": "gzip",
			"User-Agent":      "poopV2",
		}, false)
	if err != nil || respBody == "" {
		return "", "", err
	}
	parsedURL, err := url.Parse(m3u8Url)
	if err != nil {
		return "", "", err
	}
	urlPath := path.Dir(parsedURL.Path)
	latestLine := getLastLine(respBody)

	newURL := ""
	if !strings.HasPrefix(latestLine, "http") {
		newURL = fmt.Sprintf("%s://%s%s/%s", parsedURL.Scheme, parsedURL.Host, urlPath, latestLine)
	} else {
		newURL = latestLine
	}

	finalPath := newURL[:strings.LastIndex(newURL, "/")+1]

	hasM3u8 := strings.Contains(respBody, ".m3u8")

	if returnType == "url" {
		if hasM3u8 {
			return newURL, finalPath, nil
		} else {
			return m3u8Url, "", nil
		}
	}
	if !hasM3u8 {
		return respBody, finalPath, nil
	}

	return HandleM3u8Raw(newURL, cookie, returnType)
}
func RequestAndReplaceM3u8Data(playUrl, cookie, sourceUrlPath string) string {
	// lastSlash := strings.LastIndex(playUrl, "/")
	// var playUrlPath string
	// if lastSlash != -1 {
	// 	playUrlPath = playUrl[:lastSlash+1]
	// 	fmt.Println(playUrlPath)
	// } else {
	// 	LogError()
	// 	return ""
	// }
	//playUrlPath = sourceUrlPath + "?ts=" + playUrlPath

	m3u8Content, finalPath, err := HandleM3u8Raw(playUrl, cookie, "raw")
	if err != nil {
		LogError(err)
		return ""
	}
	//fmt.Println(m3u8Content)
	return ReplaceM3u8Data(m3u8Content, sourceUrlPath, finalPath, cookie)
}
func ReplaceM3u8Data(m3u8Content, sourceUrlPath, playUrlPath, cookie string) string {

	b64Cookie := base64.URLEncoding.EncodeToString([]byte(cookie))
	// 按行分割 m3u8 内容
	lines := strings.Split(m3u8Content, "\n")
	var builder strings.Builder

	// 逐行处理
	for i, line := range lines {
		// 不是第一行时添加换行符
		if i > 0 {
			builder.WriteString("\n")
		}

		// 跳过空行和以 # 开头的行（m3u8 元数据）
		if strings.TrimSpace(line) == "" || strings.HasPrefix(line, "#") {
			builder.WriteString(line)
			continue
		}

		// // 跳过已经是绝对路径的行（以 http:// 或 https:// 开头）
		// if strings.HasPrefix(line, "http://") || strings.HasPrefix(line, "https://") {
		// 	builder.WriteString(line)
		// 	continue
		// }

		// 添加 URL 前缀
		builder.WriteString(sourceUrlPath + "?ts=")
		builder.WriteString(base64.URLEncoding.EncodeToString([]byte(playUrlPath + line)))
		builder.WriteString("&cookie=")
		builder.WriteString(b64Cookie)
	}
	return builder.String()
}

func getLastLine(s string) string {
	if len(s) == 0 {
		return ""
	}

	// 从末尾向前扫描，找到最后一个非空行
	lastNonEmptyStart := -1
	for i := len(s) - 1; i >= 0; i-- {
		if s[i] == '\n' {
			// 检查当前段是否非空
			if lastNonEmptyStart != -1 && lastNonEmptyStart > i+1 {
				return s[i+1 : lastNonEmptyStart]
			}
			lastNonEmptyStart = i
		}
	}

	// 处理开头到第一个换行符或整个字符串
	if lastNonEmptyStart != -1 && lastNonEmptyStart > 0 {
		return s[0:lastNonEmptyStart]
	}
	// 如果没有换行符且字符串非空，返回整个字符串
	if lastNonEmptyStart == -1 && len(s) > 0 {
		return s
	}

	return ""
}
