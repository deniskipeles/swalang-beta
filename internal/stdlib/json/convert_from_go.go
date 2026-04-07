package pyjson

import (
	"fmt"
	"github.com/deniskipeles/pylearn/internal/object"
)

// convertInterfaceToPylearn converts a Go interface{} (typically from json.Unmarshal)
// to its Pylearn object equivalent.
// Handles nested structures recursively.
func ConvertInterfaceToPylearn(data interface{}, depth int) (object.Object, error) {
	if depth > 64 { // Arbitrary recursion depth limit
		return nil, fmt.Errorf("json.loads: maximum recursion depth exceeded during Pylearn object construction")
	}

	switch v := data.(type) {
	case nil:
		return object.NULL, nil
	case bool:
		return object.NativeBoolToBooleanObject(v), nil
	case float64:
		// json.Unmarshal decodes all numbers as float64 initially.
		// We can try to see if it's a whole number and convert to Pylearn Integer.
		if v == float64(int64(v)) {
			return &object.Integer{Value: int64(v)}, nil
		}
		return &object.Float{Value: v}, nil
	case string:
		return &object.String{Value: v}, nil
	case []interface{}: // JSON array -> Pylearn List
		elements := make([]object.Object, len(v))
		for i, item := range v {
			convertedItem, err := ConvertInterfaceToPylearn(item, depth+1)
			if err != nil {
				return nil, err
			}
			elements[i] = convertedItem
		}
		return &object.List{Elements: elements}, nil
	case map[string]interface{}: // JSON object -> Pylearn Dict
		pairs := make(map[object.HashKey]object.DictPair, len(v))
		for key, val := range v {
			pyKey := &object.String{Value: key} // JSON object keys are always strings
			pyVal, err := ConvertInterfaceToPylearn(val, depth+1)
			if err != nil {
				return nil, err
			}
			hashKeyVal, errHash := pyKey.HashKey() // pyKey is *object.String
			if errHash != nil {
				// This should ideally not happen for a simple String key,
				// but good to handle the error from the interface.
				return nil, fmt.Errorf("json.loads: failed to hash string key '%s': %w", pyKey.Value, errHash)
			}
			pairs[hashKeyVal] = object.DictPair{Key: pyKey, Value: pyVal}
		}
		return &object.Dict{Pairs: pairs}, nil
	default:
		return nil, fmt.Errorf("TypeError: cannot convert Go type %T to Pylearn object (from JSON)", v)
	}
}
