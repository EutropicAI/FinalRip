package oss

import (
	"bytes"
	"context"
	"io"
	"net/url"
	"sync"
	"time"

	"github.com/EutropicAI/FinalRip/module/config"
	"github.com/EutropicAI/FinalRip/module/log"
	"github.com/EutropicAI/FinalRip/module/util"
	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

var (
	oss  *minio.Client
	err  error
	once sync.Once
)

func Init() {
	once.Do(func() {
		initialize()
	})
}

func initialize() {
	// Initialize minio client object.
	oss, err = minio.New(config.OSSConfig.Endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(config.OSSConfig.AccessKey, config.OSSConfig.SecretKey, ""),
		Secure: config.OSSConfig.UseSSL,
	})
	err = oss.MakeBucket(context.TODO(), config.OSSConfig.Bucket,
		minio.MakeBucketOptions{Region: config.OSSConfig.Region, ObjectLocking: false})
	if err != nil {
		exists, _ := oss.BucketExists(context.TODO(), config.OSSConfig.Bucket)
		if !exists {
			log.Logger.Error("Failed to create bucket: " + err.Error())
		}
	}
}

// PutByPath uploads the content from a file to the oss key.
func PutByPath(key string, path string) error {
	info, err := oss.FPutObject(context.Background(), config.OSSConfig.Bucket, key, path,
		minio.PutObjectOptions{
			ContentType: "application/octet-stream",
		})
	if err != nil {
		return err
	}
	log.Logger.Infof("Uploaded %s of size: %v Successfully.", key, info.Size)
	return nil
}

// Put uploads the content from r to the oss key.
func Put(key string, reader io.Reader, objectSize int64) error {
	info, err := oss.PutObject(context.Background(), config.OSSConfig.Bucket, key, reader, objectSize,
		minio.PutObjectOptions{ContentType: "application/octet-stream"})

	log.Logger.Infof("Uploaded %s of size: %v Successfully.", key, info.Size)
	return err
}

// PutBytes uploads the byte array content to the oss key.
func PutBytes(key string, data []byte) error {
	return Put(key, bytes.NewReader(data), int64(len(data)))
}

// GetWithPath downloads and saves the object as a file in the local filesystem by key.
func GetWithPath(key string, path string) error {
	return oss.FGetObject(context.Background(), config.OSSConfig.Bucket, key, path, minio.GetObjectOptions{})
}

// Get gets the file pointed to by key.
func Get(key string) (*minio.Object, error) {
	return oss.GetObject(context.Background(), config.OSSConfig.Bucket, key, minio.GetObjectOptions{})
}

// Exist checks if the file pointed to by key exists.
func Exist(key string) bool {
	exist, err := oss.StatObject(context.Background(), config.OSSConfig.Bucket, key, minio.StatObjectOptions{})
	if err != nil {
		// log.Logger.Error("Failed to check if object exists: " + err.Error())
		return false
	}
	return exist.Size > 0
}

// Size returns the size of the file pointed to by key.
func Size(key string) (string, error) {
	stat, err := oss.StatObject(context.Background(), config.OSSConfig.Bucket, key, minio.StatObjectOptions{})
	if err != nil {
		log.Logger.Error("Failed to get object size: " + err.Error())
		return "", err
	}
	return util.ByteCountBinary(uint64(stat.Size)), nil
}

// Delete deletes the file pointed to by key.
func Delete(key string) error {
	if Exist(key) {
		return oss.RemoveObject(context.Background(), config.OSSConfig.Bucket, key, minio.RemoveObjectOptions{})
	}
	return nil
}

// GetPresignedURL gets the presigned URL for the file pointed to by key.
func GetPresignedURL(key string, fileName string, expiration time.Duration) (string, error) {
	// Set request parameters
	reqParams := make(url.Values)
	reqParams.Set("response-content-disposition", "attachment; filename="+fileName)

	// Generate presigned get object url
	presignedURL, err := oss.PresignedGetObject(context.Background(), config.OSSConfig.Bucket, key, expiration, reqParams)
	if err != nil {
		log.Logger.Error("Failed to generate presigned URL: " + key + err.Error())
		return "", err
	}
	return presignedURL.String(), nil
}

// GetUploadPresignedURL gets the presigned URL for the file upload, use PUT method in frontend.
func GetUploadPresignedURL(key string, expiration time.Duration) (string, error) {
	// Generate presigned put object url
	presignedURL, err := oss.PresignedPutObject(context.Background(), config.OSSConfig.Bucket, key, expiration)
	if err != nil {
		log.Logger.Error("Failed to generate presigned URL: " + key + err.Error())
		return "", err
	}
	return presignedURL.String(), nil
}
