package object

import (
	"fmt"
	"unsafe"
)

type Pointer struct {
    Value unsafe.Pointer
}

func (p *Pointer) Type() ObjectType    { return "Pointer" }
func (p *Pointer) Inspect() string     { return fmt.Sprintf("Pointer(%p)", p.Value) }
func (p *Pointer) GetAttribute(name string) Object {
    if name == "Address" {
        return &Integer{Value: int64(uintptr(p.Value))}
    }
    return NewError("AttributeError", "Pointer object has no attribute '%s'", name)
}

var _ Object = (*Pointer)(nil)
