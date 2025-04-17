package main

import (
	"encoding/json"
	"fmt"
	"net/url"
)

var Credential = "none"
var UserName = ""
var Password = ""

const UrlSignin = "https://account-api.wavve.com/v0.9/signin/wavve"

func WavveSign() {
	params := url.Values{
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
		"credential":     {Credential},
	}
	fullUrl := UrlSignin + "?" + params.Encode()
	header := map[string]string{
		"accept":           "application/json",
		"wavve-credential": Credential,
		"content-type":     "application/json",
		"accept-encoding":  "gzip",
		"user-agent":       "okhttp/5.0.0-alpha.10",
	}
	_, respBody, err := MRequest(fullUrl, "POST", map[string]string{
		"type":                 "general",
		"id":                   UserName,
		"profile":              "0",
		"pushid":               "none",
		"password":             Password,
		"adid":                 "",
		"markettype":           "unknown",
		"installerpackagename": "com.aefyr.sai",
		"carrier":              "none",
		"mcc":                  "none",
		"mnc":                  "none",
		"simoperator":          "46000",
		"networktype":          "1",
	}, header, false)
	if err != nil {
		LogError(err)
		return
	}
	fmt.Println(respBody)
	var signinResponse SigninResponse
	// 解析 JSON
	err = json.Unmarshal([]byte(respBody), &signinResponse)
	if err != nil {
		LogError("Error unmarshaling JSON:", err)
		return
	}
	Credential = signinResponse.Credential
	LogDebug("Credential:", Credential)
}

type SigninResponse struct {
	Credential         string `json:"credential"`
	Uno                string `json:"uno"`
	Name               string `json:"name"`
	Profile            string `json:"profile"`
	Type               string `json:"type"`
	JoinDate           string `json:"joindate"`
	ProfileName        string `json:"profilename"`
	ProfileCount       string `json:"profilecount"`
	ProfileImage       string `json:"profileimage"`
	NeedSelectProfile  string `json:"needselectprofile"`
	NeedChangePassword string `json:"needchangepassword"`
	AppPush            string `json:"apppush"`
	AppPushAgreeDate   string `json:"apppush_agreedate"`
	MovieUICode        string `json:"movieuicode"`
	DeviceID           string `json:"device_id"`
}
