package main

import (
	"context"
	"fmt"
	"github.com/golang/protobuf/ptypes"
	"github.com/turbot/steampipe-plugin-aws/aws"
	"github.com/turbot/steampipe-plugin-sdk/v5/grpc"
	"github.com/turbot/steampipe-plugin-sdk/v5/grpc/proto"
	"github.com/turbot/steampipe-plugin-sdk/v5/plugin"
	"golang.org/x/exp/slices"
	"google.golang.org/protobuf/reflect/protoreflect"
	"os"
	"sort"
	"strings"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/turbot/steampipe/pkg/ociinstaller"
)

var pluginServer *grpc.PluginServer
var pluginAlias = "aws"
var connection = pluginAlias

type displayRowFunc func(row *proto.ExecuteResponse, columns []string)

func main() {
	rootCmd := &cobra.Command{
		Use:   "awsdump",
		Short: "AWS Dump",
		Run:   executeCommand,
		Args:  cobra.ExactArgs(1),
	}

	// Define flags for input and output
	rootCmd.PersistentFlags().String("config", "", "Config file data")
	rootCmd.PersistentFlags().String("where", "", "where clause data")
	rootCmd.PersistentFlags().StringSlice("columns", nil, "Column data")
	rootCmd.PersistentFlags().Int("limit", 0, "Limit data")
	rootCmd.PersistentFlags().String("output", "csv", "Output CSV file")

	viper.BindPFlags(rootCmd.PersistentFlags())

	pluginServer = plugin.Server(&plugin.ServeOpts{
		PluginFunc: aws.Plugin,
	})

	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

}

func executeCommand(cmd *cobra.Command, args []string) {
	// TODO template

	table := args[0]
	if err := setConnectionConfig(); err != nil {
		// TODO display error
		os.Exit((1))
	}

	schema, err := getSchema(table)
	if err != nil {
		// TODO display error
		os.Exit((1))
	}

	// if qual, err := getQual(schema); err != nil{
	// 	// TODO display error

	// 	// validate if the requested column exists

	// 	os.Exit((1))
	// }

	columns := getColumns(schema)

	// executeQuery(table, connection, qual, displayCSVRow)
	executeQuery(table, connection, columns, displayCSVRow)
}

func getColumns(schema *proto.TableSchema) []string {
	var columns = viper.GetStringSlice("columns")
	if len(columns) == 0 {
		columns = schema.GetColumnNames()
	}
	sort.Strings(columns)
	return columns
}

func getQual() {

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

func setConnectionConfig() error {
	pluginName := ociinstaller.NewSteampipeImageRef(pluginAlias).DisplayImageRef()

	connectionConfig := &proto.ConnectionConfig{
		Connection:      connection,
		Plugin:          pluginName,
		PluginShortName: pluginAlias,
		Config:          viper.GetString("config"),
		PluginInstance:  pluginName,
	}

	configs := []*proto.ConnectionConfig{connectionConfig}
	req := &proto.SetAllConnectionConfigsRequest{
		Configs: configs,
	}

	_, err := pluginServer.SetAllConnectionConfigs(req)

	if err != nil {
		return err
	}
	return nil
}

func executeQuery(tableName string, conectionName string, columns []string, displayRow displayRowFunc) {
	// construct execute request

	var quals map[string]*proto.Quals
	var limit int64 = -1

	if viper.GetInt("limit") != 0 {
		limit = int64(viper.GetInt("limit"))
	}

	queryContext := proto.NewQueryContext(columns, quals, limit)
	req := &proto.ExecuteRequest{
		Table:                 tableName,
		QueryContext:          queryContext,
		CallId:                grpc.BuildCallId(),
		Connection:            conectionName,
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
	stream := plugin.NewLocalPluginStream(ctx)
	err := pluginServer.CallExecute(req, stream)
	if err != nil {
		fmt.Println("Error in call execute")
	}
	for {

		response, err := stream.Recv()
		if err != nil {
			fmt.Printf("[ERROR] Error receiving data from the channel: %v", err)
			break
		}
		if response == nil {
			break
		}
		displayRow(response, columns)
	}
}

var rowCount = 0

func displayCSVRow(displayRow *proto.ExecuteResponse, columns []string) {
	row := displayRow.Row

	res := make(map[string]string, len(row.Columns))
	for columnName, column := range row.Columns {
		// extract column value as interface from protobuf message
		// var i error
		var val interface{}
		if bytes := column.GetJsonValue(); bytes != nil {
			val = string(bytes)
		} else if timestamp := column.GetTimestampValue(); timestamp != nil {
			// convert from protobuf timestamp to a RFC 3339 time string
			val = ptypes.TimestampString(timestamp)
		} else {
			// get the first field descriptor and value (we only expect column message to contain a single field
			column.ProtoReflect().Range(func(descriptor protoreflect.FieldDescriptor, v protoreflect.Value) bool {
				// is this value null?
				if descriptor.JSONName() == "nullValue" {
					val = nil
				} else {
					val = v.Interface()
				}
				return false
			})
		}
		if len(viper.GetStringSlice("columns")) != 0 {
			if slices.Contains(viper.GetStringSlice("columns"), columnName) {
				res[columnName] = fmt.Sprintf("%v", val)
			}
		} else {
			res[columnName] = fmt.Sprintf("%v", val)
		}
	}

	if rowCount == 0 {
		fmt.Println(strings.Join(columns, ","))
	}

	rowCount++

	colVals := make([]string, len(columns))
	for i, c := range columns {
		colVals[i] = res[c]
	}
	fmt.Println(strings.Join(colVals, ","))
}
