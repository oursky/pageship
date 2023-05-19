package storage

import (
	"context"
	"io"

	"gocloud.dev/blob"
	_ "gocloud.dev/blob/azureblob"
	_ "gocloud.dev/blob/fileblob"
	_ "gocloud.dev/blob/gcsblob"
	_ "gocloud.dev/blob/memblob"
	_ "gocloud.dev/blob/s3blob"
)

type Storage struct {
	bucket *blob.Bucket
}

func New(ctx context.Context, uri string) (*Storage, error) {
	bucket, err := blob.OpenBucket(ctx, uri)
	if err != nil {
		return nil, err
	}
	return &Storage{bucket: bucket}, nil
}

func (s *Storage) Close() error {
	return s.bucket.Close()
}

func (s *Storage) Upload(ctx context.Context, key string, r io.Reader) (err error) {
	writer, err := s.bucket.NewWriter(ctx, key, nil)
	if err != nil {
		return err
	}

	defer func() {
		if cerr := writer.Close(); cerr != nil && err == nil {
			err = cerr
		}
	}()

	_, err = io.Copy(writer, r)
	return
}
