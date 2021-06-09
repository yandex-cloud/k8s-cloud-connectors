// Copyright (c) 2021 Yandex LLC. All rights reserved.
// Author: Martynov Pavel <covariance@yandex-team.ru>

package adapter

import (
	"context"
	"fmt"
	"strings"

	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/service/sqs"
)

// TODO (covariance) check attributes in Create and Update for correctness

type FakeYandexMessageQueueAdapter struct {
	key        string
	secret     string
	attributes map[string]map[string]*string
}

func NewFakeYandexMessageQueueAdapter(key, secret string) FakeYandexMessageQueueAdapter {
	return FakeYandexMessageQueueAdapter{
		key:        key,
		secret:     secret,
		attributes: map[string]map[string]*string{},
	}
}

const (
	prefix = "https://message-queue.api.cloud.yandex.net/"
	suffix = "/url/sqs"
)

func formURL(name string) string {
	return prefix + name + suffix
}

func getName(url string) (string, error) {
	if !strings.HasPrefix(url, prefix) || !strings.HasSuffix(url, suffix) {
		return "", fmt.Errorf("malformed url")
	}
	return strings.TrimSuffix(strings.TrimPrefix(url, prefix), suffix), nil
}

func (r *FakeYandexMessageQueueAdapter) checkCredentials(key, secret string) error {
	if r.key != key || r.secret != secret {
		return fmt.Errorf("credentials are incorrect")
	}
	return nil
}

func (r *FakeYandexMessageQueueAdapter) checkAttr(lhs, rhs map[string]*string) bool {
	checkIn := func(lhs, rhs map[string]*string) bool {
		for k, v := range lhs {
			if otherV, exists := rhs[k]; !exists || *v != *otherV {
				return false
			}
		}
		return true
	}
	return checkIn(lhs, rhs) && checkIn(rhs, lhs)
}

func (r *FakeYandexMessageQueueAdapter) Create(
	_ context.Context, key, secret string, attributes map[string]*string, name string,
) (string, error) {
	if err := r.checkCredentials(key, secret); err != nil {
		return "", err
	}
	if _, exists := r.attributes[name]; exists {
		if r.checkAttr(attributes, r.attributes[name]) {
			return formURL(name), nil
		}
		return "", awserr.New(sqs.ErrCodeQueueNameExists, "non-matching attributes for YMQ with same name", nil)
	}

	attrs := make(map[string]*string)
	for k, v := range attributes {
		tmp := *v
		attrs[k] = &tmp
	}
	r.attributes[name] = attrs
	return formURL(name), nil
}

func (r *FakeYandexMessageQueueAdapter) GetURL(_ context.Context, key, secret, queueName string) (string, error) {
	if err := r.checkCredentials(key, secret); err != nil {
		return "", err
	}

	if _, exists := r.attributes[queueName]; !exists {
		return "", awserr.New(sqs.ErrCodeQueueDoesNotExist, "no such queue", nil)
	}

	return formURL(queueName), nil
}

func (r *FakeYandexMessageQueueAdapter) GetAttributes(
	_ context.Context, key, secret, queueURL string,
) (map[string]*string, error) {
	if err := r.checkCredentials(key, secret); err != nil {
		return nil, err
	}
	name, err := getName(queueURL)
	if err != nil {
		return nil, err
	}
	if res, exists := r.attributes[name]; exists {
		tmp := make(map[string]*string)
		for k, v := range res {
			s := *v
			tmp[k] = &s
		}
		return tmp, nil
	}
	return nil, awserr.New(sqs.ErrCodeQueueDoesNotExist, "no such queue", nil)
}

func (r *FakeYandexMessageQueueAdapter) List(_ context.Context, key, secret string) ([]*string, error) {
	if err := r.checkCredentials(key, secret); err != nil {
		return nil, err
	}
	res := make([]*string, 0)
	for name := range r.attributes {
		url := formURL(name)
		res = append(res, &url)
	}
	return res, nil
}

func (r *FakeYandexMessageQueueAdapter) UpdateAttributes(
	_ context.Context, key, secret string, attributes map[string]*string, queueURL string,
) error {
	if err := r.checkCredentials(key, secret); err != nil {
		return err
	}
	name, err := getName(queueURL)
	if err != nil {
		return err
	}

	if _, exists := r.attributes[name]; !exists {
		return awserr.New(sqs.ErrCodeQueueDoesNotExist, "no such queue", nil)
	}

	tmp := make(map[string]*string)
	for k, v := range attributes {
		s := *v
		tmp[k] = &s
	}

	r.attributes[name] = tmp

	return nil
}

func (r *FakeYandexMessageQueueAdapter) Delete(_ context.Context, key, secret, queueURL string) error {
	if err := r.checkCredentials(key, secret); err != nil {
		return err
	}
	name, err := getName(queueURL)
	if err != nil {
		return err
	}
	if _, exists := r.attributes[name]; !exists {
		return awserr.New(sqs.ErrCodeQueueDoesNotExist, "no such queue", nil)
	}
	delete(r.attributes, name)
	return nil
}
