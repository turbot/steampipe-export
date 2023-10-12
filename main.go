package main

import (
	"context"
	"fmt"
	"os"
	"sort"
	"strings"

	"github.com/golang/protobuf/ptypes"
	"github.com/turbot/steampipe-plugin-aws/aws"
	"github.com/turbot/steampipe-plugin-sdk/v5/grpc"
	"github.com/turbot/steampipe-plugin-sdk/v5/grpc/proto"
	"github.com/turbot/steampipe-plugin-sdk/v5/plugin"
	"golang.org/x/exp/maps"
	"golang.org/x/exp/slices"
	"google.golang.org/protobuf/reflect/protoreflect"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/turbot/steampipe/pkg/ociinstaller"
)

var pluginServer *grpc.PluginServer
var pluginAlias = "aws"
var connection = pluginAlias

type displayRowFunc func(row *proto.ExecuteResponse)

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
	rootCmd.PersistentFlags().StringSlice("column", nil, "Column data")
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
	setConnectionConfig()
	executeQuery(table, connection, displayCSVRow)
}

func setConnectionConfig() {
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

	pluginServer.SetAllConnectionConfigs(req)
}

func executeQuery(tableName string, conectionName string, displayRow displayRowFunc) {
	// construct execute request
	var columns []string
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
		displayCSVRow(response)
	}
}

var rowCount = 0

func displayCSVRow(displayRow *proto.ExecuteResponse) {
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
		if len(viper.GetStringSlice("column")) != 0 {
			if slices.Contains(viper.GetStringSlice("column"), columnName) {
				res[columnName] = fmt.Sprintf("%v", val)
			}
		} else {
			res[columnName] = fmt.Sprintf("%v", val)
		}		
	}

	columns := maps.Keys(res)
	sort.Strings(columns)

	if rowCount == 0 {
		fmt.Println(strings.Join(columns, ","))
		rowCount ++
	}
	colVals := make([]string, len(columns))
	for i, c := range columns {
		colVals[i] = res[c]
	}
	fmt.Println(strings.Join(colVals, ","))
}
