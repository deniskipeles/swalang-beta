// pylearn/internal/stdlib/pycrypto/module.go
package pycrypto

import (
	"github.com/deniskipeles/pylearn/internal/object"
)

func init() {
	// Environment for the native '_crypto_native' module
	env := object.NewEnvironment()

	// Add the raw hashing functions
	env.Set("md5", CryptoMD5)
	env.Set("sha1", CryptoSHA1)
	env.Set("sha256", CryptoSHA256)
	env.Set("rand_bytes", CryptoRandBytes)
	env.Set("unhexlify", CryptoUnhexlify)

	// Create the Module object
	cryptoModule := &object.Module{
		Name: "_crypto_native", // Internal name
		Path: "<builtin_crypto>",
		Env:  env,
	}

	// Register the module with Pylearn's central registry
	object.RegisterNativeModule("_crypto_native", cryptoModule)
}