package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"strings"
)

func main() {
	if len(os.Args) == 1 {
		fmt.Printf("Usage: %s [key]", os.Args[0])
		os.Exit(1)
	}

	var key = os.Args[1]

	r := bufio.NewReader(os.Stdin)
	line, _, err := r.ReadLine()

	if err != nil {
		fmt.Printf("Error reading stdin: %v\n", err)
		os.Exit(1)
	}

	var data = make(map[string]interface{})

	if err := json.Unmarshal(line, &data); err != nil {
		fmt.Printf("Error unmarshaling line: %s (%s)\n", string(line), err)
		os.Exit(1)
	}

	if strings.Count(key, ".") > 0 {
		// fetch nested values
		printNested(data, strings.Split(key, "."))

		os.Exit(0)
	}

	if v, ok := data[key]; ok {
		printValue(v)
	}
}

func fetchNested(data map[string]interface{}, keyPath []string) interface{} {
	var n = len(keyPath)
	lastKey := keyPath[n-1]

	for _, key := range keyPath[0 : n-1] {
		if jsonKey(data, key) {
			data = fetchJSON(data, key)
		} else {
			return nil
		}
	}

	return data[lastKey]
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
	printValue(fetchNested(data, keyPath))
}

func printValue(val interface{}) {
	if val == nil {
		os.Exit(0)
	}

	if _v, ok := val.(map[string]interface{}); ok {
		printMap(_v)
		os.Exit(0)
	}

	if v, ok := val.(float64); ok {
		detectInt := fmt.Sprintf("%.10f", v)
		if strings.Count(detectInt, "0") == 10 {
			fmt.Printf("%d\n", int(v))
			os.Exit(0)
		}

		fmt.Printf("%f\n", v)
		os.Exit(0)
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
