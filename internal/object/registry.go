// internal/object/registry.go
package object

import (
	"fmt"
	"sync"

	"github.com/deniskipeles/pylearn/internal/constants"
)

var (
    nativeModules      = make(map[string]*Module)
    nativeModulesMutex sync.RWMutex
)

// RegisterNativeModule allows stdlib packages to register themselves centrally.
// This function is EXPORTED.
func RegisterNativeModule(name string, module *Module) {
    nativeModulesMutex.Lock()
    defer nativeModulesMutex.Unlock()
    if _, exists := nativeModules[name]; exists {
         // Handle duplicate registration? Panic or log? For now, overwrite.
         fmt.Printf(constants.REGISTRY_WARN_OVERWRITING_MODULE, name)
    }
    nativeModules[name] = module
}

// GetAllRegisteredModules allows the interpreter/VM setup code to retrieve
// all registered modules. Returns a *copy* of the map.
// This function is EXPORTED.
func GetAllRegisteredModules() map[string]*Module {
    nativeModulesMutex.RLock()
    defer nativeModulesMutex.RUnlock()

    // Return a copy to prevent modification of the original map
    modulesCopy := make(map[string]*Module, len(nativeModules))
    for name, module := range nativeModules {
        modulesCopy[name] = module
    }
    return modulesCopy
}

// GetNativeModule allows retrieving a single module by name.
// Might be useful, but GetAllRegisteredModules is often sufficient for setup.
// This function is EXPORTED.
func GetNativeModule(name string) (*Module, bool) {
	nativeModulesMutex.RLock()
	defer nativeModulesMutex.RUnlock()
	module, found := nativeModules[name]
	if found && module == nil {
		panic(fmt.Sprintf(constants.REGISTRY_CRITICAL_NIL_MODULE, name))
	}
	return module, found
}
