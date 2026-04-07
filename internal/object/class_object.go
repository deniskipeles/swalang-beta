package object

import (
	"fmt"
	"reflect"

	"github.com/deniskipeles/pylearn/internal/constants" // Import the constants package
)

// Object types
const (
	SUPER_OBJ    ObjectType = constants.OBJECT_TYPE_SUPER
	CLASS_OBJ    ObjectType = constants.OBJECT_TYPE_CLASS
	INSTANCE_OBJ ObjectType = constants.OBJECT_TYPE_INSTANCE
)

// --- Super Object for super() calls ---

type Super struct {
	SelfInstance Object  // The 'self' instance super() was called with (e.g., an *Instance)
	StartClass   *Class  // The class *after* which to start searching in the MRO
	TargetType   *Class  // The class within which super() was called (MRO_Class)
}

func (s *Super) Type() ObjectType { return SUPER_OBJ }
func (s *Super) Inspect() string {
	typeName := constants.SUPER_OBJECT_INSPECT_UNKNOWN_TYPE
	if s.TargetType != nil {
		typeName = s.TargetType.Name
	}
	selfInspect := constants.SUPER_OBJECT_INSPECT_UNKNOWN_SELF
	if s.SelfInstance != nil {
		selfInspect = s.SelfInstance.Inspect()
	}
	return fmt.Sprintf(constants.SUPER_OBJECT_INSPECT_FORMAT, typeName, selfInspect)
}

var _ Object = (*Super)(nil)

// GetObjectAttribute for Super - uses MRO of TargetType
func (s *Super) GetObjectAttribute(ctx ExecutionContext, name string) (Object, bool) {
	if s.TargetType == nil || s.SelfInstance == nil || s.StartClass == nil {
		return NewError(constants.RuntimeError, constants.SUPER_OBJECT_UNINITIALIZED_ERROR), true
	}

	mro := s.TargetType.MRO
	startIndex := -1
	for i, cls := range mro {
		if cls == s.StartClass {
			startIndex = i
			break
		}
	}

	if startIndex == -1 {
		return NewError(constants.RuntimeError, fmt.Sprintf(constants.SUPER_OBJECT_START_CLASS_NOT_IN_MRO_ERROR, s.StartClass.Name, s.TargetType.Name)), true
	}

	// Search MRO *after* StartClass
	for i := startIndex + 1; i < len(mro); i++ {
		currentSearchClass := mro[i]
		var attrValue Object
		foundInCurrent := false

		if methodFunc, methodOk := currentSearchClass.Methods[name]; methodOk {
			attrValue = methodFunc
			foundInCurrent = true
		}
		if !foundInCurrent && currentSearchClass.ClassVariables != nil {
			if classVar, varOk := currentSearchClass.ClassVariables.Get(name); varOk {
				attrValue = classVar
				foundInCurrent = true
			}
		}
		if foundInCurrent {
			// Descriptor handling for super()
			switch desc := attrValue.(type) {
			case *StaticMethod:
				return desc.Function, true
			case *ClassMethod:
				// Bound to s.TargetType or s.SelfInstance if it's a class itself
				classToBindTo := s.TargetType
				if selfCls, isCls := s.SelfInstance.(*Class); isCls {
					classToBindTo = selfCls
				}
				return &BoundMethod{Instance: classToBindTo, Method: desc.Function}, true
			default:
				// --- THIS IS THE FIX ---
				// If the found attribute is a regular Pylearn function, bind it to the
				// instance from the super() call to create a BoundMethod.
				if regularFunc, isFunc := attrValue.(*Function); isFunc {
					if _, isInst := s.SelfInstance.(*Instance); isInst {
						return &BoundMethod{Instance: s.SelfInstance, Method: regularFunc}, true
					}
					// If self is a class, return the unbound function (for classmethod behavior via super)
					return regularFunc, true
				}

				// If the found attribute is a native Go BUILTIN method, we must also bind it.
				// We do this by creating a new Builtin whose closure captures `self`.
				if builtinFunc, isBuiltin := attrValue.(*Builtin); isBuiltin {
					boundBuiltin := &Builtin{
						Name:            builtinFunc.Name,
						AcceptsKeywords: builtinFunc.AcceptsKeywords,
						Fn: func(callCtx ExecutionContext, scriptProvidedArgs ...Object) Object {
							// This closure is the new "bound" method.
							// It prepends the 'self' instance from the super() call to the
							// argument list before calling the original native Go function.
							finalArgs := make([]Object, 1+len(scriptProvidedArgs))
							finalArgs[0] = s.SelfInstance
							copy(finalArgs[1:], scriptProvidedArgs)
							return builtinFunc.Fn(callCtx, finalArgs...)
						},
					}
					return boundBuiltin, true
				}
				// --- END OF FIX ---

				// If the attribute was not a function or builtin (e.g., a class variable),
				// return it directly.
				return attrValue, true // Other class variables
			}
		}
	}
	return nil, false // Attribute not found
}

var _ AttributeGetter = (*Super)(nil)

// --- Class Object ---

type Class struct {
	Name           string
	Superclasses   []*Class             // Direct base classes (multiple inheritance support)
	MRO            []*Class             // Method Resolution Order (linearized)
	Methods        map[string]Object
	ClassVariables *Environment         // Class variables defined in *this* class
}

func (c *Class) Type() ObjectType { return CLASS_OBJ }
func (c *Class) Inspect() string  { return fmt.Sprintf(constants.CLASS_OBJECT_INSPECT_FORMAT, c.Name) }
func (c *Class) HashKey() (HashKey, error) {
	return HashKey{Type: c.Type(), Value: uint64(reflect.ValueOf(c).Pointer())}, nil
}

// GetObjectAttribute for Class - uses MRO to find class/static methods and variables
func (c *Class) GetObjectAttribute(ctx ExecutionContext, name string) (Object, bool) {
	if name == constants.CLASS_OBJECT_DUNDER_NAME_ATTR || name == constants.CLASS_OBJECT_NAME_ATTR {
		return &String{Value: c.Name}, true
	}
	for _, currentClassInMRO := range c.MRO {
		var attrValue Object
		foundInCurrent := false

		if methodFunc, methodOk := currentClassInMRO.Methods[name]; methodOk {
			attrValue = methodFunc
			foundInCurrent = true
		}
		if !foundInCurrent && currentClassInMRO.ClassVariables != nil {
			if classVar, varOk := currentClassInMRO.ClassVariables.Get(name); varOk {
				attrValue = classVar
				foundInCurrent = true
			}
		}

		if foundInCurrent {
			switch desc := attrValue.(type) {
			case *StaticMethod:
				return desc.Function, true
			case *ClassMethod:
				// When accessed on a class, a classmethod is bound to that class
				return &BoundMethod{Instance: c, Method: desc.Function}, true
			default:
				// Return other attributes like regular (unbound) functions or class variables
				return attrValue, true
			}
		}
	}
	return nil, false // Attribute not found in the entire hierarchy
}

var _ Object = (*Class)(nil)
var _ Hashable = (*Class)(nil)
var _ AttributeGetter = (*Class)(nil)

// --- Instance Object ---

type Instance struct {
	Class *Class
	Env   *Environment // Instance-specific attributes
}

func (i *Instance) Type() ObjectType { return INSTANCE_OBJ }
func (i *Instance) Inspect() string {
	if i.Class == nil {
		return fmt.Sprintf(constants.INSTANCE_OBJECT_UNINITIALIZED_FORMAT, i)
	}
	return fmt.Sprintf(constants.INSTANCE_OBJECT_INSPECT_FORMAT, i.Class.Name, i)
}

// GetObjectAttribute for Instance - uses MRO for methods and class variables
// GetObjectAttribute for Instance - uses MRO for methods and class variables
func (i *Instance) GetObjectAttribute(ctx ExecutionContext, name string) (Object, bool) {
	// 1. Check instance's own environment (instance variables) first
	if i.Env != nil {
		if val, ok := i.Env.Get(name); ok {
			return val, true
		}
	}
	if i.Class == nil {
		return NewError(constants.InternalError, constants.INSTANCE_OBJECT_NO_CLASS_ERROR), true
	}

	// 2. Search class MRO for methods and class variables
	for _, currentClassInMRO := range i.Class.MRO {
		var attrValue Object
		foundInCurrent := false

		if methodFunc, methodOk := currentClassInMRO.Methods[name]; methodOk {
			attrValue = methodFunc
			foundInCurrent = true
		}
		if !foundInCurrent && currentClassInMRO.ClassVariables != nil {
			if classVar, varOk := currentClassInMRO.ClassVariables.Get(name); varOk {
				attrValue = classVar
				foundInCurrent = true
			}
		}

		if foundInCurrent {
			// Handle descriptors (staticmethod, classmethod) and bind methods
			switch desc := attrValue.(type) {
			case *StaticMethod:
				return desc.Function, true
			case *ClassMethod:
				// Bind classmethod to the instance's class
				return &BoundMethod{Instance: i.Class, Method: desc.Function}, true
			case *Property:
				// It's a property. We need to call its getter function.
				if desc.FGet == nil || desc.FGet == NULL {
					return NewError(constants.AttributeError, "unreadable attribute"), true
				}
				// The getter function (fget) needs to be called with the instance `i` as its `self` argument.
				return ctx.Execute(desc.FGet, i), true

			// --- THIS IS THE CRITICAL FIX ---
			case *Builtin:
				// It's a native Go method (like Exception.__init__ or __str__).
				// We must bind it to the instance `i` by creating a new Builtin
				// whose closure captures `i` as the `self` argument.
				boundBuiltin := &Builtin{
					Name:            desc.Name,
					AcceptsKeywords: desc.AcceptsKeywords,
					Fn: func(callCtx ExecutionContext, scriptProvidedArgs ...Object) Object {
						// This closure is the new "bound" method.
						// It prepends the 'self' instance (`i`) to the argument list
						// before calling the original native Go function (`desc.Fn`).
						finalArgs := make([]Object, 1+len(scriptProvidedArgs))
						finalArgs[0] = i // Prepend `self`
						copy(finalArgs[1:], scriptProvidedArgs)
						return desc.Fn(callCtx, finalArgs...)
					},
				}
				return boundBuiltin, true
			// --- END OF FIX ---

			default:
				// If it's a regular Pylearn function from the class's methods dict, bind it.
				if regularFunc, isFunc := attrValue.(*Function); isFunc {
					return &BoundMethod{Instance: i, Method: regularFunc}, true
				}
				// Otherwise, it's a class variable. Return it directly.
				return attrValue, true
			}
		}
	}
	return nil, false // Attribute not found
}

var _ Object = (*Instance)(nil)
var _ AttributeGetter = (*Instance)(nil)

// --- C3 Linearization Algorithm for MRO ---

// ComputeMRO calculates the Method Resolution Order for a class using C3 linearization.
// It needs the class itself and the global Class representing "object".
func ComputeMRO(class *Class, objectClass *Class) ([]*Class, error) {
	if class == nil {
		return nil, fmt.Errorf(constants.MRO_COMPUTE_NIL_CLASS_ERROR)
	}
	if class == objectClass {
		return []*Class{objectClass}, nil
	}
	if class.MRO != nil { // MRO already computed
		return class.MRO, nil
	}

	// The list of sequences to merge, starting with the MRO of each parent
	sequences := [][]*Class{}
	for _, base := range class.Superclasses {
		baseMRO, err := ComputeMRO(base, objectClass) // Recursively compute for parents
		if err != nil {
			return nil, fmt.Errorf(constants.MRO_COMPUTATION_FAILED_BASE_CLASS, base.Name, class.Name, err)
		}
		sequences = append(sequences, baseMRO)
	}
	// Also add the list of the direct parents themselves
	sequences = append(sequences, class.Superclasses)

	// The final MRO starts with the class itself
	mergedMRO := []*Class{class}

	for {
		// If all sequences are empty, we are done
		allEmpty := true
		for _, seq := range sequences {
			if len(seq) > 0 {
				allEmpty = false
				break
			}
		}
		if allEmpty {
			break
		}

		// Find the first "good head" candidate
		var candidate *Class = nil
		for _, seq := range sequences {
			if len(seq) == 0 {
				continue
			}
			head := seq[0]
			isGoodHead := true
			// Check if this head appears in the *tail* of any other sequence
			for _, otherSeq := range sequences {
				for j, tailElement := range otherSeq {
					if j > 0 && tailElement == head { // j > 0 means it's in the tail
						isGoodHead = false
						break
					}
				}
				if !isGoodHead {
					break
				}
			}

			if isGoodHead {
				candidate = head
				break
			}
		}

		if candidate == nil {
			// This indicates an inconsistent class hierarchy (MRO error)
			return nil, fmt.Errorf(constants.MRO_CONSISTENCY_ERROR, class.Name)
		}

		mergedMRO = append(mergedMRO, candidate)

		// Remove the candidate from the head of all sequences
		newSequences := [][]*Class{}
		for _, seq := range sequences {
			if len(seq) > 0 && seq[0] == candidate {
				// If the sequence becomes empty after removing the head, it's fine
				newSequences = append(newSequences, seq[1:])
			} else {
				newSequences = append(newSequences, seq)
			}
		}
		sequences = newSequences
	}

	class.MRO = mergedMRO
	return mergedMRO, nil
}

// --- ObjectClass - Root of the class hierarchy ---

// ObjectClass is the root of the class hierarchy.
// It needs to be initialized once.
var ObjectClass *Class

func init() {
	// Initialize ObjectClass. It has no superclasses and its MRO is just itself.
	ObjectClass = &Class{
		Name:           constants.OBJECT_CLASS_NAME,
		Superclasses:   []*Class{},
		Methods:        make(map[string]Object), // object can have dunders like __str__, __repr__
		ClassVariables: NewEnvironment(),
	}
	// ObjectClass = &Class{
	// 	Name:           "object",
	// 	Superclasses:   []*Class{},
	// 	Methods:        make(map[string]*Function), // object can have dunders like __str__, __repr__
	// 	ClassVariables: NewEnvironment(),
	// }
	ObjectClass.MRO = []*Class{ObjectClass} // MRO of object is [object]

	// Add default __str__ and __repr__ to object that other classes can inherit
	// This needs the object.Function type to be defined, and a way to create them from Go.
	// This is a bit circular if Function needs ast.BlockStatement.
	// For now, these can be placeholder builtins or actual simple Pylearn functions.
	// Or, the fallback in str()/repr() builtins can handle this if no user-defined version.

	// Example placeholder for object.__repr__ (could be a Builtin that calls Inspect)
	// This would be more complex to set up here correctly.
	// The interpreter's str()/repr() builtins will provide default behavior if not found via MRO.
}