package table

import (
	"fmt"
	"slices"
	"sort"

	"github.com/spf13/viper"
	"github.com/turbot/steampipe-plugin-sdk/v5/grpc"
	"github.com/turbot/steampipe-plugin-sdk/v5/grpc/proto"
)

func GetSchema(pluginServer *grpc.PluginServer, pluginAlias string, table string) (*proto.TableSchema, error) {
	req := &proto.GetSchemaRequest{
		Connection: pluginAlias,
	}
	pluginSchema, err := pluginServer.GetSchema(req)
	if err != nil {
		return nil, err
	}
	return pluginSchema.Schema.Schema[table], nil
}

func GetColumns(schema *proto.TableSchema) ([]string, error) {
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
