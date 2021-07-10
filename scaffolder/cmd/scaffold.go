// Copyright (c) 2021 Yandex LLC. All rights reserved.
// Author: Martynov Pavel <covariance@yandex-team.ru>

package cmd

import (
	"fmt"
	"os"
	"strings"

	"k8s-connectors/scaffolder/pkg/scaffolder"

	"github.com/spf13/cobra"
)

var (
	scaffoldCmd = cobra.Command{
		Use:   "scaffolder <scaffolding dir> <output dir> <scheme file> <values specification>",
		Short: "scaffolder creates scaffolding populated with specified values based on provided scheme",
		Args:  cobra.ExactArgs(4),
		RunE:  scaffold,
	}

	valuesType string
)

func scaffold(_ *cobra.Command, args []string) error {
	var val scaffolder.Values
	var err error
	if strings.ToLower(valuesType) == "inline" {
		val, err = scaffolder.ParseValuesFromString(args[3])
		if err != nil {
			return fmt.Errorf("unable to parse inline values: %w", err)
		}
	} else if strings.ToLower(valuesType) == "json" {
		val, err = scaffolder.ParseValuesFromJson(args[3])
		if err != nil {
			return fmt.Errorf("unable to parse JSON values: %w", err)
		}
	} else if strings.ToLower(valuesType) == "yaml" || strings.ToLower(valuesType) == "yaml" {
		val, err = scaffolder.ParseValuesFromYaml(args[3])
		if err != nil {
			return fmt.Errorf("unable to parse YAML values: %w", err)
		}
	} else {
		val, err = scaffolder.ParseValuesFromFile(args[3])
		if err != nil {
			return fmt.Errorf("values format unspecified, parsing was unsuccessfull: %w", err)
		}
	}

	scheme, err := scaffolder.ParseScheme(args[2], val)
	if err != nil {
		return fmt.Errorf("unable to parse scheme: %w", err)
	}

	return scaffolder.Scaffold(args[0], args[1], val, scheme)
}

func init() {
	scaffoldCmd.PersistentFlags().StringVar(
		&valuesType,
		"values-type",
		"",
		"extension of values file if it is a file or 'inline' for inline json value",
	)
}

func Execute() {
	if err := scaffoldCmd.Execute(); err != nil {
		os.Exit(1)
	}
}
