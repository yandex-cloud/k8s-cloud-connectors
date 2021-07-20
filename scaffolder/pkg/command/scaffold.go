// Copyright (c) 2021 Yandex LLC. All rights reserved.
// Author: Martynov Pavel <covariance@yandex-team.ru>

package command

import (
	"fmt"
	"os"
	"strings"
	"unicode"

	"github.com/yandex-cloud/k8s-cloud-connectors/scaffolder/pkg/scaffolder"

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

	groupName string
	version   string
	longName  string
	shortName string

	inlineValues []string
)

func shortNameFromLong() error {
	var builder strings.Builder
	for _, rn := range longName {
		if unicode.IsUpper(rn) {
			builder.WriteRune(rn)
		}
	}

	if builder.Len() == 0 {
		return fmt.Errorf("unable to deduce short name: long name has no capital letters")
	}

	shortName = strings.ToLower(builder.String())
	return nil
}

func scaffold(_ *cobra.Command, _ []string) error {
	val := scaffolder.Values{}

	for _, s := range inlineValues {
		k, v, err := scaffolder.ParseKeyValueFromString(s)
		if err != nil {
			return fmt.Errorf("unable to parse inline values: %w", err)
		}
		val[k] = v
	}

	if shortName == "" {
		if err := shortNameFromLong(); err != nil {
			return err
		}
	}

	val["groupName"] = groupName
	val["version"] = version
	val["longName"] = longName
	val["shortName"] = shortName

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

	scaffoldCmd.PersistentFlags().StringVar(
		&groupName,
		"group",
		"connectors.cloud.yandex.com",
		"group name of the resource, such as \"yet.another.group.com\"",
	)

	scaffoldCmd.PersistentFlags().StringVar(
		&version,
		"version",
		"v1",
		"version of the resource, such as \"v1beta1\"",
	)

	scaffoldCmd.PersistentFlags().StringVar(
		&longName,
		"name",
		"",
		"name of the resource, such as \"YetAnotherYandexResource\"",
	)
	_ = scaffoldCmd.MarkFlagRequired("name")

	scaffoldCmd.PersistentFlags().StringVar(
		&shortName,
		"short",
		"",
		"short name of the resource, such as \"yayr\", is optional",
	)

	scaffoldCmd.PersistentFlags().StringSliceVar(
		&inlineValues,
		"val",
		[]string{},
		"any additional values in \"key=value\" format",
	)
}

func Execute() {
	if err := scaffoldCmd.Execute(); err != nil {
		os.Exit(1)
	}
}
