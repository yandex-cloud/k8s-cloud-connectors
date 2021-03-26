// Copyright (c) 2021 Yandex LLC. All rights reserved.
// Author: Martynov Pavel <covariance@yandex-team.ru>

package commons

import "fmt"

type ResourceNotFoundError struct {
	ResourceId string `json:"resource_id"`
	FolderId   string `json:"folder_id"`
}

func (e ResourceNotFoundError) Error() string {
	return fmt.Sprintf("resource in folder %s and id %s not found", e.ResourceId, e.FolderId)
}

func IsResourceNotFoundError(err error) bool {
	_, ok := err.(ResourceNotFoundError)
	return ok
}
