package main

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

type HandlerTestCase struct {
	name     string
	args     []Value
	expected Value
}

func TestPingHandler(t *testing.T) {
	testCases := []HandlerTestCase{
		{
			name:     "empty input",
			args:     []Value{},
			expected: Value{typ: "string", str: "PONG"},
		},
		{
			name:     "bulk string",
			args:     []Value{{typ: "bulk", bulk: "John Doe"}},
			expected: Value{typ: "string", str: "John Doe"},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := ping(tc.args)
			assert.Equal(t, tc.expected, result)
		})
	}
}

func TestSetHandler(t *testing.T) {
	originalSETs := SETs
	SETs = map[string]string{}
	defer func() { SETs = originalSETs }()

	testCases := []HandlerTestCase{
		{
			name:     "correct number of args",
			args:     []Value{{bulk: "foo"}, {bulk: "bar"}},
			expected: Value{typ: "string", str: "OK"},
		},
		{
			name:     "incorrect number of args",
			args:     []Value{{bulk: "foo"}},
			expected: Value{typ: "error", str: "ERR wrong number of arguments for SET command"},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := Handlers["SET"](tc.args)
			assert.Equal(t, tc.expected, result)
		})
	}
}

func TestGetHandler(t *testing.T) {
	originalSETs := SETs
	SETs = map[string]string{}

	defer func() { SETs = originalSETs }()

	t.Run("valid key", func(t *testing.T) {
		// set Value
		Handlers["SET"](
			[]Value{{bulk: "foo"}, {bulk: "bar"}})

		result := Handlers["GET"]([]Value{{bulk: "foo"}})
		expected := Value{typ: "bulk", bulk: "bar"}
		assert.Equal(t, expected, result)
	})

	t.Run("invalid key", func(t *testing.T) {
		result := Handlers["GET"]([]Value{{bulk: "invalid_key"}})
		expected := Value{typ: "null"}
		assert.Equal(t, expected, result)
	})
}

func TestHSetHandler(t *testing.T) {
	originalHSETs := HSETs
	HSETs = map[string]map[string]string{}

	defer func() { HSETs = originalHSETs }()

	testCases := []HandlerTestCase{
		{
			name:     "correct number of args",
			args:     []Value{{bulk: "myhash"}, {bulk: "field1"}, {bulk: "Hello"}},
			expected: Value{typ: "string", str: "OK"},
		},
		{
			name:     "incorrect number of args",
			args:     []Value{{bulk: "myhash"}},
			expected: Value{typ: "error", str: "ERR wrong number of arguments for HSET command"},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := Handlers["HSET"](tc.args)
			assert.Equal(t, tc.expected, result)
		})
	}
}

func TestHGetHandler(t *testing.T) {
	originalHSETs := HSETs
	HSETs = map[string]map[string]string{}
	defer func() { HSETs = originalHSETs }()

	HSETs["hash1"] = map[string]string{"field1": "value1"}

	testCases := []struct {
		name     string
		args     []Value
		expected Value
	}{
		{
			name:     "correct number of arguments, field exists",
			args:     []Value{{bulk: "hash1"}, {bulk: "field1"}},
			expected: Value{typ: "bulk", bulk: "value1"},
		},
		{
			name:     "correct number of arguments, field does not exist",
			args:     []Value{{bulk: "hash1"}, {bulk: "field2"}},
			expected: Value{typ: "null"},
		},
		{
			name:     "correct number of arguments, hash does not exist",
			args:     []Value{{bulk: "hash2"}, {bulk: "field1"}},
			expected: Value{typ: "null"},
		},
		{
			name:     "incorrect number of arguments",
			args:     []Value{{bulk: "hash1"}},
			expected: Value{typ: "error", str: "ERR wrong number of arguments for HSET command"},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := Handlers["HGET"](tc.args)
			assert.Equal(t, tc.expected, result)
		})
	}
}

func TestHGetAllHandler(t *testing.T) {
	// Reset the HSETs map before each test
	originalHSETs := HSETs
	HSETs = map[string]map[string]string{}
	defer func() { HSETs = originalHSETs }()

	// Initialize some data for testing
	HSETs["hash1"] = map[string]string{"field1": "value1", "field2": "value2"}

	// Define the test cases
	testCases := []struct {
		name     string
		args     []Value
		expected Value
	}{
		{
			name: "correct number of arguments, hash exists",
			args: []Value{{bulk: "hash1"}},
			expected: Value{
				typ: "array",
				array: []Value{
					{typ: "bulk", bulk: "field1"},
					{typ: "bulk", bulk: "value1"},
					{typ: "bulk", bulk: "field2"},
					{typ: "bulk", bulk: "value2"},
				},
			},
		},
		{
			name:     "correct number of arguments, hash does not exist",
			args:     []Value{{bulk: "hash2"}},
			expected: Value{typ: "null"},
		},
		{
			name:     "incorrect number of arguments",
			args:     []Value{},
			expected: Value{typ: "error", str: "ERR wrong number of arguments for HGETALL command"},
		},
	}

	// Execute each test case
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := Handlers["HGETALL"](tc.args)
			assert.Equal(t, tc.expected, result)
		})
	}
}
