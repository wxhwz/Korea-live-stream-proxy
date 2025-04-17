package main

import (
	"sync"
	"time"
)

var WavveChannelsCache sync.Map

type wavveChannelCacheItem struct {
	PlayUrl    string
	Cookie     string
	Expiration int64
}

// 从缓存中获取数据
func GetWavveChannelCache(key string) (string, string, bool) {
	// 查找缓存
	if item, found := WavveChannelsCache.Load(key); found {
		cacheItem := item.(wavveChannelCacheItem)
		// 检查缓存是否过期
		if time.Now().Unix() < cacheItem.Expiration {
			return cacheItem.PlayUrl, cacheItem.Cookie, true
		}
	}
	// 如果没有找到或缓存已过期，返回空
	return "", "", false
}

func SetWavveChannelCache(key, playUrl, cookie string, aviableTimeSec int64) {
	WavveChannelsCache.Store(key, wavveChannelCacheItem{
		PlayUrl:    playUrl,
		Cookie:     cookie,
		Expiration: time.Now().Unix() + aviableTimeSec,
	})
}
