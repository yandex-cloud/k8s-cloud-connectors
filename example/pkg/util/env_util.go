// Copyright (c) 2021 Yandex LLC. All rights reserved.
// Author: Martynov Pavel <covariance@yandex-team.ru>

package util

import (
	"log"
	"os"
)

func GetEnvOrDie(key string) string {
	res, ok := os.LookupEnv(key)
	if !ok {
		log.Fatal("unable to get environmental variable " + key)
	}
	return res
}
