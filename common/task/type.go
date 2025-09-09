package task

import (
	"github.com/EutropicAI/FinalRip/common/db"
)

// task types, note that video: can get video:xxx:xxx format task type
const (
	VIDEO_CUT    = "video:cut"
	VIDEO_ENCODE = "video:encode"
	VIDEO_MERGE  = "video:merge"
)

const (
	TASK_STATUS_PENDING   = "pending"
	TASK_STATUS_RUNNING   = "running"
	TASK_STATUS_COMPLETED = "completed"
)

// CutTaskPayload is a struct that represents the payload for cut task.
type CutTaskPayload struct {
	VideoKey string `json:"video_key"`
	Slice    bool   `json:"slice"`
}

// EncodeTaskPayload is a struct that represents the payload for encode task.
type EncodeTaskPayload struct {
	EncodeParam string           `json:"encode_param"`
	Script      string           `json:"script"`
	Clip        db.VideoClipInfo `json:"clip"`
	Retry       bool             `json:"retry"`
}

// MergeTaskPayload is a struct that represents the payload for merge task.
type MergeTaskPayload struct {
	VideoKey string `json:"video_key"`
}
