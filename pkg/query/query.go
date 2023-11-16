package query

import (
	"context"
	"fmt"

	"github.com/spf13/viper"
	"github.com/turbot/steampipe-plugin-sdk/v5/grpc"
	"github.com/turbot/steampipe-plugin-sdk/v5/grpc/proto"
	"github.com/turbot/steampipe-plugin-sdk/v5/plugin"
	"github.com/turbot/steampipe-table-dump/pkg/plugin_server"
)

func ExecuteQuery(tableName string, conectionName string, columns []string, qual map[string]*proto.Quals, displayRow DisplayRowFunc) {
	pluginServer := plugin_server.GetPluginServer()

	// construct execute request
	var qualMap = map[string]*proto.Quals{}

	if qual != nil {
		qualMap = qual
	}

	var limit int64 = -1

	if viper.GetInt("limit") != 0 {
		limit = int64(viper.GetInt("limit"))
	}

	queryContext := proto.NewQueryContext(columns, qualMap, limit)
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
