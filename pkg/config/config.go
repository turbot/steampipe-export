package config

import (
	"github.com/spf13/viper"
	"github.com/turbot/steampipe-plugin-sdk/v5/grpc"
	"github.com/turbot/steampipe-plugin-sdk/v5/grpc/proto"
	"github.com/turbot/steampipe/pkg/ociinstaller"
)

func SetConnectionConfig(pluginServer *grpc.PluginServer, pluginAlias string) error {
	pluginName := ociinstaller.NewSteampipeImageRef(pluginAlias).DisplayImageRef()

	connectionConfig := &proto.ConnectionConfig{
		Connection:      pluginAlias,
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
