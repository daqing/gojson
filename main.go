package main

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strings"
)

const version = "0.2.0"

const helpText = `gojson - A simple JSON extraction tool

VERSION:
   %s

USAGE:
   gojson <key> [json...]
   gojson <key> < input.json
   command | gojson <key>

COMMANDS:
   help, -h, --help     Show this help message
   version, -v, --version   Show version information

KEY FORMAT:
   Simple key            Extract a top-level field (e.g., "name")
   Nested key            Extract nested fields using dot notation (e.g., "server.flavor.ram")
   Array iteration       Use ".each" to iterate arrays (e.g., "servers.each.name")

INPUT FORMATS:
   1. Single JSON object from stdin
   2. Multiple JSON objects (NDJSON/JSON Lines) from stdin
   3. JSON objects as command-line arguments

EXAMPLES:

   # 1. Extract a simple field
   $ echo '{"name": "Alice", "age": 30}' | gojson name
   Alice

   # 2. Extract a nested field
   $ echo '{"server": {"flavor": {"ram": 8192}}}' | gojson server.flavor.ram
   8192

   # 3. Iterate over an array
   $ echo '{"servers": [{"name": "s1"}, {"name": "s2"}]}' | gojson servers.each.name
   s1
   s2

   # 4. Process multiple JSON objects (NDJSON)
   $ (echo '{"ram": 8192}'; echo '{"ram": 4096}') | gojson ram
   8192
   4096

   # 5. Pass JSON as arguments (useful with xargs - careful with quotes)
   $ gojson ram '{"ram": 8192}' '{"ram": 4096}'
   8192
   4096

   # 6. Chain gojson commands (process output from another gojson)
   $ cat servers.json | gojson servers.each.config | gojson memory
   8192
   4096

OUTPUT FORMAT:
   - Strings: printed as-is
   - Integers: printed without decimals (e.g., 8192)
   - Floats: printed with 6 decimal places (e.g., 3.140000)
   - Objects: printed as compact JSON
   - Arrays (from .each): each item printed on a new line

NOTES:
   - When using xargs with JSON, quotes may be stripped by the shell.
     Use a while-read loop instead:
     
     some-command | while IFS= read -r line; do gojson key "$line"; done
`

func printHelp() {
	fmt.Printf(helpText, version)
}

func main() {
	if len(os.Args) == 1 {
		fmt.Printf("Usage: %s [key] [json...]\n", os.Args[0])
		fmt.Printf("Run '%s help' for more information.\n", os.Args[0])
		os.Exit(1)
	}

	var key = os.Args[1]

	// Handle help command
	if key == "help" || key == "-h" || key == "--help" {
		printHelp()
		os.Exit(0)
	}

	// Handle version command (before reading stdin)
	if key == "version" || key == "-v" || key == "--version" {
		fmt.Printf("gojson version %s\n", version)
		os.Exit(0)
	}

	// Check if there are JSON arguments passed directly (e.g., from xargs)
	if len(os.Args) > 2 {
		// Process each JSON argument
		hasError := false
		for _, arg := range os.Args[2:] {
			if err := processJSON([]byte(arg), key); err != nil {
				fmt.Fprintf(os.Stderr, "Error processing argument: %v\n", err)
				hasError = true
			}
		}
		if hasError {
			os.Exit(1)
		}
		os.Exit(0)
	}

	// Read from stdin - support both single JSON and newline-delimited JSON (NDJSON)
	input, err := io.ReadAll(os.Stdin)
	if err != nil {
		fmt.Printf("Error reading stdin: %v\n", err)
		os.Exit(1)
	}

	input = bytes.TrimSpace(input)
	if len(input) == 0 {
		os.Exit(0)
	}

	// Try to parse as a single JSON object first
	var data map[string]interface{}
	if err := json.Unmarshal(input, &data); err == nil {
		// Single JSON object
		printValue(fetchAndResolve(data, key))
		os.Exit(0)
	}

	// Try to parse as newline-delimited JSON (NDJSON)
	scanner := bufio.NewScanner(bytes.NewReader(input))
	hasError := false
	lineNum := 0
	for scanner.Scan() {
		lineNum++
		line := bytes.TrimSpace(scanner.Bytes())
		if len(line) == 0 {
			continue
		}

		var obj map[string]interface{}
		if err := json.Unmarshal(line, &obj); err != nil {
			fmt.Fprintf(os.Stderr, "Error parsing line %d: %v\n", lineNum, err)
			hasError = true
			continue
		}

		printValue(fetchAndResolve(obj, key))
	}

	if err := scanner.Err(); err != nil {
		fmt.Fprintf(os.Stderr, "Error reading input: %v\n", err)
		os.Exit(1)
	}

	if hasError {
		os.Exit(1)
	}
}

func processJSON(input []byte, key string) error {
	input = bytes.TrimSpace(input)
	if len(input) == 0 {
		return nil
	}

	var data map[string]interface{}
	if err := json.Unmarshal(input, &data); err != nil {
		return err
	}

	printValue(fetchAndResolve(data, key))
	return nil
}

func fetchAndResolve(data map[string]interface{}, key string) interface{} {
	if strings.Count(key, ".") > 0 {
		return fetchNested(data, strings.Split(key, "."))
	}
	if v, ok := data[key]; ok {
		return v
	}
	return nil
}

func fetchNested(data map[string]interface{}, keyPath []string) interface{} {
	var n = len(keyPath)
	lastKey := keyPath[n-1]

	for i, key := range keyPath[0 : n-1] {
		if key == "each" {
			// Handle array iteration - find the array in current data level
			return fetchArrayEach(data, keyPath[i+1:])
		}
		if val, ok := data[key]; ok {
			// Check if next value is an array (for .each processing)
			if arr, ok := val.([]interface{}); ok {
				// The next key should be "each"
				if i+1 < n-1 && keyPath[i+1] == "each" {
					return processArrayEach(arr, keyPath[i+2:])
				}
				return nil
			}
			if jsonKey(data, key) {
				data = fetchJSON(data, key)
			} else {
				return nil
			}
		} else {
			return nil
		}
	}

	return data[lastKey]
}

func fetchArrayEach(data map[string]interface{}, keyPath []string) interface{} {
	if len(keyPath) == 0 {
		return nil
	}

	// Find array in current data level
	var arr []interface{}
	for _, v := range data {
		if a, ok := v.([]interface{}); ok {
			arr = a
			break
		}
	}

	if arr == nil {
		return nil
	}

	return processArrayEach(arr, keyPath)
}

func processArrayEach(arr []interface{}, keyPath []string) []interface{} {
	result := make([]interface{}, 0, len(arr))
	for _, item := range arr {
		if m, ok := item.(map[string]interface{}); ok {
			val := fetchNestedFromMap(m, keyPath)
			if val != nil {
				result = append(result, val)
			}
		}
	}
	return result
}

func fetchNestedFromMap(data map[string]interface{}, keyPath []string) interface{} {
	if len(keyPath) == 0 {
		return data
	}

	current := keyPath[0]
	if val, ok := data[current]; ok {
		if len(keyPath) == 1 {
			return val
		}
		if nextMap, ok := val.(map[string]interface{}); ok {
			return fetchNestedFromMap(nextMap, keyPath[1:])
		}
	}
	return nil
}

func fetchJSON(data map[string]interface{}, key string) map[string]interface{} {
	if val, ok := data[key]; ok {
		return val.(map[string]interface{})
	}

	return nil
}

// test if the key of data is nested map[string]interface{} data
func jsonKey(data map[string]interface{}, key string) bool {
	val := data[key]

	if val == nil {
		return false
	}

	if _, ok := val.(map[string]interface{}); ok {
		return true
	}

	return false
}

func printNested(data map[string]interface{}, keyPath []string) {
	// Check if any key is "each"
	hasEach := false
	for _, key := range keyPath {
		if key == "each" {
			hasEach = true
			break
		}
	}

	if hasEach {
		result := fetchNested(data, keyPath)
		printValue(result)
	} else {
		printValue(fetchNested(data, keyPath))
	}
}

func printValue(val interface{}) {
	if val == nil {
		return
	}

	// Handle array result from .each
	if arr, ok := val.([]interface{}); ok {
		for _, item := range arr {
			printSingleValue(item)
		}
		return
	}

	printSingleValue(val)
}

func printSingleValue(val interface{}) {
	if val == nil {
		return
	}

	if _v, ok := val.(map[string]interface{}); ok {
		printMap(_v)
		return
	}

	if v, ok := val.(float64); ok {
		detectInt := fmt.Sprintf("%.10f", v)
		if strings.Count(detectInt, "0") == 10 {
			fmt.Printf("%d\n", int(v))
			return
		}

		fmt.Printf("%f\n", v)
		return
	}

	fmt.Printf("%v\n", val)
}

func printMap(m map[string]interface{}) {
	if v, err := json.Marshal(m); err != nil {
		fmt.Printf("json marshal error: %v", err)
		os.Exit(1)
	} else {
		fmt.Printf("%s\n", v)
	}
}
