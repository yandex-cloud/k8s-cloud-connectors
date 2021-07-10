// Copyright (c) 2021 Yandex LLC. All rights reserved.
// Author: Martynov Pavel <covariance@yandex-team.ru>

package scaffolder

import (
	"bytes"
	"fmt"
	"strings"
	"text/template"
)

// Scheme holds compiled rules for scaffolding process.
type Scheme []struct{ in, out string }

// ParseScheme compiles scheme from given file, completing it with given Values.
func ParseScheme(path string, val Values) (Scheme, error) {
	tmpl, err := template.ParseFiles(path)
	if err != nil {
		return nil, fmt.Errorf("unable to parse sceheme: %w", err)
	}

	buf := bytes.NewBufferString("")
	if err := tmpl.Execute(buf, val); err != nil {
		return nil, fmt.Errorf("unable to execute template: %w", err)
	}

	var res Scheme

	// TODO (covariance) this can be implemented in a more convenient for user way, skipping whitespaces
	for _, line := range strings.Split(buf.String(), "\n") {
		if len(line) == 0 {
			// Empty lines are ignored
			continue
		}
		if line[0] == '#' {
			// Comments are ignored
			continue
		}
		split := strings.Split(line, " => ")
		if len(split) != 2 {
			return nil, fmt.Errorf("invalid Scheme format \"%s\"", line)
		}
		res = append(res, struct{ in, out string }{in: split[0], out: split[1]})
	}

	return res, nil
}
