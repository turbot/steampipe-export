package main

import (
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/turbot/steampipe-plugin-aws/aws"
	"github.com/turbot/steampipe-plugin-sdk/v5/plugin"
)

var pluginServer = plugin.Server(&plugin.ServeOpts{
	PluginFunc: aws.Plugin,
})

var pluginAlias = "aws"

var connection = pluginAlias

func setupRootCommand() *cobra.Command {
	rootCmd := &cobra.Command{
		Use:   "steampipe_export_aws TABLE_NAME [flags]",
		Short: "Steampipe export aws",
		Long: `Export data using the aws plugin.

Find detailed usage information including table names, column names, and 
examples at the Steampipe Hub: https://hub.steampipe.io/plugins/turbot/aws
`,
		Run:     executeCommand,
		Args:    cobra.ExactArgs(1),
		Version: viper.GetString("main.version"),
	}

	// Define flags
	rootCmd.PersistentFlags().String("config", "", "Connection config data")
	rootCmd.PersistentFlags().String("limiter", "", "Plugin config data")
	rootCmd.PersistentFlags().StringArray("where", []string{}, "Query 'where' clause")
	rootCmd.PersistentFlags().String("output", "csv", "Output format: csv, json or jsonl")
	rootCmd.PersistentFlags().StringSlice("select", nil, "Columns to display")
	rootCmd.PersistentFlags().Int("limit", 0, "Query limit")
	rootCmd.SetVersionTemplate("steampipe_export_aws v{{ .Version }}\n")
	/*
		rootCmd.SetVersionTemplate("steampipe_export_{{.Plugin}} v{{"{{"}} .Version {{"}}"}}\n")
	*/

	viper.BindPFlags(rootCmd.PersistentFlags())

	return rootCmd
}
