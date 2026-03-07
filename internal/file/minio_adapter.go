package file

import (
	"context"
	"net/url"
	"time"

	"github.com/minio/minio-go/v7"
)

type MinIOAdapter struct {
	client *minio.Client
}

func NewMinIOAdapter(client *minio.Client) *MinIOAdapter {
	return &MinIOAdapter{client: client}
}

func (a *MinIOAdapter) BucketExists(ctx context.Context, bucketName string) (bool, error) {
	return a.client.BucketExists(ctx, bucketName)
}

func (a *MinIOAdapter) MakeBucket(ctx context.Context, bucketName string, opts minio.MakeBucketOptions) error {
	return a.client.MakeBucket(ctx, bucketName, opts)
}

func (a *MinIOAdapter) StatObject(ctx context.Context, bucketName, objectName string, opts minio.StatObjectOptions) (minio.ObjectInfo, error) {
	return a.client.StatObject(ctx, bucketName, objectName, opts)
}

func (a *MinIOAdapter) RemoveObject(ctx context.Context, bucketName, objectName string, opts minio.RemoveObjectOptions) error {
	return a.client.RemoveObject(ctx, bucketName, objectName, opts)
}

func (a *MinIOAdapter) PresignedPutObject(ctx context.Context, bucketName, objectName string, expires time.Duration) (*url.URL, error) {
	return a.client.PresignedPutObject(ctx, bucketName, objectName, expires)
}

func (a *MinIOAdapter) PresignedGetObject(ctx context.Context, bucketName, objectName string, expires time.Duration, reqParams url.Values) (*url.URL, error) {
	return a.client.PresignedGetObject(ctx, bucketName, objectName, expires, reqParams)
}
