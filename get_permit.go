package main

import (
	"encoding/json"
	"net/url"
)

var Permit = "none"

const UrlGetPermit = "https://apis.wavve.com/updatepermit"

func WavveGetPermit() {
	if Credential == "none" {
		SecureToken = GetWavveSecureToken(GUID)
		if SecureToken == "" {
			LogError("SecureToken is none, get permit fail")
			return
		}
		WavveSign()
		if Credential == "none" {
			LogError("Credential is none, get permit fail")
			return
		}
	}

	params := url.Values{
		"permit":         {Permit},
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
	fullUrl := UrlGetPermit + "?" + params.Encode()
	_, respBody, err := MRequest(fullUrl, "GET", nil, GetNormalHeader(), false)
	if err != nil {
		LogError(err)
		return
	}

	var permitResponse PermitResponse
	// 解析 JSON
	err = json.Unmarshal([]byte(respBody), &permitResponse)
	if err != nil {
		LogError("Error unmarshaling JSON:", err)
		return
	}
	if permitResponse.ResultCode == "200" {
		Permit = permitResponse.Permit
	} else {
		LogError(permitResponse.ResultMessage)
		return
	}
	LogDebug("Permit:", Permit)

}

type PermitResponse struct {
	ResultCode    string `json:"resultcode"`
	ResultMessage string `json:"resultmessage"`
	Permit        string `json:"permit"`
	Debug         string `json:"debug"`
	Extra         string `json:"extra"`
}
