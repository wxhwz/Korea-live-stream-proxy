package main

import (
	"encoding/json"
	"fmt"
	"net/url"
	"strings"
	"time"
)

var wavveM3UExpTime int64

const cacheTime = 3 * time.Hour

func WavveGenerateM3U(host string) string {
	if time.Now().Unix() >= wavveM3UExpTime {
		WavveUpdateChannels()
	}

	var builder strings.Builder
	builder.WriteString("#EXTM3U\n")
	for _, ch := range WavveChannelslistResponse.CellTopList.CellList {
		builder.WriteString("#EXTINF:-1,")
		builder.WriteString("tvg-id=\"" + ch.TitleList[0].Text + "\" ")
		builder.WriteString("tvg-name=\"" + ch.TitleList[0].Text + "\" ")
		builder.WriteString("tvg-logo=\"" + ch.RankThumbnail + "\" ")
		builder.WriteString("group-title=\"" + "韩国" + "\",")
		builder.WriteString(ch.TitleList[0].Text + "\n")

		builder.WriteString("http://")
		builder.WriteString(host)
		builder.WriteString("/wavve/")
		builder.WriteString(ch.ContentID)
		builder.WriteString(".m3u8")
		builder.WriteString("\n")
	}
	return builder.String()
}

func WavveUpdateChannels() {
	params := url.Values{
		"isrecommend":    {"n"},
		"contenttype":    {"channel"},
		"genre":          {"all"},
		"weekday":        {"all"},
		"offset":         {"0"},
		"limit":          {"999"},
		"apikey":         {ApiKey},
		"device":         {Device},
		"modelid":        {ModelID},
		"partner":        {Partner},
		"region":         {Region},
		"drm":            {DRM},
		"VR":             {"y"},
		"targetage":      {"all"},
		"guid":           {GUID},
		"client_version": {ClientVersion},
		"securetoken":    {SecureToken},
		"pooqzone":       {PooqZone},
	}
	fullUrl := UrlGetChannels + "?" + params.Encode()
	//LogDebug("fullUrl:", fullUrl)

	_, respBody, err := MRequest(fullUrl, "GET", nil, GetNormalHeader(), false)
	if err != nil {
		LogError(err)
		wavveM3UExpTime = 0
		return
	}
	//LogDebug(respBody)

	// 解析 JSON
	err = json.Unmarshal([]byte(respBody), &WavveChannelslistResponse)
	if err != nil {
		LogError("Error unmarshaling JSON:", err)
		wavveM3UExpTime = 0
		return
	}
	// 打印解析结果
	// fmt.Printf("Parsed Response:\n")
	// fmt.Printf("Type: %s\n", WavveChannelslistResponse.Type)
	// fmt.Printf("SubType: %s\n", WavveChannelslistResponse.SubType)
	// fmt.Printf("GenTime: %s\n", WavveChannelslistResponse.GenTime)
	// fmt.Printf("Version: %s\n", WavveChannelslistResponse.Version)
	// fmt.Printf("CellTopList Title: %s\n", WavveChannelslistResponse.CellTopList.TitleList[0].Text)
	// fmt.Printf("CellList Title: %s\n", WavveChannelslistResponse.CellTopList.CellList[0].TitleList[0].Text)
	// fmt.Printf("Filter BaseAPI: %s\n", WavveChannelslistResponse.Filter.BaseAPI)
	fmt.Printf("Wavve Channel Count %s\n", WavveChannelslistResponse.CellTopList.Count)
	for i, ch := range WavveChannelslistResponse.CellTopList.CellList {
		//fmt.Printf("Index:%d ContentID:%s ChannelName:%s\n", i+1, ch.ContentID, ch.TitleList[0].Text)
		if !strings.HasPrefix(ch.RankThumbnail, "http") {
			WavveChannelslistResponse.CellTopList.CellList[i].RankThumbnail = "https://" + WavveChannelslistResponse.CellTopList.CellList[i].RankThumbnail
		}
	}
	LogInfo("Update Wavve Channels Finish!")
	wavveM3UExpTime = time.Now().Unix() + int64(cacheTime)
}
