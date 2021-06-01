// Copyright (c) 2021 Yandex LLC. All rights reserved.
// Author: Martynov Pavel <covariance@yandex-team.ru>

package adapter

import (
	"context"
	"fmt"
	"time"

	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/jinzhu/copier"
)

type FakeAdapter struct {
	key     string
	secret  string
	storage map[string]s3.Bucket
}

func (r *FakeAdapter) checkCredentials(key, secret string) error {
	if r.key != key || r.secret != secret {
		return fmt.Errorf("credentials are incorrect")
	}
	return nil
}

func (r *FakeAdapter) Create(_ context.Context, key, secret, name string) error {
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

func (r *FakeAdapter) List(_ context.Context, key, secret string) ([]*s3.Bucket, error) {
	if err := r.checkCredentials(key, secret); err != nil {
		return nil, err
	}

	var lst []*s3.Bucket
	for i := range r.storage {
		var tmp s3.Bucket
		if err := copier.Copy(tmp, r.storage[i]); err != nil {
			return nil, err
		}
		lst = append(lst, &tmp)
	}

	return lst, nil
}

func (r *FakeAdapter) Delete(_ context.Context, key, secret, name string) error {
	if err := r.checkCredentials(key, secret); err != nil {
		return err
	}

	if _, exists := r.storage[name]; !exists {
		return fmt.Errorf("bucket does not exist: %s", name)
	}

	delete(r.storage, name)

	return nil
}
