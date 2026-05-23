package object

import (
	"fmt"
	"unsafe"

	"github.com/deniskipeles/pylearn/internal/constants"
)

type Pointer struct {
	Value unsafe.Pointer
}

func (p *Pointer) Type() ObjectType    { return constants.OBJECT_TYPE_POINTER }
func (p *Pointer) Inspect() string     { return fmt.Sprintf(constants.OBJECT_POINTER_INSPECT_FORMAT, p.Value) }
func (p *Pointer) GetAttribute(name string) Object {
	if name == constants.OBJECT_POINTER_ADDRESS_ATTR {
		return &Integer{Value: int64(uintptr(p.Value))}
	}
	return NewError(constants.AttributeError, constants.OBJECT_POINTER_HAS_NO_ATTR_ERROR, name)
}

var _ Object = (*Pointer)(nil)