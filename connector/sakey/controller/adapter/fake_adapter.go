// Copyright (c) 2021 Yandex LLC. All rights reserved.
// Author: Martynov Pavel <covariance@yandex-team.ru>

package adapter

import (
	"context"
	"strconv"

	"github.com/yandex-cloud/go-genproto/yandex/cloud/iam/v1/awscompatibility"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type FakeStaticAccessKeyAdapter struct {
	Storage map[string]*awscompatibility.AccessKey
	FreeID  int
}

func NewFakeStaticAccessKeyAdapter() FakeStaticAccessKeyAdapter {
	return FakeStaticAccessKeyAdapter{
		Storage: map[string]*awscompatibility.AccessKey{},
		FreeID:  0,
	}
}

func (r *FakeStaticAccessKeyAdapter) Create(
	_ context.Context, saID, description string,
) (*awscompatibility.CreateAccessKeyResponse, error) {
	res := &awscompatibility.CreateAccessKeyResponse{
		AccessKey: &awscompatibility.AccessKey{
			Id:               strconv.Itoa(r.FreeID),
			ServiceAccountId: saID,
			CreatedAt:        timestamppb.Now(),
			Description:      description,
			KeyId:            strconv.Itoa(r.FreeID),
		},
		Secret: strconv.Itoa(r.FreeID),
	}
	r.Storage[strconv.Itoa(r.FreeID)] = res.AccessKey
	r.FreeID++
	return res, nil
}

func (r *FakeStaticAccessKeyAdapter) Read(_ context.Context, keyID string) (*awscompatibility.AccessKey, error) {
	if _, ok := r.Storage[keyID]; !ok {
		return nil, status.Errorf(codes.NotFound, "key not found: "+keyID)
	}
	return r.Storage[keyID], nil
}

func (r *FakeStaticAccessKeyAdapter) Delete(_ context.Context, sakeyID string) error {
	if _, ok := r.Storage[sakeyID]; !ok {
		return status.Errorf(codes.NotFound, "key not found: "+sakeyID)
	}
	delete(r.Storage, sakeyID)
	return nil
}

func (r *FakeStaticAccessKeyAdapter) List(_ context.Context, saID string) ([]*awscompatibility.AccessKey, error) {
	list := []*awscompatibility.AccessKey{}
	for _, key := range r.Storage {
		if key.ServiceAccountId == saID {
			list = append(list, key)
		}
	}
	return list, nil
}
