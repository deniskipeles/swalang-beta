// internal/stdlib/net/sub_icmp.go
package net

import (
	"net"
	"time"

	"github.com/deniskipeles/pylearn/internal/constants"
	"github.com/deniskipeles/pylearn/internal/object"
	"golang.org/x/net/icmp"
	"golang.org/x/net/ipv4"
)

// pyIcmpPing implements icmp.ping(destination_address, timeout=4.0)
func pyIcmpPing(ctx object.ExecutionContext, args ...object.Object) object.Object {
	// <<< START FIX >>>
	// Argument Parsing with Keyword Support
	if len(args) < 1 {
		return object.NewError(constants.TypeError, "ping() missing 1 required positional argument: 'destination'")
	}
	addrObj, ok := args[0].(*object.String)
	if !ok {
		return object.NewError(constants.TypeError, "destination must be a string")
	}

	timeout := 4.0 // Default timeout

	// Check for kwargs passed as the last argument (a Dict)
	if len(args) > 1 {
		kwargs, ok := args[len(args)-1].(*object.Dict)
		if ok {
			if timeoutValObj, found := kwargs.Get("timeout"); found {
				timeoutFloat, isFloat := timeoutValObj.(*object.Float)
				timeoutInt, isInt := timeoutValObj.(*object.Integer)
				if isFloat {
					timeout = timeoutFloat.Value
				} else if isInt {
					timeout = float64(timeoutInt.Value)
				} else {
					return object.NewError(constants.TypeError, "timeout must be a float or integer")
				}
			}
		} else if len(args) > 2 {
			// This handles the case of `ping(addr, 1.0)` - positional timeout
			timeoutFloat, isFloat := args[1].(*object.Float)
			timeoutInt, isInt := args[1].(*object.Integer)
			if isFloat {
				timeout = timeoutFloat.Value
			} else if isInt {
				timeout = float64(timeoutInt.Value)
			} else {
				return object.NewError(constants.TypeError, "timeout must be a float or integer")
			}
		}
	}
	// <<< END FIX >>>

	c, err := icmp.ListenPacket("ip4:icmp", "0.0.0.0")
	if err != nil {
		return object.NewError("ICMPError", "failed to listen for ICMP packets: %v", err)
	}
	defer c.Close()

	wm := icmp.Message{
		Type: ipv4.ICMPTypeEcho, Code: 0,
		Body: &icmp.Echo{
			ID: 1234, Seq: 1, Data: []byte("PYLEARN-PING"),
		},
	}
	wb, err := wm.Marshal(nil)
	if err != nil {
		return object.NewError("ICMPError", "failed to marshal ICMP message: %v", err)
	}

	start := time.Now()
	// Resolve the address properly. net.ParseIP is sufficient here.
	ipAddr := net.ParseIP(addrObj.Value)
	if ipAddr == nil {
		return object.NewError(constants.ValueError, "invalid IP address format: %s", addrObj.Value)
	}

	if _, err := c.WriteTo(wb, &net.IPAddr{IP: ipAddr}); err != nil {
		return object.NewError("SendError", "failed to send ICMP packet: %v", err)
	}

	replyBuf := make([]byte, 1500)
	c.SetReadDeadline(time.Now().Add(time.Duration(timeout * float64(time.Second))))
	n, _, err := c.ReadFrom(replyBuf)
	if err != nil {
		if netErr, ok := err.(net.Error); ok && netErr.Timeout() {
			return object.NULL // Return None on timeout
		}
		return object.NewError("ICMPError", "error reading reply: %v", err)
	}
	rtt := time.Since(start)

	rm, err := icmp.ParseMessage(1, replyBuf[:n])
	if err != nil {
		return object.NewError("ICMPError", "error parsing reply: %v", err)
	}

	switch rm.Type {
	case ipv4.ICMPTypeEchoReply:
		return &object.Float{Value: float64(rtt.Microseconds()) / 1000.0} // RTT in milliseconds
	default:
		return object.NewError("ICMPError", "received unexpected ICMP type %v", rm.Type)
	}
}

func createIcmpModule() *object.Module {
	env := object.NewEnvironment()

	// <<< FIX: Register 'timeout' as an accepted keyword argument. >>>
	pingBuiltin := &object.Builtin{
		Name: "net.icmp.ping",
		Fn:   pyIcmpPing,
		AcceptsKeywords: map[string]bool{
			"timeout": true,
		},
	}

	env.Set("ping", pingBuiltin)
	env.Set("ICMPError", ICMPErrorClass)
	return &object.Module{
		Name: "icmp",
		Path: "<builtin_net_icmp>",
		Env:  env,
	}
}
