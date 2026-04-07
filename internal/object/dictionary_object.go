// pylearn/internal/object/dictionary_object.go
package object

import (
	"bytes"
	"fmt"
	"sort"
	"strings"

	"github.com/deniskipeles/pylearn/internal/constants"
)

// Dict (Mutable Mapping, Not Hashable)
type Dict struct{ Pairs map[HashKey]DictPair }
type DictPair struct{ Key, Value Object }

func (d *Dict) Type() ObjectType { return DICT_OBJ }
func (d *Dict) Inspect() string {
	var out bytes.Buffer
	pairs := []string{}
	keys := make([]HashKey, 0, len(d.Pairs))
	for k := range d.Pairs {
		keys = append(keys, k)
	}
	// Sort keys for deterministic Inspect output
	sort.Slice(keys, func(i, j int) bool {
		// Basic sort, can be improved if keys are complex
		keyIStr := d.Pairs[keys[i]].Key.Inspect()
		keyJStr := d.Pairs[keys[j]].Key.Inspect()
		return keyIStr < keyJStr
	})
	for _, k := range keys {
		pair := d.Pairs[k]
		pairs = append(pairs, fmt.Sprintf(constants.DICT_INSPECT_PAIR_SEPARATOR_FORMAT, pair.Key.Inspect(), pair.Value.Inspect()))
	}
	out.WriteString(constants.DICT_INSPECT_OPEN_BRACE)
	out.WriteString(strings.Join(pairs, constants.DICT_INSPECT_SEPARATOR))
	out.WriteString(constants.DICT_INSPECT_CLOSE_BRACE)
	return out.String()
}

// Get is a Go helper method to retrieve a value from the dictionary using a Go string key.
// It returns the Pylearn object and a boolean indicating if the key was found.
func (d *Dict) Get(key string) (Object, bool) {
	hashKey, err := (&String{Value: key}).HashKey()
	if err != nil {
		// This should not happen for a string key
		return nil, false
	}
	pair, found := d.Pairs[hashKey]
	if !found {
		return nil, false
	}
	return pair.Value, true
}

// GetObjectItem and SetObjectItem remain the same (used for d[key] and d[key]=value)
func (d *Dict) GetObjectItem(key Object) Object {
	hashableKey, ok := key.(Hashable)
	if !ok {
		return NewError(constants.TypeError, constants.DICT_UNHASHABLE_TYPE_ERROR_FORMAT, key.Type())
	}
	dictMapKey, err := hashableKey.HashKey()
	if err != nil {
		return NewError(constants.TypeError, constants.DICT_FAILED_TO_HASH_KEY_ERROR_FORMAT, err)
	}
	pair, found := d.Pairs[dictMapKey]
	if !found {
		return NewError(constants.KeyError, constants.DICT_KEY_ERROR_FORMAT, key.Inspect())
	}
	return pair.Value
}

func (d *Dict) SetObjectItem(key Object, value Object) Object {
	hashableKey, ok := key.(Hashable)
	if !ok {
		return NewError(constants.TypeError, constants.DICT_UNHASHABLE_TYPE_ERROR_FORMAT, key.Type())
	}
	hashKeyVal, err := hashableKey.HashKey()
	if err != nil {
		return NewError(constants.TypeError, constants.DICT_FAILED_TO_HASH_KEY_ERROR_FORMAT, err)
	}
	if d.Pairs == nil {
		d.Pairs = make(map[HashKey]DictPair)
	}
	d.Pairs[hashKeyVal] = DictPair{Key: key, Value: value}
	return nil
}

// --- Go functions for Dict methods ---

func pyDictGetFn(ctx ExecutionContext, args ...Object) Object {
	// Expects: self (Dict), key, [default_value]
	numScriptArgs := len(args) - 1 // Number of arguments provided by the script
	if numScriptArgs < 1 || numScriptArgs > 2 {
		return NewError(constants.TypeError, constants.DICT_GET_ARG_COUNT_ERROR, numScriptArgs)
	}
	self, ok := args[0].(*Dict)
	if !ok {
		return NewError(constants.TypeError, constants.DICT_GET_ON_DICT_ERROR)
	}

	keyToGet := args[1]
	// Declare defaultValue as object.Object
	var defaultValue Object = NULL // Initialize with Pylearn's NULL
	if numScriptArgs == 2 {        // Check using numScriptArgs
		defaultValue = args[2] // This is now a valid assignment
	}

	hashableKey, ok := keyToGet.(Hashable)
	if !ok {
		return defaultValue
	}
	hashedKey, err := hashableKey.HashKey()
	if err != nil {
		return defaultValue
	}

	if pair, found := self.Pairs[hashedKey]; found {
		return pair.Value
	}
	return defaultValue
}

func pyDictKeysFn(ctx ExecutionContext, args ...Object) Object {
	if len(args) != 1 { // self
		return NewError(constants.TypeError, constants.DICT_KEYS_ON_DICT_ERROR, len(args)-1)
	}
	self, ok := args[0].(*Dict)
	if !ok {
		return NewError(constants.TypeError, constants.DICT_METHOD_KEYS_ON_DICT_ERROR)
	}

	// keysList := make([]Object, 0, len(self.Pairs))
	itemsList := make([]Object, 0, len(self.Pairs))
	for _, pair := range self.Pairs {
		itemPair := &Tuple{Elements: []Object{pair.Key, pair.Value}}
		// keysList = append(keysList, pair.Key)
		itemsList = append(itemsList, itemPair)
	}
	// return &List{Elements: keysList}
	return &List{Elements: itemsList}
}

func pyDictValuesFn(ctx ExecutionContext, args ...Object) Object {
	if len(args) != 1 { // self
		return NewError(constants.TypeError, constants.DICT_VALUES_ON_DICT_ERROR, len(args)-1)
	}
	self, ok := args[0].(*Dict)
	if !ok {
		return NewError(constants.TypeError, constants.DICT_METHOD_VALUES_ON_DICT_ERROR)
	}

	valuesList := make([]Object, 0, len(self.Pairs))
	for _, pair := range self.Pairs {
		valuesList = append(valuesList, pair.Value)
	}
	return &List{Elements: valuesList}
}

func pyDictItemsFn(ctx ExecutionContext, args ...Object) Object {
	if len(args) != 1 { // self
		return NewError(constants.TypeError, constants.DICT_ITEMS_ON_DICT_ERROR, len(args)-1)
	}
	self, ok := args[0].(*Dict)
	if !ok {
		return NewError(constants.TypeError, constants.DICT_METHOD_ITEMS_ON_DICT_ERROR)
	}

	itemsList := make([]Object, 0, len(self.Pairs))
	for _, pair := range self.Pairs {
		// Using List for (key, value) pairs. Replace with Tuple if available.
		itemPair := &List{Elements: []Object{pair.Key, pair.Value}}
		// itemPair := &Tuple{Elements: []Object{pair.Key, pair.Value}} // If Tuple is ready
		itemsList = append(itemsList, itemPair)
	}
	return &List{Elements: itemsList}
}

func pyDictUpdateFn(ctx ExecutionContext, args ...Object) Object {
	// Expects: self (Dict), other_dict_or_iterable
	if len(args) != 2 {
		return NewError(constants.TypeError, constants.DICT_UPDATE_ARG_COUNT_ERROR, len(args)-1)
	}
	self, okSelf := args[0].(*Dict)
	if !okSelf {
		return NewError(constants.TypeError, constants.DICT_UPDATE_ON_DICT_ERROR)
	}

	otherArg := args[1]
	if otherDict, okOther := otherArg.(*Dict); okOther {
		for hashKey, pair := range otherDict.Pairs {
			self.Pairs[hashKey] = pair
		}
	} else {
		// TODO: Python's update also accepts iterables of (key, value) pairs.
		// For now, only Dict-to-Dict update.
		return NewError(constants.TypeError, constants.DICT_UPDATE_ARG_TYPE_ERROR, otherArg.Type())
	}
	return NULL
}

func pyDictLenFn(ctx ExecutionContext, args ...Object) Object {
	if len(args) != 1 { // self
		return NewError(constants.TypeError, constants.DICT_LEN_ON_DICT_ERROR)
	}
	self, ok := args[0].(*Dict)
	if !ok {
		return NewError(constants.TypeError, constants.DICT_LEN_ON_DICT_ERROR)
	}
	return &Integer{Value: int64(len(self.Pairs))}
}

func pyDictContainsFn(ctx ExecutionContext, args ...Object) Object {
	// Expects: self (Dict), key
	if len(args) != 2 {
		return NewError(constants.TypeError, constants.DICT_CONTAINS_ARG_COUNT_ERROR)
	}
	self, okSelf := args[0].(*Dict)
	if !okSelf {
		return NewError(constants.TypeError, constants.DICT_CONTAINS_ON_DICT_ERROR)
	}

	keyToTest := args[1]
	hashableKey, okHashable := keyToTest.(Hashable)
	if !okHashable {
		return FALSE
	} // Non-hashable key cannot be in the dict

	hashedKey, err := hashableKey.HashKey()
	if err != nil {
		return FALSE
	} // Error hashing key

	_, found := self.Pairs[hashedKey]
	return NativeBoolToBooleanObject(found)
}

// GetObjectAttribute for Dict to expose methods
func (d *Dict) GetObjectAttribute(ctx ExecutionContext, name string) (Object, bool) {
	// Helper to create the method Builtin objects
	makeMethod := func(methodName string, goFn BuiltinFunction) *Builtin {
		return &Builtin{
			Name: constants.DictMethodPrefix + methodName,
			Fn: func(callCtx ExecutionContext, scriptProvidedArgs ...Object) Object {
				methodArgs := make([]Object, 0, 1+len(scriptProvidedArgs))
				methodArgs = append(methodArgs, d) // Prepend self (the Dict 'd')
				methodArgs = append(methodArgs, scriptProvidedArgs...)
				return goFn(callCtx, methodArgs...)
			},
		}
	}

	switch name {
	case constants.DictGetMethodName:
		return makeMethod(constants.DictGetMethodName, pyDictGetFn), true
	case constants.DictKeysMethodName:
		return makeMethod(constants.DictKeysMethodName, pyDictKeysFn), true
	case constants.DictValuesMethodName:
		return makeMethod(constants.DictValuesMethodName, pyDictValuesFn), true
	case constants.DictItemsMethodName:
		return makeMethod(constants.DictItemsMethodName, pyDictItemsFn), true
	case constants.DictUpdateMethodName:
		return makeMethod(constants.DictUpdateMethodName, pyDictUpdateFn), true
	case constants.DunderLen:
		return makeMethod(constants.DunderLen, pyDictLenFn), true
	case constants.DunderContains:
		return makeMethod(constants.DunderContains, pyDictContainsFn), true
	}
	return nil, false
}

var _ Object = (*Dict)(nil)
var _ ItemGetter = (*Dict)(nil)
var _ ItemSetter = (*Dict)(nil)
var _ AttributeGetter = (*Dict)(nil)
