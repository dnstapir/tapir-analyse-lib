package common

import (
	"time"
)

type TapirObs struct {
	SrcName   string          `json:"src_name"`
	Creator   string          `json:"creator"`
	MsgType   string          `json:"msg_type"`
	ListType  string          `json:"list_type"`
	Added     []FlaggedDomain `json:"added"`
	Removed   []FlaggedDomain `json:"removed"`
	Msg       string          `json:"msg"`
	TimeStamp time.Time       `json:"timestamp"`
	TimeStr   string          `json:"time_str"`
}

type FlaggedDomain struct {
	Name         string    `json:"name"`
	TimeAdded    time.Time `json:"time_added"`
	TTL          int       `json:"ttl"`
	TagMask      uint32    `json:"tag_mask"`
	ExtendedTags []string  `json:"extended_tags"`
}
