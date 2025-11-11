package storage

import (
	"bytes"
	"context"
	"inventory/backend/internal/config"
	"io"
	"mime/multipart"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

type Uploader interface {
	UploadFile(bucketName, objectName string, file *bytes.Buffer, fileSize int64) (minio.UploadInfo, error)
	UploadFileFromMultipart(bucketName, objectName string, file *multipart.FileHeader) (minio.UploadInfo, error)
	DownloadFile(bucketName, objectName string) (*minio.Object, error)
	GetFile(bucketName string, objectName string) (*minio.Object, error)
}

type MinIOUploader struct {
	Client *minio.Client
}

func NewMinIOUploader(cfg *config.Config) (*MinIOUploader, error) {
	minioClient, err := minio.New(cfg.MinioEndpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(cfg.MinioAccessKeyID, cfg.MinioSecretAccessKey, ""),
		Secure: cfg.MinioUseTLS,
	})
	if err != nil {
		return nil, err
	}

	return &MinIOUploader{Client: minioClient}, nil
}

func (u *MinIOUploader) UploadFile(bucketName, objectName string, file *bytes.Buffer, fileSize int64) (minio.UploadInfo, error) {
	ctx := context.Background()
	exists, err := u.Client.BucketExists(ctx, bucketName)
	if err != nil {
		return minio.UploadInfo{}, err
	}
	if !exists {
		err = u.Client.MakeBucket(ctx, bucketName, minio.MakeBucketOptions{})
		if err != nil {
			return minio.UploadInfo{}, err
		}
	}

	return u.Client.PutObject(ctx, bucketName, objectName, file, fileSize, minio.PutObjectOptions{
		ContentType: "application/octet-stream",
	})
}

func (u *MinIOUploader) UploadFileFromMultipart(bucketName, objectName string, file *multipart.FileHeader) (minio.UploadInfo, error) {
	src, err := file.Open()
	if err != nil {
		return minio.UploadInfo{}, err
	}
	defer src.Close()

	ctx := context.Background()
	exists, err := u.Client.BucketExists(ctx, bucketName)
	if err != nil {
		return minio.UploadInfo{}, err
	}
	if !exists {
		err = u.Client.MakeBucket(ctx, bucketName, minio.MakeBucketOptions{})
		if err != nil {
			return minio.UploadInfo{}, err
		}
	}

	return u.Client.PutObject(ctx, bucketName, objectName, src, file.Size, minio.PutObjectOptions{
		ContentType: file.Header.Get("Content-Type"),
	})
}

func (u *MinIOUploader) DownloadFile(bucketName, objectName string) (*minio.Object, error) {
	return u.Client.GetObject(context.Background(), bucketName, objectName, minio.GetObjectOptions{})
}

func (u *MinIOUploader) GetFile(bucketName string, objectName string) (*minio.Object, error) {
	object, err := u.Client.GetObject(context.Background(), bucketName, objectName, minio.GetObjectOptions{})
	if err != nil {
		return nil, err
	}
	return object, nil
}

func (u *MinIOUploader) GetFileContent(bucketName string, objectName string) ([]byte, error) {
	object, err := u.GetFile(bucketName, objectName)
	if err != nil {
		return nil, err
	}
	defer object.Close()

	buffer := new(bytes.Buffer)
	if _, err := io.Copy(buffer, object); err != nil {
		return nil, err
	}

	return buffer.Bytes(), nil
}
