package db

import (
	"context"
	"errors"
	"sort"

	"github.com/EutropicAI/FinalRip/module/db"
	"go.mongodb.org/mongo-driver/bson"
)

// CheckVideoExist 检查视频记录是否存在
func CheckVideoExist(info VideoClipInfo) bool {
	coll := db.DB.Collection(VIDEO_COLLECTION)
	count, _ := coll.CountDocuments(context.TODO(), info)
	return count > 0
}

// InsertVideo 插入视频信息
func InsertVideo(info VideoClipInfo) error {
	coll := db.DB.Collection(VIDEO_COLLECTION)
	_, err := coll.InsertOne(context.TODO(), info)
	return err
}

// GetVideoClip 获取视频切片信息
func GetVideoClip(info VideoClipInfo) (VideoClipInfo, error) {
	coll := db.DB.Collection(VIDEO_COLLECTION)

	var res VideoClipInfo
	err := coll.FindOne(context.TODO(), info).Decode(&res)
	if err != nil {
		return VideoClipInfo{}, err
	}

	return res, nil
}

// GetVideoClips 获取所有视频切片信息，按照索引排序
func GetVideoClips(videoKey string) ([]VideoClipInfo, error) {
	coll := db.DB.Collection(VIDEO_COLLECTION)

	cursor, err := coll.Find(context.TODO(), VideoClipInfo{Key: videoKey})
	if err != nil {
		return nil, err
	}

	infos := make([]VideoClipInfo, 0)
	if err = cursor.All(context.TODO(), &infos); err != nil {
		return nil, err
	}

	if len(infos) > 0 && len(infos) != infos[0].Total {
		return nil, errors.New("video clip not complete")
	}

	sort.Slice(infos, func(i, j int) bool {
		return infos[i].Index < infos[j].Index
	})

	return infos, nil
}

// UpdateVideoClip 更新视频信息
func UpdateVideoClip(filter VideoClipInfo, update VideoClipInfo) error {
	coll := db.DB.Collection(VIDEO_COLLECTION)

	up := bson.D{{"$set", update}} //nolint: govet

	_, err := coll.UpdateOne(context.TODO(), filter, up)
	if err != nil {
		return err
	}
	return nil
}

// DeleteVideoClips 删除所有视频切片
func DeleteVideoClips(videoKey string) error {
	coll := db.DB.Collection(VIDEO_COLLECTION)
	_, err := coll.DeleteMany(context.TODO(), VideoClipInfo{Key: videoKey})
	return err
}

type VideoProgressITEM struct {
	Completed bool   `json:"completed"`
	Index     int    `json:"index"`
	ClipKey   string `json:"clip_key"`
	ClipURL   string `json:"clip_url"`
	EncodeKey string `json:"encode_key"`
	EncodeURL string `json:"encode_url"`
}

// GetVideoProgress 获取视频处理进度和每个切片的状态
func GetVideoProgress(videoKey string) ([]VideoProgressITEM, error) {
	infos, err := GetVideoClips(videoKey)
	if err != nil {
		return nil, err
	}

	status := make([]VideoProgressITEM, 0)
	for _, info := range infos {
		status = append(status, VideoProgressITEM{
			Completed: info.EncodeKey != "",
			Index:     info.Index,
			ClipKey:   info.ClipKey,
			ClipURL:   "",
			EncodeKey: info.EncodeKey,
			EncodeURL: "",
		})
	}

	return status, nil
}

// UnsetVideoClipEncodeKey 清除视频切片的编码键
func UnsetVideoClipEncodeKey(info VideoClipInfo) error {
	coll := db.DB.Collection(VIDEO_COLLECTION)
	_, err := coll.UpdateOne(context.TODO(), info, bson.D{{"$unset", bson.D{{"encode_key", ""}}}}) //nolint:govet
	return err
}
