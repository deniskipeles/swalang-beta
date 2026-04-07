// In internal/stdlib/pyre/module.go

package pyre
import (
	"github.com/deniskipeles/pylearn/internal/object"
)
func init() {
	env := object.NewEnvironment()
	env.Set("compile", ReCompile)
	env.Set("search", ReSearch)
	env.Set("match", ReMatch)
	env.Set("findall", ReFindAll)
	// =========================== START: Add this line ===========================
	env.Set("escape", ReEscape)
	// =========================== END: Add this line =============================
	
	reModule := &object.Module{
		Name: "re",
		Path: "<builtin>",
		Env:  env,
	}
	object.RegisterNativeModule("re", reModule)
}