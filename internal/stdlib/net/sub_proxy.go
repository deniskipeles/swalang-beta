// internal/stdlib/net/sub_proxy.go
package net

import (
	"github.com/deniskipeles/pylearn/internal/constants"
	"github.com/deniskipeles/pylearn/internal/object"
	"golang.org/x/net/proxy"
)

// pyProxySOCKS5 implements proxy.socks5(net, addr, auth=None, forward=None)
func pyProxySOCKS5(ctx object.ExecutionContext, args ...object.Object) object.Object {
	if len(args) < 2 || len(args) > 4 {
		return object.NewError(constants.TypeError, "socks5() takes 2 to 4 arguments")
	}
	network, ok := args[0].(*object.String)
	if !ok {
		return object.NewError(constants.TypeError, "network must be a string")
	}
	addr, ok := args[1].(*object.String)
	if !ok {
		return object.NewError(constants.TypeError, "address must be a string")
	}

	var auth *proxy.Auth
	if len(args) >= 3 && args[2] != object.NULL {
		authDict, ok := args[2].(*object.Dict)
		if !ok {
			return object.NewError(constants.TypeError, "auth must be a dict or None")
		}
		// <<< START FIX >>>
		// Use the new Get helper method on the Dict object
		user, _ := authDict.Get("user")
		pass, _ := authDict.Get("password")
		// <<< END FIX >>>
		userStr, uOk := user.(*object.String)
		passStr, pOk := pass.(*object.String)
		if uOk && pOk {
			auth = &proxy.Auth{User: userStr.Value, Password: passStr.Value}
		}
	}

	var forward proxy.Dialer = proxy.Direct
	if len(args) == 4 && args[3] != object.NULL {
		forwardDialer, ok := args[3].(*ProxyDialer)
		if !ok {
			return object.NewError(constants.TypeError, "forward must be a Dialer object")
		}
		forward = forwardDialer.GoDialer
	}

	dialer, err := proxy.SOCKS5(network.Value, addr.Value, auth, forward)
	if err != nil {
		return object.NewError("ProxyError", err.Error())
	}

	return &ProxyDialer{GoDialer: dialer}
}

// pyProxyDirect returns the direct dialer.
func pyProxyDirect(ctx object.ExecutionContext, args ...object.Object) object.Object {
	if len(args) != 0 {
		return object.NewError(constants.TypeError, "direct() takes no arguments")
	}
	return &ProxyDialer{GoDialer: proxy.Direct}
}

func createProxyModule() *object.Module {
	env := object.NewEnvironment()
	env.Set("socks5", &object.Builtin{Name: "net.proxy.socks5", Fn: pyProxySOCKS5})
	env.Set("direct", &object.Builtin{Name: "net.proxy.direct", Fn: pyProxyDirect})
	env.Set("Dialer", ProxyDialerClass)
	env.Set("ProxyError", ProxyErrorClass)
	return &object.Module{
		Name: "proxy",
		Path: "<builtin_net_proxy>",
		Env:  env,
	}
}
