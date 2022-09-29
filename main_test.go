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
