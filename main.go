package main

import (
	"flag"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

var wavveObj = NewWavve()

func GetNormalHeader() map[string]string {
	return map[string]string{
		"wavve-credential": Credential,
		"accept-encoding":  "gzip",
		"user-agent":       "okhttp/5.0.0-alpha.10",
	}
}

// var HeaderPoop = map[string]string{
// 	"user-agent": "poopV2",
// }

const UrlGetChannels = "https://apis.wavve.com/cf/live/recommend-channels"
const ClientVersion = "7.0.91"
const Device = "Android"
const ModelID = ""
const Partner = "pooq"
const ApiKey = "6A87455D54481A536DFB8AD397C5EC4D"
const Region = "kor"
const DRM = "wm"
const PooqZone = "none"

var GUID = ""

var ValidToken = ""

// 鉴权中间件
func authMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {

		if ValidToken != "" {
			token := c.Query("token")

			// 验证Token
			if token != ValidToken {
				c.JSON(401, gin.H{
					"error": "Invalid or missing token",
				})
				c.Abort()
				return
			}
		}

		// Token有效，继续处理请求
		c.Next()
	}
}

// 设置路由和处理逻辑
func setupRouter() *gin.Engine {
	// 设置Gin为发布模式
	gin.SetMode(gin.ReleaseMode)
	r := gin.Default()

	// 创建需要鉴权的路由组
	authorized := r.Group("/")
	authorized.Use(authMiddleware())

	// 配置获取tv.m3u文件的路由
	authorized.GET("/", func(c *gin.Context) {
		c.Writer.Header().Set("Content-Type", "application/octet-stream")
		c.Writer.Header().Set("Content-Disposition", "attachment; filename=wavve.m3u")
		m3uStr := WavveGenerateM3U(c.Request.Host)
		if m3uStr != "" {
			c.String(200, m3uStr)
			return
		} else {
			c.String(404, "")
		}
	})

	authorized.GET("/wavve/:rid", func(c *gin.Context) {
		ts := c.Query("ts")
		cookie := c.Query("cookie")
		if ts == "" && cookie == "" {
			rid := c.Param("rid")
			contentID := strings.ReplaceAll(rid, ".m3u8", "")

			if contentID != "" {
				wavveObj.HandleMainRequest(c, contentID)
			} else {
				c.JSON(400, gin.H{"error": "Missing required parameters"})
			}
		} else if ts != "" && cookie != "" {
			if !EnableTSMemCache {
				//wavveObj.HandleTsRequest(c, ts, cookie)
				wavveObj.HandleTsRequestSF(c, ts, cookie)
			} else {
				// if EnableSingleFlight {
				// 	wavveObj.HandleTsRequestCacheSingleFlight(c, ts, cookie)
				// } else {
				// 	wavveObj.HandleTsRequestCache(c, ts, cookie)
				// }
				wavveObj.HandleTsRequestCacheSF(c, ts, cookie)
			}

		} else {
			c.JSON(404, gin.H{"error": "Bad resquest"})
		}

	})

	return r
}

func main() {
	host := flag.String("host", "0.0.0.0", "host")
	port := flag.String("p", "18090", "port")
	flag.StringVar(&ValidToken, "token", "", "Set token authentication")
	flag.BoolVar(&DebugMode, "debug", false, "Enable debug mode")
	flag.BoolVar(&EnableTSMemCache, "tsmemcache", false, "Enable TS file MemCache")
	//flag.BoolVar(&EnableSingleFlight, "singleflight", false, "Enable TS file MemCache singleflight")
	// flag.StringVar(&UserName, "username", "", "用户名")
	// flag.StringVar(&Password, "password", "", "密码")
	flag.Parse()

	creds := GetWavveMCredntials()
	if creds.GUID == "" {
		GUID = uuid.New().String()
	} else {
		GUID = creds.GUID
		UserName = creds.Username
		Password = creds.Password
	}
	if UserName != "" && Password != "" {
		WavveSign()
		SecureToken = GetWavveSecureToken(GUID)
		WavveGetPermit()
	} else {
		SecureToken = GetWavveSecureToken(GUID)
	}

	LogDebug("SecureToken:", SecureToken)
	// playUrl, awsCookie, _, _ := GetWavveLive("C4101")
	// LogDebug("playUrl:", playUrl, "awsCookie:", awsCookie)

	WavveUpdateChannels()

	// fmt.Printf("Can't watch channels: ")
	// for i, ch := range WavveChannelslistResponse.CellTopList.CellList {
	// 	//fmt.Printf("Index:%d ContentID:%s ChannelName:%s\n", i+1, ch.ContentID, ch.TitleList[0].Text)
	// 	playUrl, _, _, _ := WavveGetLive(ch.ContentID)
	// 	if playUrl == "" {
	// 		fmt.Printf("Index:%d ContentID:%s ChannelName:%s\n", i+1, ch.ContentID, ch.TitleList[0].Text)
	// 	}
	// 	time.Sleep(1 * time.Second)
	// }
	// fmt.Printf("\n")

	r := setupRouter()

	LogInfo("use -h to visit help")
	LogInfo("Listen on "+*host+":"+*port, "...")
	if ValidToken != "" {
		LogInfo("visit http://ip:" + *port + "?token=" + ValidToken + " to get m3u file!")
	} else {
		LogInfo("visit http://ip:" + *port + " to get m3u file!")
	}

	r.Run(*host + ":" + *port)
}
