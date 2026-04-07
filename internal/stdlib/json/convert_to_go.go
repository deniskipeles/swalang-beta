package pyjson

import (
	"fmt"
	"github.com/deniskipeles/pylearn/internal/object"
)

// convertPylearnToInterface converts a Pylearn object to its Go equivalent
// suitable for json.Marshal.
// It handles nested structures recursively.
// `depth` is used to prevent infinite recursion on circular structures.
func ConvertPylearnToInterface(obj object.Object, depth int) (interface{}, error) {
	if depth > 64 { // Arbitrary recursion depth limit
		return nil, fmt.Errorf("json.dumps: maximum recursion depth exceeded")
	}

	switch o := obj.(type) {
	case *object.Null:
		return nil, nil // Pylearn None becomes Go nil
	case *object.Boolean:
		return o.Value, nil
	case *object.Integer:
		// json.Marshal handles int64 fine
		return o.Value, nil
	case *object.Float:
		// json.Marshal handles float64 fine
		return o.Value, nil
	case *object.String:
		return o.Value, nil
	case *object.List:
		arr := make([]interface{}, len(o.Elements))
		for i, elem := range o.Elements {
			convertedElem, err := ConvertPylearnToInterface(elem, depth+1)
			if err != nil {
				return nil, err // Propagate error from nested conversion
			}
			arr[i] = convertedElem
		}
		return arr, nil
	case *object.Dict:
		// For JSON, keys must be strings.
		// If Pylearn Dict allows non-string keys, we need to decide how to handle them.
		// Option 1: Error if non-string key encountered.
		// Option 2: Try to str() the key (less ideal as it changes key type).
		// Let's go with Option 1 for now.
		m := make(map[string]interface{}, len(o.Pairs))
		for _, pair := range o.Pairs {
			keyStr, ok := pair.Key.(*object.String)
			if !ok {
				return nil, fmt.Errorf("json.dumps: dictionary keys must be strings, got %s", pair.Key.Type())
			}
			convertedVal, err := ConvertPylearnToInterface(pair.Value, depth+1)
			if err != nil {
				return nil, err // Propagate error
			}
			m[keyStr.Value] = convertedVal
		}
		return m, nil
	// TODO: Add *object.Bytes later? json.Marshal handles []byte.
	default:
		// What to do with Functions, Classes, Instances, Modules, etc.?
		// Python's json.dumps raises TypeError for un-serializable types.
		return nil, fmt.Errorf("TypeError: Object of type '%s' is not JSON serializable", obj.Type())
	}
}
