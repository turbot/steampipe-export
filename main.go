package main

import (
	"encoding/csv"
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

func main() {
	rootCmd := &cobra.Command{Use: "awsdump", Short: "AWS Dump", Run: executeDump}

	// Define flags for input and output
	rootCmd.PersistentFlags().String("input", "", "Table name")
	rootCmd.PersistentFlags().String("config", "", "Config file data")
	rootCmd.PersistentFlags().String("where", "", "where clause data")
	rootCmd.PersistentFlags().String("column", "", "Column data")
	rootCmd.PersistentFlags().String("limit", "", "Limit data")
	rootCmd.PersistentFlags().String("output", "csv", "Output CSV file")

	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func executeDump(cmd *cobra.Command, args []string) {

}

func streamCSVRow(row map[string]any) {

}

func createCSVFile() {
	file, err := os.OpenFile(viper.GetString("output"), os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		fmt.Println("Error opening the output file:", err)
		os.Exit(1)
	}
	defer file.Close()

	// Create a CSV writer
	writer := csv.NewWriter(file)
	defer writer.Flush()
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

// Append the column input to the start of the CSV file
func addHeaders(column string, file *os.File) {
	// if column == "" {
	// 	fmt.Println("Column mut be defined")
	// 	os.Exit(1)
	// }
	// columnData := strings.Split(column, ",")
	// if len(columnData) == 0 {
	// 	fmt.Print
	// }
	// if err := writer.Write(inputData); err != nil {
	// 	fmt.Println("Error writing to output CSV file:", err)
	// 	os.Exit(1)
	// }

}

// Stream function
// Execute function
