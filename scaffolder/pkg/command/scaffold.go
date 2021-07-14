// Copyright (c) 2021 Yandex LLC. All rights reserved.
// Author: Martynov Pavel <covariance@yandex-team.ru>

package command

import (
	"fmt"
	"os"

	"k8s-connectors/scaffolder/pkg/scaffolder"

	"github.com/spf13/cobra"
)

var (
	scaffoldCmd = cobra.Command{
		Use:   "scaffolder",
		Short: "scaffolder creates scaffolding populated with specified values based on provided scheme",
		Args:  cobra.NoArgs,
		RunE:  scaffold,
	}

	templatesDir string
	outputDir    string
	scheme       string

	jsonValues   []string
	yamlValues   []string
	fileValues   []string
	inlineValues []string
)

func appendToValuesFromMappedSlice(
	val *scaffolder.Values, slice []string, mapper func(string) (scaffolder.Values, error),
) error {
	for _, s := range slice {
		appendant, err := mapper(s)
		if err != nil {
			return err
		}
		for k, v := range appendant {
			(*val)[k] = v
		}
	}

	return nil
}

func scaffold(_ *cobra.Command, _ []string) error {
	val := scaffolder.Values{}

	if err := appendToValuesFromMappedSlice(&val, jsonValues, scaffolder.ParseValuesFromJSON); err != nil {
		return fmt.Errorf("unable to parse JSON values: %w", err)
	}

	if err := appendToValuesFromMappedSlice(&val, yamlValues, scaffolder.ParseValuesFromYAML); err != nil {
		return fmt.Errorf("unable to parse YAML values: %w", err)
	}

	if err := appendToValuesFromMappedSlice(&val, fileValues, scaffolder.ParseValuesFromFile); err != nil {
		return fmt.Errorf("unable to parse file values: %w", err)
	}

	if err := appendToValuesFromMappedSlice(&val, inlineValues, scaffolder.ParseValuesFromString); err != nil {
		return fmt.Errorf("unable to parse inline values: %w", err)
	}

	scheme, err := scaffolder.ParseScheme(scheme, val)
	if err != nil {
		return fmt.Errorf("unable to parse scheme: %w", err)
	}

	if err := scaffolder.Scaffold(templatesDir, outputDir, val, scheme); err != nil {
		return fmt.Errorf("unable to perform scaffolding: %w", err)
	}

	return nil
}

func init() {
	scaffoldCmd.PersistentFlags().StringVar(
		&templatesDir,
		"templates-dir",
		"templates",
		"sets directory with scaffolding templates",
	)

	scaffoldCmd.PersistentFlags().StringVar(
		&outputDir,
		"output",
		"output",
		"sets custom output directory",
	)

	scaffoldCmd.PersistentFlags().StringVar(
		&scheme,
		"scheme",
		"scheme",
		"sets scheme for this scaffolding",
	)

	scaffoldCmd.PersistentFlags().StringSliceVar(
		&jsonValues,
		"values-from-json",
		[]string{},
		"set value files that are to be parsed as json",
	)

	scaffoldCmd.PersistentFlags().StringSliceVar(
		&yamlValues,
		"values-from-yaml",
		[]string{},
		"set value files that are to be parsed as yaml",
	)

	scaffoldCmd.PersistentFlags().StringSliceVar(
		&fileValues,
		"values-from-file",
		[]string{},
		"set value files that are to be parsed as either json or yaml, depending on file extension",
	)

	scaffoldCmd.PersistentFlags().StringSliceVar(
		&inlineValues,
		"val",
		[]string{},
		"inlined values in \"key=value\" format",
	)
}

func Execute() {
	if err := scaffoldCmd.Execute(); err != nil {
		os.Exit(1)
	}
}
