// Copyright (c) 2021 Yandex LLC. All rights reserved.
// Author: Martynov Pavel <covariance@yandex-team.ru>

package scaffolder

import (
	"fmt"
	"os"
	"path/filepath"
	"text/template"
)

func process(input, output string, values Values) error {
	processedInput, err := template.ParseFiles(input)
	if err != nil {
		return fmt.Errorf("unable to parse input file \"%s\": %w", input, err)
	}

	if err := os.MkdirAll(filepath.Dir(output), os.ModePerm); err != nil {
		return fmt.Errorf("unable to create path to output file \"%s\": %w", output, err)
	}

	out, err := os.OpenFile(output, os.O_WRONLY|os.O_CREATE, 0666)
	if err != nil {
		return fmt.Errorf("unable to open output file: \"%s\": %w", output, err)
	}
	defer func() { _ = out.Close() }()

	if err := processedInput.Execute(out, values); err != nil {
		return fmt.Errorf("unable to process template and write to \"%s\": %w", output, err)
	}

	return nil
}

// Scaffold processes all scaffolding rules from Scheme and populates scaffolding templates with given values.
func Scaffold(scaffoldingDir, outputDir string, values Values, scheme Scheme) error {
	if err := os.MkdirAll(outputDir, os.ModePerm); err != nil {
		return fmt.Errorf("unable to create output folder: %w", err)
	}

	for i, line := range scheme {
		if err := process(
			filepath.Join(scaffoldingDir, line.in), filepath.Join(outputDir, line.out), values,
		); err != nil {
			return fmt.Errorf("unable to process scheme directive %d: %w", i+1, err)
		}
	}

	return nil
}
