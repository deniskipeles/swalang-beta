package net

import (
	"github.com/deniskipeles/pylearn/internal/constants"
	"github.com/deniskipeles/pylearn/internal/object"
	"golang.org/x/net/publicsuffix"
)

func pyPublicSuffixTLDPlusOne(ctx object.ExecutionContext, args ...object.Object) object.Object {
	if len(args) != 1 {
		return object.NewError(constants.TypeError, "effective_tld_plus_one() takes 1 argument")
	}
	s, ok := args[0].(*object.String)
	if !ok {
		return object.NewError(constants.TypeError, "argument must be a string")
	}
	domain, err := publicsuffix.EffectiveTLDPlusOne(s.Value)
	if err != nil {
		// This can happen for invalid domains, return None like some libraries do
		return object.NULL
	}
	return &object.String{Value: domain}
}
func createPublicSuffixModule() *object.Module {
	env := object.NewEnvironment()
	env.Set("effective_tld_plus_one", &object.Builtin{Name: "net.publicsuffix.effective_tld_plus_one", Fn: pyPublicSuffixTLDPlusOne})
	return &object.Module{
		Name: "publicsuffix",
		Path: "<builtin_net_publicsuffix>",
		Env:  env,
	}
}
