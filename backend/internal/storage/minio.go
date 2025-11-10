package storage

import (
	"bytes"
	"context"
	"inventory/backend/internal/config"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

type Uploader interface {
	UploadFile(bucketName, objectName string, file *bytes.Buffer, fileSize int64) (minio.UploadInfo, error)
}

type MinIOUploader struct {
	client *minio.Client
}

func NewMinIOUploader(cfg *config.Config) (*MinIOUploader, error) {
	minioClient, err := minio.New(cfg.MinioEndpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(cfg.MinioAccessKeyID, cfg.MinioSecretAccessKey, ""),
		Secure: false,
	})
	if err != nil {
		return nil, err
	}

	return &MinIOUploader{client: minioClient}, nil
}

func (u *MinIOUploader) UploadFile(bucketName, objectName string, file *bytes.Buffer, fileSize int64) (minio.UploadInfo, error) {
	ctx := context.Background()
	exists, err := u.client.BucketExists(ctx, bucketName)
	if err != nil {
		return minio.UploadInfo{}, err
	}
	if !exists {
		err = u.client.MakeBucket(ctx, bucketName, minio.MakeBucketOptions{})
		if err != nil {
			return minio.UploadInfo{}, err
		}
	}

	return u.client.PutObject(ctx, bucketName, objectName, file, fileSize, minio.PutObjectOptions{
		ContentType: "application/octet-stream",
	})
}
