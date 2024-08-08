package main

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewResp(t *testing.T) {
	rd := strings.NewReader("$5\r\nAhmed\r\n")
	resp := NewResp(rd)

	if resp.reader == nil {
		t.Fail()
	}

	assert.NotNil(t, resp.reader)
}

func TestReadLine(t *testing.T) {
	testCases := []struct {
		name      string
		input     string
		expected  string
		length    int
		expectErr bool
	}{
		{
			name:      "should read line",
			input:     "$5\r\n",
			expected:  "$5",
			length:    4,
			expectErr: false,
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			rd := strings.NewReader(tc.input)
			resp := NewResp(rd)
			line, n, err := resp.readLine()
			if tc.expectErr {
				assert.Error(t, err)
			} else {
				assert.Equal(t, tc.expected, string(line))
				assert.Equal(t, tc.length, n)
			}
		})
	}
}

func TestReadInteger(t *testing.T) {
	testCases := []struct {
		name      string
		input     string
		value     int
		expected  int
		length    int
		expectErr bool
	}{
		{
			name:      "should read integer",
			input:     "$5\r\n",
			value:     5,
			expected:  5,
			length:    4,
			expectErr: false,
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			rd := strings.NewReader(tc.input)
			resp := NewResp(rd)
			_, _, err := resp.readInteger()
			if tc.expectErr {
				assert.Error(t, err)
			} else {
				assert.Equal(t, tc.expected, tc.value)
			}
		})
	}
}

func TestRead(t *testing.T) {
	testCases := []struct {
		name      string
		input     string
		expected  Value
		expectErr bool
	}{
		{
			name:      "should parse integer",
			input:     ":5\r\n",
			expected:  Value{typ: "int", num: 5},
			expectErr: false,
		},
		{
			name:      "should parse string",
			input:     "+foo\r\n",
			expected:  Value{typ: "string", str: "foo"},
			expectErr: false,
		},
		{
			name:  "should parse array",
			input: "*4\r\n+foo\r\n+bar\r\n+baz\r\n:5\r\n",
			expected: Value{typ: "array", array: []Value{
				{typ: "string", str: "foo"},
				{typ: "string", str: "bar"},
				{typ: "string", str: "baz"},
				{typ: "int", num: 5},
			}},
		},
		{
			name:      "should parse bulk",
			input:     "$6\r\nfoobar\r\n",
			expected:  Value{typ: "bulk", bulk: "foobar"},
			expectErr: false,
		},
		{
			name:      "should parse error",
			input:     "-ERR unknown command 'foobar'\r\n",
			expected:  Value{typ: "error", str: "ERR unknown command 'foobar'"},
			expectErr: false,
		},
		{
			name:      "should parse null",
			input:     "$-1\r\n",
			expected:  Value{typ: "null"},
			expectErr: false,
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			rd := strings.NewReader(tc.input)
			r := NewResp(rd)
			v, err := r.Read()

			if tc.expectErr {
				assert.Error(t, err)
			} else {
				assert.Equal(t, tc.expected, v)
			}
		})
	}
}

func Test_marshalString(t *testing.T) {
	testCases := []struct {
		name     string
		expected []byte
		value    Value
	}{
		{
			name:     "empty string",
			value:    Value{str: ""},
			expected: []byte{'+', '\r', '\n'},
		}, {
			name:     "non-empty string",
			value:    Value{str: "hello"},
			expected: []byte{'+', 'h', 'e', 'l', 'l', 'o', '\r', '\n'},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := tc.value.marshalString()
			assert.Equal(t, tc.expected, result)
		})
	}
}

func Test_marsallArray(t *testing.T) {
	{
		testCases := []struct {
			name     string
			expected []byte
			value    Value
		}{
			{
				name:     "Test empty array",
				value:    Value{typ: "array", array: []Value{}},
				expected: []byte{'*', '0', '\r', '\n'},
			},
			{
				name:     "Test array with one element",
				value:    Value{typ: "array", array: []Value{{typ: "string", str: "hello"}}},
				expected: append([]byte{'*', '1', '\r', '\n'}, []byte{'+', 'h', 'e', 'l', 'l', 'o', '\r', '\n'}...),
			},
			{
				name:  "Test array with multiple elements",
				value: Value{typ: "array", array: []Value{{typ: "string", str: "hello"}, {typ: "string", str: "world"}}},
				expected: append([]byte{'*', '2', '\r', '\n'},
					append([]byte{'+', 'h', 'e', 'l', 'l', 'o', '\r', '\n'},
						[]byte{'+', 'w', 'o', 'r', 'l', 'd', '\r', '\n'}...)...),
			},
		}

		for _, tc := range testCases {
			t.Run(tc.name, func(t *testing.T) {
				result := tc.value.Marshal()
				assert.Equal(t, tc.expected, result)
			})
		}
	}
}
