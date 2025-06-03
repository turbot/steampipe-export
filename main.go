package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"slices"
	"sort"
	"time"

	"github.com/hashicorp/go-hclog"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/turbot/steampipe-export/constants"
	"github.com/turbot/steampipe-plugin-sdk/v5/anywhere"
	"github.com/turbot/steampipe-plugin-sdk/v5/grpc"
	"github.com/turbot/steampipe-plugin-sdk/v5/grpc/proto"
	"github.com/turbot/steampipe-plugin-sdk/v5/logging"
)

var (
	// These variables will be set by GoReleaser.
	version = constants.DefaultVersion
	commit  = constants.DefaultCommit
	date    = constants.DefaultDate
	builtBy = constants.DefaultBuiltBy
)

func main() {
	// add the auto-populated version properties into viper
	setVersionProperties()
	setupLogger(pluginAlias)
	rootCmd := setupRootCommand()

	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func setVersionProperties() {
	viper.SetDefault(constants.ConfigKeyVersion, version)
	viper.SetDefault(constants.ConfigKeyCommit, commit)
	viper.SetDefault(constants.ConfigKeyDate, date)
	viper.SetDefault(constants.ConfigKeyBuiltBy, builtBy)
}

func executeCommand(cmd *cobra.Command, args []string) {
	table := args[0]

	// set the connection and rate limiter config
	if err := setConfig(cmd.Context()); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	schema, err := getSchema(table)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	columns, err := getColumns(schema)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	quals, err := buildQuals(viper.GetStringSlice("where"), schema)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	limiterConfigStr := viper.GetString("limiter")
	if limiterConfigStr != "" {
		if err = setRateLimiters(limiterConfigStr); err != nil {
			fmt.Println(err)
			os.Exit(1)
		}

	}

	var displayFunc displayRowFunc
	var onCompleteFunc func() error
	outputFormat := viper.GetString("output")
	switch outputFormat {
	case "json":
		displayFunc = displayJSONRow
		onCompleteFunc = finishJSONOutput
	case "jsonl":
		displayFunc = displayJSONLRow
	case "csv":
		displayFunc = displayCSVRow
	default:
		fmt.Printf("Unsupported output format: %s\n", outputFormat)
		os.Exit(1)
	}

	// execute the query
	if err := executeQuery(table, connection, columns, quals, displayFunc); err != nil {
		fmt.Printf("[ERROR] Error executing query: %v", err)
		os.Exit(1)
	}

	// finalize the output if needed
	if onCompleteFunc != nil {
		if err := finishJSONOutput(); err != nil {
			fmt.Printf("[ERROR] Error finishing JSON output: %v", err)
			os.Exit(1)
		}
	}
}

func executeQuery(tableName string, connectionName string, columns []string, qual map[string]*proto.Quals, displayRow displayRowFunc) error {
	// construct execute request

	var qualMap = map[string]*proto.Quals{}

	if qual != nil {
		qualMap = qual
	}

	var limit int64 = -1

	if viper.GetInt("limit") != 0 {
		limit = int64(viper.GetInt("limit"))
	}

	queryContext := proto.NewQueryContext(columns, qualMap, limit, nil)
	req := &proto.ExecuteRequest{
		Table:                 tableName,
		QueryContext:          queryContext,
		CallId:                grpc.BuildCallId(),
		Connection:            connectionName,
		TraceContext:          nil,
		ExecuteConnectionData: make(map[string]*proto.ExecuteConnectionData),
	}
	req.ExecuteConnectionData = map[string]*proto.ExecuteConnectionData{
		req.Connection: {
			Limit:        req.QueryContext.Limit,
			CacheEnabled: false,
		},
	}
	ctx := context.Background()
	stream := anywhere.NewLocalPluginStream(ctx)
	pluginServer.CallExecuteAsync(req, stream)

	// Wait for first row or error
	select {
	case <-stream.Ready():
		// Stream is ready to receive data
	case <-ctx.Done():
		fmt.Printf("[ERROR] Context cancelled while waiting for stream: %v", ctx.Err())
		return ctx.Err()
	}

	for {
		response, err := stream.Recv()
		if err != nil {
			fmt.Printf("[ERROR] Error receiving data from the channel: %v", err)
			return err
		}
		if response == nil {
			// Stream is complete
			break
		}
		if err := displayRow(response, columns); err != nil {
			fmt.Printf("[ERROR] Error displaying row: %v", err)
			return err
		}
	}

	// Check for any errors that might have occurred during streaming
	_, err := stream.Recv()
	if err != nil {
		fmt.Printf("[ERROR] Error after stream completion: %v", err)
		return err
	}

	return nil
}

func getColumns(schema *proto.TableSchema) ([]string, error) {
	var columns = viper.GetStringSlice("select")
	if len(columns) != 0 {
		tableColumn := schema.GetColumnNames()
		for _, item := range columns {
			if !slices.Contains(tableColumn, item) {
				return nil, fmt.Errorf("column %s does not exist", item)
			}
		}
	}
	if len(columns) == 0 {
		columns = schema.GetColumnNames()
	}
	sort.Strings(columns)
	return columns, nil
}

func getSchema(table string) (*proto.TableSchema, error) {
	req := &proto.GetSchemaRequest{
		Connection: connection,
	}
	pluginSchema, err := pluginServer.GetSchema(req)
	if err != nil {
		return nil, err
	}
	return pluginSchema.Schema.Schema[table], nil
}

func setupLogger(plugin string) {
	level := logging.LogLevel()
	hcLevel := hclog.LevelFromString(level)

	options := &hclog.LoggerOptions{
		// make the name unique so that logs from this instance can be filtered
		Name:       fmt.Sprintf("[%s]", plugin),
		Level:      hcLevel,
		Output:     os.Stderr,
		TimeFn:     func() time.Time { return time.Now().UTC() },
		TimeFormat: "2006-01-02 15:04:05.000 UTC",
	}
	logger := logging.NewLogger(options)
	log.SetOutput(logger.StandardWriter(&hclog.StandardLoggerOptions{InferLevels: true}))
	log.SetPrefix("")
	log.SetFlags(0)
}
