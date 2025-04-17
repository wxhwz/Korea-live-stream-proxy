package main

import (
	"encoding/json"
	"net/url"
)

const UrlGetSecureToken = "https://apis.wavve.com/ip"

var SecureToken = ""

type SecureTokenResponse struct {
	IP          string                 `json:"ip"`
	Iskr        string                 `json:"iskr"`
	IsJaws      string                 `json:"isjaws"`
	SecureToken string                 `json:"securetoken"`
	IsAnonymous string                 `json:"isanonymous"`
	AnonymousIP map[string]interface{} `json:"anonymous_ip"`
}

func GetWavveSecureToken(guid string) string {
	params := url.Values{
		"apikey":         {ApiKey},
		"device":         {"android"},
		"modelid":        {ModelID},
		"partner":        {Partner},
		"region":         {Region},
		"drm":            {DRM},
		"VR":             {"y"},
		"targetage":      {"all"},
		"guid":           {guid},
		"client_version": {ClientVersion},
		"securetoken":    {"none"},
		"pooqzone":       {PooqZone},
	}
	fullUrl := UrlGetSecureToken + "?" + params.Encode()
	//LogDebug("fullUrl:", fullUrl)

	_, respBody, err := MRequest(fullUrl, "GET", nil, GetNormalHeader(), false)
	if err != nil {
		LogError(err)
		return ""
	}
	var stResponse SecureTokenResponse
	// 解析 JSON
	err = json.Unmarshal([]byte(respBody), &stResponse)
	if err != nil {
		LogError("Error unmarshaling JSON:", err)
		return ""
	}
	return stResponse.SecureToken

}
