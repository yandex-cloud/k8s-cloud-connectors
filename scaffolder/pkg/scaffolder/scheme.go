// Copyright (c) 2021 Yandex LLC. All rights reserved.
// Author: Martynov Pavel <covariance@yandex-team.ru>

package scaffolder

import (
	"bytes"
	"fmt"
	"path/filepath"
	"text/template"

	"github.com/Masterminds/sprig"
	"gopkg.in/yaml.v3"
)

// Scheme holds a set of rules for scaffolding processFile.
type Scheme struct {
	Entries []SchemeEntry `yaml:"scheme"`
}

// SchemeEntry holds one rule for scaffolding processFile.
type SchemeEntry struct {
	Source      string `yaml:"source"`
	Destination string `yaml:"destination"`
	Recursive   bool   `yaml:"recursive,omitempty"`
}

// ParseScheme compiles scheme from given file, completing it with given Values.
func ParseScheme(path string, val Values) (Scheme, error) {
	tmpl, err := template.New(filepath.Base(path)).Funcs(sprig.TxtFuncMap()).ParseFiles(path)
	if err != nil {
		return Scheme{}, fmt.Errorf("unable to parse sceheme: %w", err)
	}

	buf := bytes.NewBufferString("")
	if err := tmpl.Execute(buf, val); err != nil {
		return Scheme{}, fmt.Errorf("unable to execute template: %w", err)
	}

	res := Scheme{}
	if err := yaml.Unmarshal(buf.Bytes(), &res); err != nil {
		return Scheme{}, fmt.Errorf("unable to parse scheme yaml structure: %w", err)
	}

	return res, nil
}
