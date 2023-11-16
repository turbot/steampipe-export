package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/turbot/steampipe-plugin-sdk/v5/grpc/proto"
	"github.com/turbot/steampipe-table-dump/pkg/config"
	"github.com/turbot/steampipe-table-dump/pkg/plugin_server"
	"github.com/turbot/steampipe-table-dump/pkg/query"
	"github.com/turbot/steampipe-table-dump/pkg/table"
	"github.com/turbot/steampipe-table-dump/pkg/version"
	"github.com/turbot/steampipe-table-dump/utils"
)

var rootCmd = &cobra.Command{
	Use:   "spdump",
	Short: "Steampipe Data Dump",
	Run:   executeCommand,
	Args:  cobra.ExactArgs(1),
}

func InitCmd() {
	rootCmd.SetVersionTemplate(fmt.Sprintf("spdump v%s\n", version.SpDumpVersion.String()))

	// Define flags
	rootCmd.PersistentFlags().String("config", "", "Config file data")
	rootCmd.PersistentFlags().String("where", "", "where clause data")
	rootCmd.PersistentFlags().StringSlice("select", nil, "Column data to display")
	rootCmd.PersistentFlags().Int("limit", 0, "Limit data")

	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func executeCommand(cmd *cobra.Command, args []string) {
	// get the plugin server
	pluginServer := plugin_server.GetPluginServer()
	pluginAlias := plugin_server.GetPluginAlias()

	tableName := args[0]
	if err := config.SetConnectionConfig(pluginServer, pluginAlias); err != nil {
		// TODO: Handle the error
		fmt.Println(err)
		return
	}

	schema, err := table.GetSchema(pluginServer, pluginAlias, tableName)
	if err != nil {
		// TODO: Handle the error
		fmt.Println(err)
		return
	}

	columns, err := table.GetColumns(schema)
	if err != nil {
		// TODO: Handle the error
		fmt.Println(err)
		return
	}

	var qual map[string]*proto.Quals
	if viper.GetString("where") != "" {
		whereFlag := viper.GetString("where")
		qual, err = utils.FilterStringToQuals(whereFlag, schema)
		if err != nil {
			// TODO: Handle the error
			fmt.Println(err)
			return
		}
	}

	query.ExecuteQuery(tableName, pluginAlias, columns, qual, query.DisplayCSVRow)
}
