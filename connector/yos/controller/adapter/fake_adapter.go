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
	key     string
	secret  string
	storage map[string]s3.Bucket
}

func NewFakeYandexObjectStorageAdapter(key, secret string) FakeYandexObjectStorageAdapter {
	return FakeYandexObjectStorageAdapter{
		key,
		secret,
		make(map[string]s3.Bucket),
	}
}

func (r *FakeYandexObjectStorageAdapter) checkCredentials(key, secret string) error {
	if r.key != key || r.secret != secret {
		return fmt.Errorf("credentials are incorrect")
	}
	return nil
}

func (r *FakeYandexObjectStorageAdapter) Create(_ context.Context, key, secret, name string) error {
	if err := r.checkCredentials(key, secret); err != nil {
		return err
	}

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

func (r *FakeYandexObjectStorageAdapter) List(_ context.Context, key, secret string) ([]*s3.Bucket, error) {
	if err := r.checkCredentials(key, secret); err != nil {
		return nil, err
	}

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

func (r *FakeYandexObjectStorageAdapter) Delete(_ context.Context, key, secret, name string) error {
	if err := r.checkCredentials(key, secret); err != nil {
		return err
	}

	if _, exists := r.storage[name]; !exists {
		return awserr.New(s3.ErrCodeNoSuchBucket, "no such bucket", nil)
	}

	delete(r.storage, name)

	return nil
}
