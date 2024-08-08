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
