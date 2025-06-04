package main

import (
	"context"
	"fmt"
	"github.com/spf13/viper"
	"github.com/turbot/go-kit/files"
	filehelpers "github.com/turbot/go-kit/files"
	typehelpers "github.com/turbot/go-kit/types"
	pconstants "github.com/turbot/pipe-fittings/v2/constants"
	"github.com/turbot/pipe-fittings/v2/error_helpers"
	"github.com/turbot/pipe-fittings/v2/modconfig"
	pparse "github.com/turbot/pipe-fittings/v2/parse"
	"github.com/turbot/pipe-fittings/v2/schema"
	"github.com/turbot/pipe-fittings/v2/sperr"
	"github.com/turbot/steampipe-plugin-sdk/v5/grpc/proto"
	"log"
	"os"
	"path/filepath"
)

func initConfig(ctx context.Context) error {
	// resolve args
	connectionConfigStr := viper.GetString("config")
	connectionName := viper.GetString("connection")

	// if a connection config string was provided, we  will NOT load the connection config from the config file, so
	// we will not be setting rate limiter config - just set the connection config
	if connectionConfigStr != "" {
		if connectionName != "" {
			return fmt.Errorf("you cannot use both --config and --connection flags at the same time")
		}
		return setConnectionConfig(connectionConfigStr)
	}

	// if no connection name was specified just set an empty config
	if connectionName == "" {
		return setConnectionConfig("")
	}

	// set app_specific.InstallDir
	configFolder, err := resolveConfigDir()

	// so we will try to load the connection config from the config location
	// load config from the installation folder -  load all spc files from config directory
	include := filehelpers.InclusionsFromExtensions(pconstants.ConnectionConfigExtension())
	loadOptions := &loadConfigOptions{include: include}
	steampipeConfig, err := loadConfig(ctx, configFolder, loadOptions)
	if err != nil {
		return err
	}

	conn, ok := steampipeConfig.Connections[connectionName]
	if !ok {
		return fmt.Errorf("connection '%s' not found in config", connectionName)
	}
	// if we have a connection, set the rate limiter config (if any)
	// set the connection config - this may be empty
	if err := setConnectionConfig(conn.Config); err != nil {
		return fmt.Errorf("error setting connection config: %w", err)
	}

	// set rate limiters if any
	return setRateLimiter(steampipeConfig, conn)
}

func resolveConfigDir() (string, error) {
	if configDir := viper.GetString("config-dir"); configDir != "" {
		if _, err := os.Stat(configDir); os.IsNotExist(err) {
			return "", fmt.Errorf("config directory '%s' does not exist", configDir)
		}
		return configDir, nil
	}
	// return the default config directory for the current install directory

	// set the install directory
	installDir := os.Getenv("STEAMPIPE_INSTALL_DIR")
	if installDir == "" {
		var err error
		installDir, err = files.Tildefy("~/.steampipe")
		if err != nil {
			return "", fmt.Errorf("error resolving install directory: %w", err)
		}
	}

	configFolder := filepath.Join(installDir, "config")
	return configFolder, nil

}

func setRateLimiter(steampipeConfig *SteampipeConfig, connection *modconfig.SteampipeConnection) error {
	// set the rate limiter config
	plugin, ok := steampipeConfig.PluginsInstances[typehelpers.SafeString(connection.PluginInstance)]
	if !ok {
		return nil
	}
	if len(plugin.Limiters) == 0 {
		return nil
	}

	var defs []*proto.RateLimiterDefinition
	for _, l := range plugin.Limiters {
		defs = append(defs, RateLimiterAsProto(l))
	}

	req := &proto.SetRateLimitersRequest{Definitions: defs}

	_, err := pluginServer.SetRateLimiters(req)
	return err
}

// set the connection HCL config for the plugin
func setConnectionConfig(connectionConfigStr string) error {
	connectionConfig := &proto.ConnectionConfig{
		Connection:      connection,
		Plugin:          pluginAlias,
		PluginShortName: pluginAlias,
		Config:          connectionConfigStr,
		PluginInstance:  pluginAlias,
	}

	configs := []*proto.ConnectionConfig{connectionConfig}
	req := &proto.SetAllConnectionConfigsRequest{
		Configs:        configs,
		MaxCacheSizeMb: -1,
	}

	_, err := pluginServer.SetAllConnectionConfigs(req)
	if err != nil {
		return err
	}

	return nil
}

// load config from the given folder and update steampipeConfig
// NOTE: this mutates steampipe config
type loadConfigOptions struct {
	include        []string
	allowedOptions []string
}

func loadConfig(ctx context.Context, configFolder string, opts *loadConfigOptions) (*SteampipeConfig, error) {
	steampipeConfig := newSteampipeConfig()

	// get all the config files in the directory
	configPaths, err := filehelpers.ListFilesWithContext(ctx, configFolder, &filehelpers.ListOptions{
		Flags:   filehelpers.FilesFlat,
		Include: opts.include,
	})

	if err != nil {
		log.Printf("[WARN] loadConfig: failed to get config file paths: %v\n", err)
		return nil, err
	}
	if len(configPaths) == 0 {
		return steampipeConfig, nil // no config files found, return empty config
	}

	fileData, diags := pparse.LoadFileData(configPaths...)
	if diags.HasErrors() {
		log.Printf("[WARN] loadConfig: failed to load all config files: %v\n", err)
		return nil, error_helpers.HclDiagsToError("Failed to load all config files", diags)
	}

	body, diags := pparse.ParseHclFiles(fileData)
	if diags.HasErrors() {
		return nil, error_helpers.HclDiagsToError("Failed to load all config files", diags)
	}

	// do a partial decode

	content, moreDiags := body.Content(pparse.SteampipeConfigBlockSchema)
	if moreDiags.HasErrors() {
		diags = append(diags, moreDiags...)
		return nil, error_helpers.HclDiagsToError("Failed to load config", diags)
	}

	for _, block := range content.Blocks {
		switch block.Type {

		case schema.BlockTypePlugin:
			plugin, moreDiags := pparse.DecodePlugin(block)
			diags = append(diags, moreDiags...)
			if moreDiags.HasErrors() {
				continue
			}
			// only add if the plugin alias matches the export pluginAlias
			if plugin.Alias != pluginAlias {
				continue
			}

			// add plugin to steampipeConfig
			// NOTE: this errors if there is a plugin block with a duplicate label
			if err := steampipeConfig.addPlugin(plugin); err != nil {
				return nil, err
			}

		case schema.BlockTypeConnection:
			connection, moreDiags := pparse.DecodeConnection(block)
			diags = append(diags, moreDiags...)
			if moreDiags.HasErrors() {
				continue
			}
			// only add if the plugin alias matches the pluginAlias
			if connection.PluginAlias != pluginAlias {
				continue
			}

			if existingConnection, alreadyThere := steampipeConfig.Connections[connection.Name]; alreadyThere {
				err := getDuplicateConnectionError(existingConnection, connection)
				return nil, err
			}

			steampipeConfig.Connections[connection.Name] = connection
		}
	}

	if diags.HasErrors() {
		return nil, error_helpers.HclDiagsToError("Failed to load config", diags)
	}

	log.Printf("[INFO] loadConfig calling initializePlugins")

	// resolve the plugins for each connection and create default plugin config
	// for all plugins mentioned in connection config which have no explicit config
	steampipeConfig.initializePlugins()

	return steampipeConfig, nil
}

func getDuplicateConnectionError(existingConnection, newConnection *modconfig.SteampipeConnection) error {
	return sperr.New("duplicate connection name: '%s'\n\t(%s:%d)\n\t(%s:%d)",
		existingConnection.Name, existingConnection.DeclRange.Filename, existingConnection.DeclRange.Start.Line,
		newConnection.DeclRange.Filename, newConnection.DeclRange.Start.Line)
}
