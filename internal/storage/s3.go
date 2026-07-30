package storage

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/king12-D/cligy/internal/model"
)

type S3Storage struct {
	client *s3.Client
	bucket string
	prefix string
}

func NewS3Storage(client *s3.Client, bucket, prefix string) *S3Storage {
	return &S3Storage{client: client, bucket: bucket, prefix: prefix}
}

func (s *S3Storage) objKey(id string) string {
	if s.prefix != "" {
		return s.prefix + "/" + id
	}
	return id
}

func (s *S3Storage) metaKey(id string) string {
	return s.objKey(id) + ".meta"
}

func (s *S3Storage) Save(name, contentType string, reader io.Reader) (*model.FileInfo, error) {
	id, err := generateID()
	if err != nil {
		return nil, err
	}

	data, err := io.ReadAll(reader)
	if err != nil {
		return nil, fmt.Errorf("read upload: %w", err)
	}

	_, err = s.client.PutObject(context.Background(), &s3.PutObjectInput{
		Bucket:      aws.String(s.bucket),
		Key:         aws.String(s.objKey(id)),
		Body:        bytes.NewReader(data),
		ContentType: aws.String(contentType),
	})
	if err != nil {
		return nil, fmt.Errorf("s3 put: %w", err)
	}

	info := &model.FileInfo{
		ID:          id,
		Name:        name,
		Size:        int64(len(data)),
		ContentType: contentType,
		CreatedAt:   time.Now(),
	}
	info.GenerateETag()

	metaData, err := json.Marshal(info)
	if err != nil {
		s.Delete(id)
		return nil, fmt.Errorf("marshal meta: %w", err)
	}

	_, err = s.client.PutObject(context.Background(), &s3.PutObjectInput{
		Bucket: aws.String(s.bucket),
		Key:    aws.String(s.metaKey(id)),
		Body:   bytes.NewReader(metaData),
	})
	if err != nil {
		s.Delete(id)
		return nil, fmt.Errorf("s3 put meta: %w", err)
	}

	return info, nil
}

func (s *S3Storage) Get(id string) (io.ReadCloser, *model.FileInfo, error) {
	info, err := s.getMeta(id)
	if err != nil {
		return nil, nil, err
	}

	out, err := s.client.GetObject(context.Background(), &s3.GetObjectInput{
		Bucket: aws.String(s.bucket),
		Key:    aws.String(s.objKey(id)),
	})
	if err != nil {
		return nil, nil, fmt.Errorf("s3 get: %w", err)
	}

	return out.Body, info, nil
}

func (s *S3Storage) Delete(id string) error {
	_, err := s.client.DeleteObject(context.Background(), &s3.DeleteObjectInput{
		Bucket: aws.String(s.bucket),
		Key:    aws.String(s.objKey(id)),
	})
	if err != nil {
		return fmt.Errorf("s3 delete: %w", err)
	}

	_, err = s.client.DeleteObject(context.Background(), &s3.DeleteObjectInput{
		Bucket: aws.String(s.bucket),
		Key:    aws.String(s.metaKey(id)),
	})
	if err != nil {
		return fmt.Errorf("s3 delete meta: %w", err)
	}

	return nil
}

func (s *S3Storage) List() ([]*model.FileInfo, error) {
	var prefix *string
	if s.prefix != "" {
		prefix = aws.String(s.prefix + "/")
	}

	out, err := s.client.ListObjectsV2(context.Background(), &s3.ListObjectsV2Input{
		Bucket: aws.String(s.bucket),
		Prefix: prefix,
	})
	if err != nil {
		return nil, fmt.Errorf("s3 list: %w", err)
	}

	var files []*model.FileInfo
	metaKeys := make(map[string]bool)

	for _, obj := range out.Contents {
		key := *obj.Key
		if len(key) > 5 && key[len(key)-5:] == ".meta" {
			metaKeys[key[:len(key)-5]] = true
			continue
		}
	}

	for _, obj := range out.Contents {
		key := *obj.Key
		if len(key) > 5 && key[len(key)-5:] == ".meta" {
			continue
		}
		if metaKeys[key] {
			info, err := s.getMetaFromKey(key)
			if err == nil {
				files = append(files, info)
				continue
			}
		}
		files = append(files, &model.FileInfo{
			ID:   key,
			Name: key,
			Size: *obj.Size,
		})
	}

	return files, nil
}

func (s *S3Storage) getMeta(id string) (*model.FileInfo, error) {
	return s.getMetaFromKey(s.metaKey(id))
}

func (s *S3Storage) getMetaFromKey(key string) (*model.FileInfo, error) {
	out, err := s.client.GetObject(context.Background(), &s3.GetObjectInput{
		Bucket: aws.String(s.bucket),
		Key:    aws.String(key),
	})
	if err != nil {
		return nil, fmt.Errorf("s3 get meta: %w", err)
	}
	defer out.Body.Close()

	var info model.FileInfo
	if err := json.NewDecoder(out.Body).Decode(&info); err != nil {
		return nil, fmt.Errorf("decode meta: %w", err)
	}
	return &info, nil
}
