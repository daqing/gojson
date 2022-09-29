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
		fmt.Printf("Error unmarshaling line: %v (%s)\n", string(line), err)
	}

	if v, ok := data[key]; ok {
		printValue(v)
	}
}

func printValue(val interface{}) {
	if v, ok := val.(float64); ok {
		detectInt := fmt.Sprintf("%.10f", v)
		if strings.Count(detectInt, "0") == 10 {
			fmt.Printf("%d\n", int(v))
			os.Exit(0)
		}

		fmt.Printf("%f\n", v)
		os.Exit(0)
	}

	fmt.Printf("%s\n", val)
}
