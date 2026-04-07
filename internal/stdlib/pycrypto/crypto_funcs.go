// pylearn/internal/stdlib/pycrypto/crypto_funcs.go
package pycrypto

import (
	"crypto/md5"
	"crypto/rand"
	"crypto/sha1"
	"crypto/sha256"
	"encoding/hex"
	"hash"

	"github.com/deniskipeles/pylearn/internal/constants"
	"github.com/deniskipeles/pylearn/internal/object"
)

// hashData is a private helper to perform the hashing logic.
func hashData(h hash.Hash, data []byte) object.Object {
	_, err := h.Write(data)
	if err != nil {
		// This is highly unlikely to happen with standard hashers
		return object.NewError(constants.InternalError, "failed to write to hash: %v", err)
	}
	hashedBytes := h.Sum(nil)
	return &object.String{Value: hex.EncodeToString(hashedBytes)}
}

// pyCryptoMD5 implements crypto.md5(data)
func pyCryptoMD5(ctx object.ExecutionContext, args ...object.Object) object.Object {
	if len(args) != 1 {
		return object.NewError(constants.TypeError, "md5() takes exactly 1 argument (data)")
	}
	var dataToHash []byte
	switch arg := args[0].(type) {
	case *object.String:
		dataToHash = []byte(arg.Value)
	case *object.Bytes:
		dataToHash = arg.Value
	default:
		return object.NewError(constants.TypeError, "md5() argument must be str or bytes")
	}
	return hashData(md5.New(), dataToHash)
}

// pyCryptoSHA1 implements crypto.sha1(data)
func pyCryptoSHA1(ctx object.ExecutionContext, args ...object.Object) object.Object {
	if len(args) != 1 {
		return object.NewError(constants.TypeError, "sha1() takes exactly 1 argument (data)")
	}
	var dataToHash []byte
	switch arg := args[0].(type) {
	case *object.String:
		dataToHash = []byte(arg.Value)
	case *object.Bytes:
		dataToHash = arg.Value
	default:
		return object.NewError(constants.TypeError, "sha1() argument must be str or bytes")
	}
	return hashData(sha1.New(), dataToHash)
}

// pyCryptoSHA256 implements crypto.sha256(data)
func pyCryptoSHA256(ctx object.ExecutionContext, args ...object.Object) object.Object {
	if len(args) != 1 {
		return object.NewError(constants.TypeError, "sha256() takes exactly 1 argument (data)")
	}
	var dataToHash []byte
	switch arg := args[0].(type) {
	case *object.String:
		dataToHash = []byte(arg.Value)
	case *object.Bytes:
		dataToHash = arg.Value
	default:
		return object.NewError(constants.TypeError, "sha256() argument must be str or bytes")
	}
	return hashData(sha256.New(), dataToHash)
}

// pyCryptoRandBytes implements crypto.rand_bytes(n)
func pyCryptoRandBytes(ctx object.ExecutionContext, args ...object.Object) object.Object {
	if len(args) != 1 {
		return object.NewError(constants.TypeError, "rand_bytes() takes exactly 1 argument (n)")
	}
	nObj, ok := args[0].(*object.Integer)
	if !ok {
		return object.NewError(constants.TypeError, "rand_bytes() argument must be an integer")
	}
	n := nObj.Value
	if n < 0 {
		return object.NewError(constants.ValueError, "rand_bytes() argument must be non-negative")
	}

	b := make([]byte, n)
	_, err := rand.Read(b)
	if err != nil {
		return object.NewError(constants.OSError, "failed to generate random bytes: %v", err)
	}

	return &object.Bytes{Value: b}
}


// pyCryptoUnhexlify implements crypto.unhexlify(hexstr)
func pyCryptoUnhexlify(ctx object.ExecutionContext, args ...object.Object) object.Object {
	if len(args) != 1 {
		return object.NewError(constants.TypeError, "unhexlify() takes exactly 1 argument")
	}
	hexStrObj, ok := args[0].(*object.String)
	if !ok {
		return object.NewError(constants.TypeError, "unhexlify() argument must be a string")
	}

	bytes, err := hex.DecodeString(hexStrObj.Value)
	if err != nil {
		// Python's unhexlify raises a ValueError for invalid hex.
		return object.NewError(constants.ValueError, err.Error())
	}
	return &object.Bytes{Value: bytes}
}


// --- Builtin Objects for the native crypto module ---
var (
	CryptoMD5       = &object.Builtin{Name: "crypto.md5", Fn: pyCryptoMD5}
	CryptoSHA1      = &object.Builtin{Name: "crypto.sha1", Fn: pyCryptoSHA1}
	CryptoSHA256    = &object.Builtin{Name: "crypto.sha256", Fn: pyCryptoSHA256}
	CryptoRandBytes = &object.Builtin{Name: "crypto.rand_bytes", Fn: pyCryptoRandBytes}
	CryptoUnhexlify = &object.Builtin{Name: "crypto.unhexlify", Fn: pyCryptoUnhexlify}
)