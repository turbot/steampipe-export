package main

import (
	"encoding/json"
	"fmt"
	"os"

	encodingcsv "encoding/csv"
	"github.com/golang/protobuf/ptypes"
	"github.com/spf13/viper"
	"github.com/turbot/steampipe-plugin-sdk/v6/grpc/proto"
	"google.golang.org/protobuf/reflect/protoreflect"
)
type displayRowFunc func(row *proto.ExecuteResponse, columns []string) error

// Global variables to manage the state of JSON output
var isFirstJSONRow = true
var isJSONStarted = false
var rowCount = 0

// displayCSVRow formats and outputs the row data in CSV format, managing headers and selected columns.
func displayCSVRow(displayRow *proto.ExecuteResponse, columns []string) error {
	row := displayRow.Row
	selectColumns := viper.GetStringSlice("select")

	// Process each column and store values in a map
	res := make(map[string]string, len(row.Columns))
	for columnName, column := range row.Columns {
		var val interface{}
		if bytes := column.GetJsonValue(); bytes != nil {
			val = string(bytes)
		} else if timestamp := column.GetTimestampValue(); timestamp != nil {
			val = ptypes.TimestampString(timestamp)
		} else {
			column.ProtoReflect().Range(func(descriptor protoreflect.FieldDescriptor, v protoreflect.Value) bool {
				if descriptor.JSONName() == "nullValue" {
					val = nil
				} else {
					val = v.Interface()
				}
				return false
			})
		}
		res[columnName] = fmt.Sprintf("%v", val)
	}

	// Prepare CSV writer
	writer := encodingcsv.NewWriter(os.Stdout)
	defer writer.Flush()

	// Write headers
	if rowCount == 0 {
		if len(selectColumns) > 0 {
			// Write headers based on selectColumns
			if err := writer.Write(selectColumns); err != nil {
				return fmt.Errorf("error writing headers: %v", err)
			}
		} else {
			// Write all headers
			if err := writer.Write(columns); err != nil {
				return fmt.Errorf("error writing headers: %v", err)
			}
		}
		writer.Flush()

		if err := writer.Error(); err != nil {
			return fmt.Errorf("error flushing headers: %v", err)
		}
	}

	rowCount++

	// Generate row data
	var colVals []string
	if len(selectColumns) > 0 {
		colVals = make([]string, len(selectColumns))
		for i, columnName := range selectColumns {
			colVals[i], _ = res[columnName] // Using _ to ignore whether columnName is present in res
		}
	} else {
		colVals = make([]string, len(columns))
		for i, columnName := range columns {
			colVals[i], _ = res[columnName]
		}
	}

	// Write the row data
	if err := writer.Write(colVals); err != nil {
		return fmt.Errorf("error writing row data: %v", err)
	}
	writer.Flush()

	// Handle potential errors from the writer
	if err := writer.Error(); err != nil {
		return fmt.Errorf("error flushing row data: %v", err)
	}

	return nil
}

// displayJSONRow formats and outputs the row data in JSON format, managing array formatting.
func displayJSONRow(displayRow *proto.ExecuteResponse, columns []string) error {
	row := displayRow.Row
	selectColumns := viper.GetStringSlice("select")

	// Process each column and store values in a map
	res := make(map[string]interface{}, len(row.Columns))
	for columnName, column := range row.Columns {
		var val interface{}
		if bytes := column.GetJsonValue(); bytes != nil {
			val = string(bytes)
		} else if timestamp := column.GetTimestampValue(); timestamp != nil {
			val = ptypes.TimestampString(timestamp)
		} else {
			column.ProtoReflect().Range(func(descriptor protoreflect.FieldDescriptor, v protoreflect.Value) bool {
				if descriptor.JSONName() == "nullValue" {
					val = nil
				} else {
					val = v.Interface()
				}
				return false
			})
		}
		res[columnName] = val
	}

	// Create a map for the selected columns
	selectedRes := make(map[string]interface{})
	if len(selectColumns) > 0 {
		for _, columnName := range selectColumns {
			if val, ok := res[columnName]; ok {
				selectedRes[columnName] = val
			}
		}
	} else {
		selectedRes = res
	}

	// Convert to JSON
	jsonData, err := json.Marshal(selectedRes)
	if err != nil {
		return fmt.Errorf("error marshaling JSON: %v", err)
	}

	// Print the JSON
	if isFirstJSONRow {
		fmt.Print("[")
		isFirstJSONRow = false
		isJSONStarted = true
	} else {
		fmt.Print(",")
	}
	fmt.Print(string(jsonData))

	return nil
}

// finishJSONOutput checks if JSON output has started and closes the JSON array if it has.
// It should be called if the output format is JSON and all rows have been processed.
func finishJSONOutput() error {
	if isJSONStarted {
		fmt.Println("]")
	}
	return nil
}

// displayJSONLRow formats and outputs the row data in JSON Lines (JSONL) format for selected columns.
func displayJSONLRow(displayRow *proto.ExecuteResponse, columns []string) error {
	row := displayRow.Row
	selectColumns := viper.GetStringSlice("select")

	// Process each column and store values in a map
	res := make(map[string]interface{}, len(row.Columns))
	for columnName, column := range row.Columns {
		var val interface{}
		if bytes := column.GetJsonValue(); bytes != nil {
			val = string(bytes)
		} else if timestamp := column.GetTimestampValue(); timestamp != nil {
			val = ptypes.TimestampString(timestamp)
		} else {
			column.ProtoReflect().Range(func(descriptor protoreflect.FieldDescriptor, v protoreflect.Value) bool {
				if descriptor.JSONName() == "nullValue" {
					val = nil
				} else {
					val = v.Interface()
				}
				return false
			})
		}
		res[columnName] = val
	}

	// Create a map for the selected columns
	selectedRes := make(map[string]interface{})
	if len(selectColumns) > 0 {
		for _, columnName := range selectColumns {
			if val, ok := res[columnName]; ok {
				selectedRes[columnName] = val
			}
		}
	} else {
		selectedRes = res
	}

	// Convert to JSON
	jsonData, err := json.Marshal(selectedRes)
	if err != nil {
		return fmt.Errorf("error marshaling JSON: %v", err)
	}

	// Print the JSON line
	fmt.Println(string(jsonData))

	return nil
}
