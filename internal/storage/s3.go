package storage

import (
	"context"
	"io"

	minio "github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

type S3Storage struct {
	client *minio.Client
	bucket string
}

type S3Config struct {
	Endpoint  string `env:"ENDPOINT,required"`
	AccessKey string `env:"ACCESS_KEY_FILE,required,file,notEmpty,unset"` //nolint:gosec // false-positive: no json
	SecretKey string `env:"SECRET_KEY_FILE,required,file,notEmpty,unset"`
	UseSSL    bool   `env:"USE_SSL" envDefault:"true"`
	Bucket    string `env:"BUCKET,required"`
	Region    string `env:"REGION" envDefault:"ru-central-1"`
}

func NewS3Storage(cfg S3Config) (*S3Storage, error) {
	client, err := minio.New(cfg.Endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(cfg.AccessKey, cfg.SecretKey, ""),
		Secure: cfg.UseSSL,
		Region: "ru-central-1",
	})
	if err != nil {
		return nil, err
	}

	exists, err := client.BucketExists(context.Background(), cfg.Bucket)
	if err != nil {
		return nil, err
	}
	if !exists {
		return nil, &BucketNotFoundError{Bucket: cfg.Bucket}
	}

	return &S3Storage{client: client, bucket: cfg.Bucket}, nil
}

func (s *S3Storage) Put(key string, data io.Reader) error {
	_, err := s.client.PutObject(context.Background(), s.bucket, key, data, -1, minio.PutObjectOptions{})
	return err
}

func (s *S3Storage) List(prefix string) ([]FileInfo, error) {
	ctx := context.Background()
	objectCh := s.client.ListObjects(ctx, s.bucket, minio.ListObjectsOptions{
		Prefix:    prefix,
		Recursive: false,
	})

	var objects []FileInfo
	for obj := range objectCh {
		if obj.Err != nil {
			return nil, obj.Err
		}
		objects = append(objects, FileInfo{Key: obj.Key, Size: obj.Size})
	}
	return objects, nil
}

func (s *S3Storage) Delete(ctx context.Context, key string) error {
	return s.client.RemoveObject(ctx, s.bucket, key, minio.RemoveObjectOptions{})
}

func (s *S3Storage) Close() error {
	return nil
}

type BucketNotFoundError struct{ Bucket string }

func (e *BucketNotFoundError) Error() string {
	return "bucket not found: " + e.Bucket
}
