// Copyright (c) 2021 Yandex LLC. All rights reserved.
// Author: Martynov Pavel <covariance@yandex-team.ru>

package adapter

import (
	"context"
	"fmt"
	"strings"
)

type FakeAdapter struct {
	key        string
	secret     string
	attributes map[string]map[string]*string
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

func (r *FakeAdapter) checkCredentials(key, secret string) error {
	if r.key != key || r.secret != secret {
		return fmt.Errorf("credentials are incorrect")
	}
	return nil
}

func (r *FakeAdapter) checkAttr(lhs, rhs map[string]*string) bool {
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

func (r *FakeAdapter) Create(
	_ context.Context, key, secret string, attributes map[string]*string, name string,
) (string, error) {
	if err := r.checkCredentials(key, secret); err != nil {
		return "", err
	}
	if _, exists := r.attributes[name]; exists {
		if r.checkAttr(attributes, r.attributes[name]) {
			return formURL(name), nil
		}
		return "", fmt.Errorf("non-matching attributes for YMQ with same name")
	}

	attrs := make(map[string]*string)
	for k, v := range attributes {
		tmp := *v
		attrs[k] = &tmp
	}
	r.attributes[name] = attrs
	return formURL(name), nil
}

func (r *FakeAdapter) List(_ context.Context, key, secret string) ([]*string, error) {
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

func (r *FakeAdapter) Delete(_ context.Context, key, secret, queueURL string) error {
	if err := r.checkCredentials(key, secret); err != nil {
		return err
	}
	name, err := getName(queueURL)
	if err != nil {
		return err
	}
	delete(r.attributes, name)
	return nil
}
