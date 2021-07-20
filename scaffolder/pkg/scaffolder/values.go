// Copyright (c) 2021 Yandex LLC. All rights reserved.
// Author: Martynov Pavel <covariance@yandex-team.ru>

package scaffolder

import (
	"fmt"
	"strings"
)

// Values is a container with strings that are to be substituted in scaffolding and scheme.
type Values map[string]string

func ParseKeyValueFromString(contents string) (key, value string, err error) {
	split := strings.SplitN(contents, "=", 2)
	if len(split) == 1 {
		if strings.HasPrefix(contents, "=") {
			err = fmt.Errorf("invalid \"key=value\" pair provided: %s", contents)
			return
		}
		key, value = split[0], ""
		return
	}
	key, value = split[0], split[1]
	return
}
