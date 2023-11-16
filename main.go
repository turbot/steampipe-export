package main

import (
	"github.com/turbot/steampipe-plugin-sdk/v5/grpc"
	"github.com/turbot/steampipe-plugin-sdk/v5/grpc/proto"
	"github.com/turbot/steampipe-table-dump/cmd"
)

var pluginServer *grpc.PluginServer

type displayRowFunc func(row *proto.ExecuteResponse, columns []string)

func main() {
	cmd.InitCmd()
}
