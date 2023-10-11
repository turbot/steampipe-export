package main

import (
	"encoding/csv"
	"fmt"
	"os"
	"strings"

	"github.com/turbot/steampipe-plugin-aws/aws"
	"github.com/turbot/steampipe-plugin-sdk/v5/grpc"
	"github.com/turbot/steampipe-plugin-sdk/v5/grpc/proto"
	"github.com/turbot/steampipe-plugin-sdk/v5/plugin"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var rowCount int
var pluginServer *grpc.PluginServer

type displayRowFunc func(row *proto.ExecuteResponse)

func main() {
	rootCmd := &cobra.Command{
		Use:   "awsdump",
		Short: "AWS Dump",
		Run:   executeCommand,
		Args:  cobra.ExactArgs(1),
	}

	// Define flags for input and output
	rootCmd.PersistentFlags().String("input", "", "Table name")
	rootCmd.PersistentFlags().String("config", "", "Config file data")
	rootCmd.PersistentFlags().String("where", "", "where clause data")
	rootCmd.PersistentFlags().String("column", "", "Column data")
	rootCmd.PersistentFlags().String("limit", "", "Limit data")
	rootCmd.PersistentFlags().String("output", "csv", "Output CSV file")

	plugin := plugin.NewPluginServer(&plugin.ServeOpts{
		PluginFunc: aws.Plugin,
	})

	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

}

func executeCommand(cmd *cobra.Command, args []string) {
	// TODO template
	connection := "aws"
	table := args[0]
	executeQuery(table, connection, displayCSVRow)
}

func executeQuery(tableName string, conectionName string, displayRow displayRowFunc) {
	// construct execute request
	var columns []string
	var quals map[string]*proto.Quals
	var limit int64

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
}

func displayCSVRow(displayRow *proto.ExecuteResponse) {
	
}

func createQueryContext() {

}

func generateCSV(cmd *cobra.Command, args []string) {
	if viper.GetString("output") == "" {
		fmt.Println("Output flags are required")
		os.Exit(1)
	}

	// Split the input into separate fields using a comma as a separator
	inputData := strings.Split(viper.GetString("input"), ",")

	// fmt.Println("Input data:", input)
	if len(inputData) == 0 {
		fmt.Println("No input data provided")
		os.Exit(1)
	}

	// Add code for unnamed arguments

	// Open the output file for appending
	file, err := os.OpenFile(viper.GetString("output"), os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		fmt.Println("Error opening the output file:", err)
		os.Exit(1)
	}
	defer file.Close()

	// Create a CSV writer
	writer := csv.NewWriter(file)
	defer writer.Flush()

	// Add function for adding the headers
	// addHeaders(viper.GetString("column"), file)
	// columns := args

	// Write the input data to the output CSV file
	if err := writer.Write(inputData); err != nil {
		fmt.Println("Error writing to output CSV file:", err)
		os.Exit(1)
	}

	fmt.Println("Input data successfully added to the CSV file.")
}
