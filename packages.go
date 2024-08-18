package main

import (
	"context"
	"go/constant"
	"go/token"
	"net"
	"reflect"

	"github.com/traefik/yaegi/interp"
	"golang.org/x/net/idna"
	"golang.org/x/net/proxy"
	"golang.org/x/text/unicode/bidi"
)

var Symbols = interp.Exports{}

func init() {
	Symbols["golang.org/x/text/unicode/bidi/bidi"] = map[string]reflect.Value{
		// function, constant and variable definitions
		"AL":               reflect.ValueOf(bidi.AL),
		"AN":               reflect.ValueOf(bidi.AN),
		"AppendReverse":    reflect.ValueOf(bidi.AppendReverse),
		"B":                reflect.ValueOf(bidi.B),
		"BN":               reflect.ValueOf(bidi.BN),
		"CS":               reflect.ValueOf(bidi.CS),
		"Control":          reflect.ValueOf(bidi.Control),
		"DefaultDirection": reflect.ValueOf(bidi.DefaultDirection),
		"EN":               reflect.ValueOf(bidi.EN),
		"ES":               reflect.ValueOf(bidi.ES),
		"ET":               reflect.ValueOf(bidi.ET),
		"FSI":              reflect.ValueOf(bidi.FSI),
		"L":                reflect.ValueOf(bidi.L),
		"LRE":              reflect.ValueOf(bidi.LRE),
		"LRI":              reflect.ValueOf(bidi.LRI),
		"LRO":              reflect.ValueOf(bidi.LRO),
		"LeftToRight":      reflect.ValueOf(bidi.LeftToRight),
		"Lookup":           reflect.ValueOf(bidi.Lookup),
		"LookupRune":       reflect.ValueOf(bidi.LookupRune),
		"LookupString":     reflect.ValueOf(bidi.LookupString),
		"Mixed":            reflect.ValueOf(bidi.Mixed),
		"NSM":              reflect.ValueOf(bidi.NSM),
		"Neutral":          reflect.ValueOf(bidi.Neutral),
		"ON":               reflect.ValueOf(bidi.ON),
		"PDF":              reflect.ValueOf(bidi.PDF),
		"PDI":              reflect.ValueOf(bidi.PDI),
		"R":                reflect.ValueOf(bidi.R),
		"RLE":              reflect.ValueOf(bidi.RLE),
		"RLI":              reflect.ValueOf(bidi.RLI),
		"RLO":              reflect.ValueOf(bidi.RLO),
		"ReverseString":    reflect.ValueOf(bidi.ReverseString),
		"RightToLeft":      reflect.ValueOf(bidi.RightToLeft),
		"S":                reflect.ValueOf(bidi.S),
		"UnicodeVersion":   reflect.ValueOf(constant.MakeFromLiteral("\"15.0.0\"", token.STRING, 0)),
		"WS":               reflect.ValueOf(bidi.WS),

		// type definitions
		"Class":      reflect.ValueOf((*bidi.Class)(nil)),
		"Direction":  reflect.ValueOf((*bidi.Direction)(nil)),
		"Option":     reflect.ValueOf((*bidi.Option)(nil)),
		"Ordering":   reflect.ValueOf((*bidi.Ordering)(nil)),
		"Paragraph":  reflect.ValueOf((*bidi.Paragraph)(nil)),
		"Properties": reflect.ValueOf((*bidi.Properties)(nil)),
		"Run":        reflect.ValueOf((*bidi.Run)(nil)),
	}

	Symbols["golang.org/x/net/idna/idna"] = map[string]reflect.Value{
		// function, constant and variable definitions
		"BidiRule":                reflect.ValueOf(idna.BidiRule),
		"CheckHyphens":            reflect.ValueOf(idna.CheckHyphens),
		"CheckJoiners":            reflect.ValueOf(idna.CheckJoiners),
		"Display":                 reflect.ValueOf(&idna.Display).Elem(),
		"Lookup":                  reflect.ValueOf(&idna.Lookup).Elem(),
		"MapForLookup":            reflect.ValueOf(idna.MapForLookup),
		"New":                     reflect.ValueOf(idna.New),
		"Punycode":                reflect.ValueOf(&idna.Punycode).Elem(),
		"Registration":            reflect.ValueOf(&idna.Registration).Elem(),
		"RemoveLeadingDots":       reflect.ValueOf(idna.RemoveLeadingDots),
		"StrictDomainName":        reflect.ValueOf(idna.StrictDomainName),
		"ToASCII":                 reflect.ValueOf(idna.ToASCII),
		"ToUnicode":               reflect.ValueOf(idna.ToUnicode),
		"Transitional":            reflect.ValueOf(idna.Transitional),
		"UnicodeVersion":          reflect.ValueOf(constant.MakeFromLiteral("\"15.0.0\"", token.STRING, 0)),
		"ValidateForRegistration": reflect.ValueOf(idna.ValidateForRegistration),
		"ValidateLabels":          reflect.ValueOf(idna.ValidateLabels),
		"VerifyDNSLength":         reflect.ValueOf(idna.VerifyDNSLength),

		// type definitions
		"Option":  reflect.ValueOf((*idna.Option)(nil)),
		"Profile": reflect.ValueOf((*idna.Profile)(nil)),
	}

	Symbols["golang.org/x/net/proxy/proxy"] = map[string]reflect.Value{
		// function, constant and variable definitions
		"Dial":                 reflect.ValueOf(proxy.Dial),
		"Direct":               reflect.ValueOf(&proxy.Direct).Elem(),
		"FromEnvironment":      reflect.ValueOf(proxy.FromEnvironment),
		"FromEnvironmentUsing": reflect.ValueOf(proxy.FromEnvironmentUsing),
		"FromURL":              reflect.ValueOf(proxy.FromURL),
		"NewPerHost":           reflect.ValueOf(proxy.NewPerHost),
		"RegisterDialerType":   reflect.ValueOf(proxy.RegisterDialerType),
		"SOCKS5":               reflect.ValueOf(proxy.SOCKS5),

		// type definitions
		"Auth":          reflect.ValueOf((*proxy.Auth)(nil)),
		"ContextDialer": reflect.ValueOf((*proxy.ContextDialer)(nil)),
		"Dialer":        reflect.ValueOf((*proxy.Dialer)(nil)),
		"PerHost":       reflect.ValueOf((*proxy.PerHost)(nil)),

		// interface wrapper definitions
		"_ContextDialer": reflect.ValueOf((*_golang_org_x_net_proxy_ContextDialer)(nil)),
		"_Dialer":        reflect.ValueOf((*_golang_org_x_net_proxy_Dialer)(nil)),
	}
}

// _golang_org_x_net_proxy_ContextDialer is an interface wrapper for ContextDialer type
type _golang_org_x_net_proxy_ContextDialer struct {
	IValue       interface{}
	WDialContext func(ctx context.Context, network string, address string) (net.Conn, error)
}

func (W _golang_org_x_net_proxy_ContextDialer) DialContext(ctx context.Context, network string, address string) (net.Conn, error) {
	return W.WDialContext(ctx, network, address)
}

// _golang_org_x_net_proxy_Dialer is an interface wrapper for Dialer type
type _golang_org_x_net_proxy_Dialer struct {
	IValue interface{}
	WDial  func(network string, addr string) (c net.Conn, err error)
}

func (W _golang_org_x_net_proxy_Dialer) Dial(network string, addr string) (c net.Conn, err error) {
	return W.WDial(network, addr)
}
