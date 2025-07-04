package main

import (
	"encoding/base64"
	"encoding/json"
	"time"
)

// 基于游标的分页
type Page struct {
	NextID        string `json:"next_id"`          // 下一个ID
	NextTimeAtUTC int64  `json:"next_time_at_utc"` // 下一个时间戳
	PageSize      int64  `json:"page_size"`        // 页面大小
}

type Token string

func (p *Page) Encode() Token {
	b, err := json.Marshal(p)
	if err != nil {
		return Token("")
	}
	return Token(base64.StdEncoding.EncodeToString(b))
}

func (p *Page) InValid() bool {
	return p.NextID == "" || p.NextTimeAtUTC == 0 || p.NextTimeAtUTC > time.Now().Unix() || p.PageSize <= 0
}

func (t Token) Decode() *Page {
	var result Page
	if len(t) == 0 {
		return &result
	}

	bytes, err := base64.StdEncoding.DecodeString(string(t))
	if err != nil {
		return &result
	}

	err = json.Unmarshal(bytes, &result)
	if err != nil {
		return &result
	}
	return &result
}
