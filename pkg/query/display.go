package query

import (
	"encoding/csv"
	"fmt"
	"os"
	"slices"
	"strings"

	"github.com/golang/protobuf/ptypes"
	"github.com/spf13/viper"
	"github.com/turbot/steampipe-plugin-sdk/v5/grpc/proto"
	"google.golang.org/protobuf/reflect/protoreflect"
)

type DisplayRowFunc func(row *proto.ExecuteResponse, columns []string)

func DisplayCSVRow(displayRow *proto.ExecuteResponse, columns []string) {
	var rowCount = 0
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
		if len(viper.GetStringSlice("select")) != 0 {
			if slices.Contains(viper.GetStringSlice("select"), columnName) {
				res[columnName] = fmt.Sprintf("%v", val)
			}
		} else {
			res[columnName] = fmt.Sprintf("%v", val)
		}
	}

	var dataHeader string
	var dataRows string
	writer := csv.NewWriter(os.Stdout)

	defer writer.Flush()
	if rowCount == 0 {
		dataHeader = strings.Join(columns, ",")
		fields := strings.Split(dataHeader, ",")
		writer.Write(fields)
		writer.Flush()

		if err := writer.Error(); err != nil {
			fmt.Println(err)
		}
	}

	rowCount++

	colVals := make([]string, len(columns))
	for i, c := range columns {
		colVals[i] = res[c]
	}
	dataRows = strings.Join(colVals, ",")
	fields := strings.Split(dataRows, ",")
	writer.Write(fields)
	writer.Flush()

	if err := writer.Error(); err != nil {
		fmt.Println(err)
	}
}
