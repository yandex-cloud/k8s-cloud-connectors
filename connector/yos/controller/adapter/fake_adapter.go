// Copyright (c) 2021 Yandex LLC. All rights reserved.
// Author: Martynov Pavel <covariance@yandex-team.ru>

package adapter

import (
	"context"
	"fmt"
	"time"

	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/service/s3"
)

type FakeYandexObjectStorageAdapter struct {
	storage map[string]s3.Bucket
}

func NewFakeYandexObjectStorageAdapter() YandexObjectStorageAdapter {
	return &FakeYandexObjectStorageAdapter{
		make(map[string]s3.Bucket),
	}
}

func (r *FakeYandexObjectStorageAdapter) Create(_ context.Context, _ *s3.S3, name string) error {
	if _, exists := r.storage[name]; exists {
		return fmt.Errorf("bucket %s already exists", name)
	}

	creationDate := time.Now()
	r.storage[name] = s3.Bucket{
		CreationDate: &creationDate,
		Name:         &name,
	}

	return nil
}

func (r *FakeYandexObjectStorageAdapter) List(_ context.Context, _ *s3.S3) ([]*s3.Bucket, error) {
	var lst []*s3.Bucket
	for _, v := range r.storage {
		tmp := s3.Bucket{
			CreationDate: v.CreationDate,
			Name:         v.Name,
		}
		lst = append(lst, &tmp)
	}

	return lst, nil
}

func (r *FakeYandexObjectStorageAdapter) Delete(_ context.Context, _ *s3.S3, name string) error {
	if _, exists := r.storage[name]; !exists {
		return awserr.New(s3.ErrCodeNoSuchBucket, "no such bucket", nil)
	}

	delete(r.storage, name)

	return nil
}
