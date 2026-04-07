// internal/object/environment.go
package object

// Environment holds the bindings (variable names to objects)
type Environment struct {
	store map[string]Object // Stores the actual bindings
	outer *Environment      // Pointer to the enclosing environment (for scope)
	globalNames map[string]bool
}

// NewEnvironment creates a new, empty top-level environment.
func NewEnvironment() *Environment {
	s := make(map[string]Object)
	// return &Environment{store: s, outer: nil}
	// <<< INITIALIZE globalNames HERE
	return &Environment{store: s, outer: nil, globalNames: make(map[string]bool)}
}

// NewEnclosedEnvironment creates a new environment nested within an outer one.
// This is used for function calls to create local scopes.
func NewEnclosedEnvironment(outer *Environment) *Environment {
	env := NewEnvironment()
	env.outer = outer
	return env
}

// Get retrieves an object associated with a name.
// It checks the current environment first, then recursively checks outer environments.
func (e *Environment) Get(name string) (Object, bool) {
	obj, ok := e.store[name]
	if !ok && e.outer != nil {
		// If not found here, try the outer scope
		obj, ok = e.outer.Get(name)
	}
	return obj, ok
}

// Set binds a name to an object in the current environment.
// It returns the object that was set.
func (e *Environment) Set(name string, val Object) Object {
	// <<< START OF MODIFICATION >>>
	if e.globalNames[name] {
		// This name is global. Find the root environment and set it there.
		root := e
		for root.outer != nil {
			root = root.outer
		}
		root.store[name] = val
		return val
	}
	// <<< END OF MODIFICATION >>>
	e.store[name] = val
	return val
}

// Update attempts to update the value of an existing variable in the current
// or any outer environment. Returns true if update was successful, false otherwise.
// Crucial for assignment (`=`) to modify existing variables, not just create new ones locally.
func (e *Environment) Update(name string, val Object) (Object, bool) {
	// <<< START OF MODIFICATION >>>
	if e.globalNames[name] {
		root := e
		for root.outer != nil {
			root = root.outer
		}
		// In the global scope, it doesn't matter if it's a new variable or not,
		// the assignment will create or update it there.
		root.store[name] = val
		return val, true
	}
    // <<< END OF MODIFICATION >>>
    _, ok := e.store[name]
    if ok {
        e.store[name] = val // Found in current scope, update it
        return val, true
    }
    if e.outer != nil {
        // Try updating in the outer scope
        return e.outer.Update(name, val)
    }
    // Variable not found in any scope
    return nil, false
}


// --- ADD THIS METHOD ---
// Delete removes a name binding *only* from the current environment scope.
// Returns true if the name existed in this scope and was deleted, false otherwise.
func (e *Environment) Delete(name string) bool {
	_, ok := e.store[name] // Check if it exists in *this* scope
	if ok {
		delete(e.store, name) // Delete from *this* scope's map
		return true
	}
	// Note: This does NOT delete from outer scopes like Python's del would if
	// the variable wasn't local. Implementing that requires more complex scope handling.
	// For now, it only deletes if found directly in the instance's environment.
	return false
}
// --- END ADD ---


// Items returns a map of the bindings stored *only* in this specific environment layer.
// It does NOT include items from outer environments.
// Returns a copy to prevent external modification of the internal store.
func (e *Environment) Items() map[string]Object {
	// Create a copy to return
	items := make(map[string]Object, len(e.store))
	for k, v := range e.store {
		items[k] = v
	}
	return items
}

// RegisterGlobal marks a name as global for the current scope.
func (e *Environment) RegisterGlobal(name string) {
	e.globalNames[name] = true
}
