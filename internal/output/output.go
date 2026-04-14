package output

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/spf13/viper"
)

// Print outputs data as JSON or table based on the --json flag.
func Print(data interface{}, tableFn func(data interface{})) {
	if viper.GetBool("json") {
		JSON(data)
		return
	}
	tableFn(data)
}

// JSON prints data as indented JSON to stdout.
func JSON(data interface{}) {
	enc := json.NewEncoder(os.Stdout)
	enc.SetIndent("", "  ")
	if err := enc.Encode(data); err != nil {
		fmt.Fprintf(os.Stderr, "Error encoding JSON: %v\n", err)
	}
}

// PrintRaw prints the raw byte body as JSON (for untyped API responses).
func PrintRaw(body []byte, tableFn func(data map[string]interface{})) {
	var data map[string]interface{}
	if err := json.Unmarshal(body, &data); err != nil {
		// Try as array
		var arr []interface{}
		if err2 := json.Unmarshal(body, &arr); err2 != nil {
			fmt.Fprintln(os.Stderr, "Error parsing response")
			return
		}
		if viper.GetBool("json") {
			JSON(arr)
		} else {
			JSON(arr)
		}
		return
	}
	if viper.GetBool("json") {
		JSON(data)
	} else {
		tableFn(data)
	}
}

// Success prints a success message.
func Success(msg string) {
	fmt.Println(msg)
}

// Successf prints a formatted success message.
func Successf(format string, args ...interface{}) {
	fmt.Printf(format+"\n", args...)
}
