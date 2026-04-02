package main

import (
	"encoding/json"
	"testing"
)

func TestFetchNested(t *testing.T) {
	t.Run("Not Nested", func(t *testing.T) {
		var input = `{"foo": "bar"}`
		var data map[string]interface{}

		json.Unmarshal([]byte(input), &data)

		val := fetchNested(data, []string{"foo", "c1"})

		if val != nil {
			t.Errorf("should return nil with wrong nested key")
		}
	})

	t.Run("Nested", func(t *testing.T) {
		var input = `{"foo": {"c1": "bar"}}`
		var data map[string]interface{}

		json.Unmarshal([]byte(input), &data)

		val := fetchNested(data, []string{"foo", "c1"})

		if val.(string) != "bar" {
			t.Errorf("should fetch correctly nested key")
		}

	})

	t.Run("Nested wrong key", func(t *testing.T) {
		var input = `{"foo": {"c2": "bar"}}`
		var data map[string]interface{}

		json.Unmarshal([]byte(input), &data)

		val := fetchNested(data, []string{"foo", "c1"})

		if val != nil {
			t.Errorf("should not fetch value with wrong key")
		}
	})

}

func TestJsonKey(t *testing.T) {
	t.Run("Not Nested map[string]interface{}", func(t *testing.T) {
		var input = `{"foo": "bar"}`
		var data map[string]interface{}

		json.Unmarshal([]byte(input), &data)

		if jsonKey(data, "foo") {
			t.Errorf("key 'foo' is not map[string]interface{} key")
		}
	})

	t.Run("Nested map[string]interface{}", func(t *testing.T) {
		var input = `{"foo": {"bazz": "bar"}}`
		var data map[string]interface{}

		json.Unmarshal([]byte(input), &data)

		if !jsonKey(data, "foo") {
			t.Errorf("key 'foo' should be map[string]interface{} key")
		}
	})
}

func TestProcessJSON(t *testing.T) {
	t.Run("Valid JSON", func(t *testing.T) {
		input := `{"ram": 8192}`
		// Should not panic or return error
		err := processJSON([]byte(input), "ram")
		if err != nil {
			t.Errorf("should process valid JSON without error, got: %v", err)
		}
	})

	t.Run("Invalid JSON", func(t *testing.T) {
		input := `not json`
		err := processJSON([]byte(input), "ram")
		if err == nil {
			t.Errorf("should return error for invalid JSON")
		}
	})

	t.Run("Empty input", func(t *testing.T) {
		input := ``
		err := processJSON([]byte(input), "ram")
		if err != nil {
			t.Errorf("should handle empty input without error")
		}
	})

	t.Run("Whitespace only", func(t *testing.T) {
		input := `   
		  `
		err := processJSON([]byte(input), "ram")
		if err != nil {
			t.Errorf("should handle whitespace-only input without error")
		}
	})
}

func TestFetchAndResolve(t *testing.T) {
	t.Run("Simple key", func(t *testing.T) {
		input := `{"name": "test"}`
		var data map[string]interface{}
		json.Unmarshal([]byte(input), &data)

		val := fetchAndResolve(data, "name")
		if val != "test" {
			t.Errorf("should fetch simple key, got: %v", val)
		}
	})

	t.Run("Nested key", func(t *testing.T) {
		input := `{"server": {"flavor": {"ram": 8192}}}`
		var data map[string]interface{}
		json.Unmarshal([]byte(input), &data)

		val := fetchAndResolve(data, "server.flavor.ram")
		if val.(float64) != 8192 {
			t.Errorf("should fetch nested key, got: %v", val)
		}
	})

	t.Run("Missing key", func(t *testing.T) {
		input := `{"name": "test"}`
		var data map[string]interface{}
		json.Unmarshal([]byte(input), &data)

		val := fetchAndResolve(data, "missing")
		if val != nil {
			t.Errorf("should return nil for missing key, got: %v", val)
		}
	})
}
