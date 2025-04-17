package main

import (
	"encoding/json"
	"net/url"
	"strconv"
)

const UrlGetLive = "https://delivery.wavve.com/v1/streaming/live"

var ProxyUrlWhiteList = map[string]bool{}

func WavveGetLive(contentID string) (string, string, bool, int64) {
	params := url.Values{
		"contenttype":    {"live"},
		"contentid":      {contentID},
		"protocol":       {"hls"},
		"quality":        {"auto"},
		"authtype":       {"cookie"},
		"lastplayid":     {"none"},
		"isabr":          {"n"},
		"ishevc":         {"y"},
		"ismno":          {"n"},
		"carrier":        {"none"},
		"mcc":            {"none"},
		"mnc":            {"none"},
		"timestamp":      {"20250416215508"},
		"permit":         {Permit}, //需要登录
		"withinsubtitle": {"n"},
		"hdr":            {"HDR"},
		"issurround":     {"y"},
		"videocodec":     {"HEVC"},
		"audiocodec":     {"aac"},
		"audioChannel":   {"51ch"},
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
	// 拼接 URL 和编码后的查询参数
	fullUrl := UrlGetLive + "?" + params.Encode()
	//LogDebug("fullUrl:", fullUrl)

	statusCode, respBody, err := MRequestCustom(fullUrl, "GET", nil, GetNormalHeader(), false, 1, 0)
	if statusCode == 550 {
		LogError("Channel", contentID, "Need Subsciption, or your ip is restricted")
		return "", "", true, 0
	}
	if err != nil {
		LogError(err)
		return "", "", true, 0
	}
	//fmt.Println(respBody)
	var liveResponse LiveResponse
	// 解析 JSON
	err = json.Unmarshal([]byte(respBody), &liveResponse)
	if err != nil {
		LogError("Error unmarshaling JSON:", err)
		return "", "", true, 0
	}
	parsedURL, err := url.Parse(liveResponse.PlayURL)
	if err != nil {
		LogError(err)
		return "", "", true, 0
	}
	ProxyUrlWhiteList[parsedURL.Host] = true
	//var isPreview = false
	if liveResponse.Play == "p" {
		//isPreview = true
		previewtime, _ := strconv.Atoi(liveResponse.PreviewTime)
		return liveResponse.PlayURL, liveResponse.AWSCookie, true, int64(previewtime)
	} else {
		return liveResponse.PlayURL, liveResponse.AWSCookie, false, 0
	}

}

type LiveResponse struct {
	Play               string                 `json:"play"`
	PlayID             string                 `json:"playid"`
	Issue              string                 `json:"issue"`
	PlayTime           string                 `json:"playtime"`
	PreviewTime        string                 `json:"previewtime"`
	OnAirVOD           map[string]interface{} `json:"onairvod"`
	PlayURL            string                 `json:"playurl"`
	LiveURL            string                 `json:"liveurl"`
	EtcURL             string                 `json:"etcurl"`
	Subtitles          []interface{}          `json:"subtitles"`
	MediaType          string                 `json:"mediatype"`
	Quality            string                 `json:"quality"`
	Qualities          qualities              `json:"qualities"`
	DRMType            string                 `json:"drmtype"`
	DRM                map[string]interface{} `json:"drm"`
	Country            string                 `json:"country"`
	AuthType           string                 `json:"authtype"`
	AWSCookie          string                 `json:"awscookie"`
	ChargedType        string                 `json:"chargedtype"`
	PriceType          string                 `json:"pricetype"`
	ShiftDuration      string                 `json:"shiftduration"`
	NextTriggerSeconds string                 `json:"nexttriggerseconds"`
	Preview            preview                `json:"preview"`
	BookmarkExtra      bookmarkExtra          `json:"bookmarkextra"`
	PrerollAd          prerollAd              `json:"prerollad"`
	ErrorMessage       string                 `json:"errormessage"`
	ExtraItem          string                 `json:"extraitem"`
	ConcurrencyGroup   string                 `json:"concurrencygroup"`
	PostScreen         string                 `json:"postscreen"`
	Marketing          map[string]interface{} `json:"marketing"`
	Version            string                 `json:"version"`
	From               string                 `json:"from"`
	Debug              debug                  `json:"debug"`
	Duration           int                    `json:"duration"`
}

type qualities struct {
	IsHEVC     string        `json:"ishevc"`
	PageCount  string        `json:"pagecount"`
	Count      string        `json:"count"`
	MediaTypes []string      `json:"mediatypes"`
	List       []qualityItem `json:"list"`
}

type qualityItem struct {
	ID       string `json:"id"`
	Name     string `json:"name"`
	Marks    string `json:"marks"`
	FileSize string `json:"filesize"`
}

type preview struct {
	PreviewBtn string `json:"previewbtn"`
	PreviewMsg string `json:"previewmsg"`
	ExitBtn    string `json:"exitbtn"`
	ExitMsg    string `json:"exitmsg"`
	ExitTitle  string `json:"exittitle"`
}

type bookmarkExtra struct {
	Product string `json:"product"`
	PCA     string `json:"pca"`
	Origin  string `json:"origin"`
}

type prerollAd struct {
	ContentNumber string `json:"contentnumber"`
	BroadDate     string `json:"broaddate"`
	StartTime     string `json:"starttime"`
	EndTime       string `json:"endtime"`
	APIVersion    string `json:"adapiversion"`
	Media         string `json:"media"`
	RequestTime   string `json:"requesttime"`
	TargetNation  string `json:"targetnation"`
	Gender        string `json:"gender"`
	Age           string `json:"age"`
	IsOnAir       string `json:"isonair"`
	IsPay         string `json:"ispay"`
	VODType       string `json:"vodtype"`
	PlayTime      string `json:"playtime"`
	AdType        string `json:"adtype"`
	Referrer      string `json:"referrer"`
	AdLink        string `json:"adlink"`
	CustomKeyword string `json:"customkeyword"`
	LogAPIVersion string `json:"logapiversion"`
	URL           string `json:"url"`
	VideoLogURL   string `json:"videologUrl"`
	AdCompanyType string `json:"adcompanytype"`
	ClipID        string `json:"clipid"`
	Site          string `json:"site"`
	Category      string `json:"category"`
	Section       string `json:"section"`
	CPID          string `json:"cpid"`
	ChannelID     string `json:"channelid"`
	ProgramID     string `json:"programid"`
	Like          string `json:"like"`
	PlayCount     string `json:"playcount"`
	FirstPlay     string `json:"firstplay"`
}

type debug struct {
	DeployState     string   `json:"process.env.deploystate"`
	Version         string   `json:"version"`
	HasUserPass     bool     `json:"hasUserPass"`
	IsFree          bool     `json:"isFree"`
	ValidResolution []string `json:"validResolution"`
}
