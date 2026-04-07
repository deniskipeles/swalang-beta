// internal/stdlib/net/funcs.go
package net

import (
	"net"

	"github.com/deniskipeles/pylearn/internal/constants"
	"github.com/deniskipeles/pylearn/internal/object"
	"golang.org/x/net/proxy"
)

// pyNetDial implements net.dial(network, address, *, dialer=None)
func pyNetDial(ctx object.ExecutionContext, args ...object.Object) object.Object {
	if len(args) < 2 {
		return object.NewError(constants.TypeError, "dial() missing required arguments (network, address)")
	}
	network, ok := args[0].(*object.String)
	if !ok {
		return object.NewError(constants.TypeError, "network must be a string")
	}
	address, ok := args[1].(*object.String)
	if !ok {
		return object.NewError(constants.TypeError, "address must be a string")
	}

	// Default to a direct connection
	var dialer proxy.Dialer = proxy.Direct

	// <<< START FIX >>>
	// Check for kwargs passed as the last argument (a Dict)
	if len(args) > 2 {
		kwargs, ok := args[len(args)-1].(*object.Dict)
		if ok {
			if dialerValObj, found := kwargs.Get("dialer"); found {
				if dialerValObj != object.NULL {
					pylearnDialer, isDialer := dialerValObj.(*ProxyDialer)
					if !isDialer {
						return object.NewError(constants.TypeError, "dialer keyword argument must be a net.proxy.Dialer object")
					}
					dialer = pylearnDialer.GoDialer
				}
			}
		} else {
			// This path is for legacy positional argument support, but keyword is preferred.
			// It's better to enforce keywords for optional args.
			// We can assume if len(args) > 2 and the last arg is not a Dict, it's an error
			// unless we explicitly support positional optional args.
			// For now, let's assume the user passed a positional dialer.
			pylearnDialer, ok := args[2].(*ProxyDialer)
			if !ok {
				return object.NewError(constants.TypeError, "optional argument 3 must be a net.proxy.Dialer object")
			}
			dialer = pylearnDialer.GoDialer
		}
	}
	// <<< END FIX >>>

	conn, err := dialer.Dial(network.Value, address.Value)
	if err != nil {
		return object.NewError("ConnectionError", err.Error())
	}

	return &Connection{GoConn: conn}
}

// pyNetListen implements net.listen(network, address)
func pyNetListen(ctx object.ExecutionContext, args ...object.Object) object.Object {
	if len(args) != 2 {
		return object.NewError(constants.TypeError, "listen() takes exactly 2 arguments (network, address)")
	}
	network, ok := args[0].(*object.String)
	if !ok {
		return object.NewError(constants.TypeError, "network must be a string")
	}
	address, ok := args[1].(*object.String)
	if !ok {
		return object.NewError(constants.TypeError, "address must be a string")
	}

	listener, err := net.Listen(network.Value, address.Value)
	if err != nil {
		return object.NewError("IOError", err.Error())
	}

	return &Listener{GoListener: listener}
}

// pyNetParseIP implements net.parse_ip(ip_string)
func pyNetParseIP(ctx object.ExecutionContext, args ...object.Object) object.Object {
	if len(args) != 1 {
		return object.NewError(constants.TypeError, "parse_ip() takes exactly 1 argument")
	}
	ipStr, ok := args[0].(*object.String)
	if !ok {
		return object.NewError(constants.TypeError, "ip address must be a string")
	}

	ip := net.ParseIP(ipStr.Value)
	if ip == nil {
		return object.NULL // Return None for invalid IP strings
	}

	// How to represent an IP address? A Bytes object is a good fit.
	return &object.Bytes{Value: []byte(ip)}
}

var (
	Dial = &object.Builtin{
		Name: "net.dial",
		Fn:   pyNetDial,
		AcceptsKeywords: map[string]bool{
			"dialer": true,
		},
	}
	Listen  = &object.Builtin{Name: "net.listen", Fn: pyNetListen}
	ParseIP = &object.Builtin{Name: "net.parse_ip", Fn: pyNetParseIP}
)
