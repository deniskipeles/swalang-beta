package object

import (
	"github.com/deniskipeles/pylearn/internal/constants"
)

// ItemDeleter allows deleting items using square brackets (del obj[key])
type ItemDeleter interface {
	DeleteObjectItem(key Object) Object // Returns NULL on success, Error on failure
}

// Add this new method to the Dict struct
func (d *Dict) DeleteObjectItem(key Object) Object {
	hashableKey, ok := key.(Hashable)
	if !ok {
		return NewError(constants.KeyError, constants.DICT_KEY_ERROR_FORMAT, key.Inspect())
	}

	dictMapKey, err := hashableKey.HashKey()
	if err != nil {
		return NewError(constants.TypeError, constants.DICT_FAILED_TO_HASH_KEY_ERROR_FORMAT, err)
	}

	_, found := d.Pairs[dictMapKey]
	if !found {
		return NewError(constants.KeyError, constants.DICT_KEY_ERROR_FORMAT, key.Inspect())
	}

	delete(d.Pairs, dictMapKey)
	return NULL // Success
}

// Add this line at the bottom of the file to confirm implementation
var _ ItemDeleter = (*Dict)(nil)
