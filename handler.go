package main

import (
	"sync"
)

var (
	SETs  = map[string]string{}
	HSETs = map[string]map[string]string{}
)

// ensure SETsMu is not modified by multiple threads at the same time
var (
	SETsMu  = sync.RWMutex{}
	HSETsMu = sync.RWMutex{}
)

var Handlers = map[string]func([]Value) Value{
	"PING":    ping,
	"SET":     set,
	"GET":     get,
	"HSET":    hset,
	"HGET":    hget,
	"HGETALL": hgetAll,
}

func ping(args []Value) Value {
	if len(args) == 0 {
		return Value{typ: "string", str: "PONG"}
	}

	return Value{typ: "string", str: args[0].bulk}
}

func set(args []Value) Value {
	if len(args) != 2 {
		return Value{typ: "error", str: "ERR wrong number of arguments for SET command"}
	}

	key := args[0].bulk
	value := args[1].bulk

	SETsMu.Lock()
	SETs[key] = value
	SETsMu.Unlock()

	return Value{typ: "string", str: "OK"}
}

func get(args []Value) Value {
	if len(args) != 1 {
		return Value{typ: "error", str: "ERR wrong number of arguments for GET command"}
	}

	key := args[0].bulk

	SETsMu.RLock()
	value, ok := SETs[key]
	if !ok {
		return Value{typ: "null"}
	}

	return Value{typ: "bulk", bulk: value}
}

func hset(args []Value) Value {
	if len(args) != 3 {
		return Value{typ: "error", str: "ERR wrong number of arguments for HSET command"}
	}

	hash := args[0].bulk
	key := args[1].bulk
	value := args[2].bulk

	HSETsMu.Lock()
	if _, ok := HSETs[hash]; !ok {
		HSETs[hash] = map[string]string{}
	}
	HSETs[hash][key] = value
	HSETsMu.Unlock()

	return Value{typ: "string", str: "OK"}
}

func hget(args []Value) Value {
	if len(args) != 2 {
		return Value{typ: "error", str: "ERR wrong number of arguments for HSET command"}
	}

	hash := args[0].bulk
	key := args[1].bulk

	HSETsMu.RLock()
	value, ok := HSETs[hash][key]
	HSETsMu.RUnlock()

	if !ok {
		return Value{typ: "null"}
	}

	return Value{typ: "bulk", bulk: value}
}

func hgetAll(args []Value) Value {
	if len(args) != 1 {
		return Value{typ: "error", str: "ERR wrong number of arguments for HGETALL command"}
	}

	hash := args[0].bulk

	HSETsMu.RLock()
	value, ok := HSETs[hash]
	HSETsMu.RUnlock()

	if !ok {
		return Value{typ: "null"}
	}

	values := []Value{}
	for k, v := range value {
		values = append(values, Value{typ: "bulk", bulk: k})
		values = append(values, Value{typ: "bulk", bulk: v})
	}

	return Value{typ: "array", array: values}
}
