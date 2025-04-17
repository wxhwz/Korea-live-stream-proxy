package main

// 定义结构体变量
var WavveChannelslistResponse wavveChannelslistResponse

// wavveChannelslistResponse 根结构体
type wavveChannelslistResponse struct {
	Type        string      `json:"type"`
	SubType     string      `json:"sub_type"`
	GenTime     string      `json:"gen_time"`
	Version     string      `json:"version"`
	CellTopList cellTopList `json:"cell_toplist"`
	Filter      filter      `json:"filter"`
}

// cellTopList 顶部列表
type cellTopList struct {
	TitleList     []titleListItem `json:"title_list"`
	BgColor       string          `json:"bgcolor"`
	TopButtonList []interface{}   `json:"topbuttonlist"`
	PageCount     string          `json:"pagecount"`
	Count         string          `json:"count"`
	CellType      string          `json:"cell_type"`
	CellList      []cellListItem  `json:"celllist"`
}

// titleListItem 标题列表项
type titleListItem struct {
	Icon    string `json:"icon"`
	Text    string `json:"text"`
	MaxLine string `json:"maxline,omitempty"`
}

// cellListItem 单元列表项
type cellListItem struct {
	Thumbnail     string          `json:"thumbnail"`
	ForceRefresh  string          `json:"force_refresh"`
	TopTagList    []string        `json:"top_taglist"`
	Age           string          `json:"age"`
	AgeTag        string          `json:"age_tag"`
	BottomTagList []interface{}   `json:"bottom_taglist"`
	Time          string          `json:"time"`
	IsZzim        string          `json:"iszzim"`
	Progress      string          `json:"progress"`
	Rank          string          `json:"rank"`
	RankThumbnail string          `json:"rank_thumbnail"`
	ContentID     string          `json:"contentid"`
	TitleList     []titleListItem `json:"title_list"`
	EventList     []eventListItem `json:"event_list"`
}

// eventListItem 事件列表项
type eventListItem struct {
	Type            string                 `json:"type"`
	URL             string                 `json:"url"`
	Method          string                 `json:"method"`
	BodyList        []string               `json:"bodylist,omitempty"`
	BodyJSON        map[string]interface{} `json:"bodyjson,omitempty"`
	AddCommonParams string                 `json:"add_common_params"`
	AddCredential   string                 `json:"add_credential"`
}

// filter 过滤器
type filter struct {
	BaseAPI          string       `json:"baseapi"`
	DefaultAPIParams string       `json:"default_api_parameters"`
	AddCommonParams  string       `json:"add_common_params"`
	AddCredential    string       `json:"add_credential"`
	FilterList       []filterList `json:"filterlist"`
}

// filterList 过滤器列表
type filterList struct {
	Title          string           `json:"title"`
	FilterItemList []filterItemList `json:"filter_item_list"`
}

// filterItemList 过滤器项
type filterItemList struct {
	Title     string `json:"title"`
	APIParams string `json:"api_parameters"`
	Adult     string `json:"adult"`
}
