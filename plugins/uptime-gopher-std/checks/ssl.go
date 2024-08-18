package checks

import (
	"crypto/tls"
	"fmt"
	"net"
	"net/url"
	"strings"
	"time"
	uptimegopher "uptime-gopher/uptime-gopher"
)

type CheckSsl struct{}

func (c *CheckSsl) Check(address string, args map[string]string) uptimegopher.CheckResult {
	notifyAfterRaw, ok := args["notify_after"]
	if !ok || notifyAfterRaw == "" {
		notifyAfterRaw = "720h"
	}

	notifyAfter, err := time.ParseDuration(notifyAfterRaw)
	if err != nil {
		return uptimegopher.CheckResult{
			Success:  false,
			Severity: uptimegopher.SeverityFatal,
			Message:  err.Error(),
		}
	}

	errorAfterRaw, ok := args["error_after"]
	if !ok || errorAfterRaw == "" {
		errorAfterRaw = "168h"
	}

	errorAfter, err := time.ParseDuration(errorAfterRaw)
	if err != nil {
		return uptimegopher.CheckResult{
			Success:  false,
			Severity: uptimegopher.SeverityFatal,
			Message:  err.Error(),
		}
	}

	if !strings.HasPrefix(address, "https://") && !strings.HasPrefix(address, "http://") {
		address = "https://" + address
	}

	url, err := url.Parse(address)
	if err != nil {
		return uptimegopher.CheckResult{
			Success:  false,
			Severity: uptimegopher.SeverityFatal,
			Message:  err.Error(),
		}
	}

	addr := net.JoinHostPort(url.Hostname(), "https")

	conn, err := net.Dial("tcp", addr)
	if err != nil {
		return uptimegopher.CheckResult{
			Success:  false,
			Severity: uptimegopher.SeverityDown,
			Message:  err.Error(),
		}
	}
	defer conn.Close()

	tlsConn := tls.Client(conn, &tls.Config{
		ServerName: url.Hostname(),
	})

	err = tlsConn.Handshake()
	if err != nil {
		return uptimegopher.CheckResult{
			Success:  false,
			Severity: uptimegopher.SeverityDown,
			Message:  err.Error(),
		}
	}

	state := tlsConn.ConnectionState()

	if len(state.PeerCertificates) == 0 {
		return uptimegopher.CheckResult{
			Success:  false,
			Severity: uptimegopher.SeverityDown,
			Message:  "No certificates found",
		}
	}

	now := time.Now()
	validBefore := state.PeerCertificates[0].NotAfter
	if len(state.PeerCertificates) > 1 {
		for _, cert := range state.PeerCertificates[1:] {
			if cert.NotAfter.Before(validBefore) {
				validBefore = cert.NotAfter
			}
		}
	}

	if validBefore.Before(now) {
		return uptimegopher.CheckResult{
			Success:  false,
			Severity: uptimegopher.SeverityDown,
			Message:  fmt.Sprintf("Certificate is not valid anymore. Expiration date: %s", validBefore.Format(time.RFC1123)),
		}
	}

	if validBefore.Before(now.Add(notifyAfter)) {
		return uptimegopher.CheckResult{
			Success:  false,
			Severity: uptimegopher.SeverityWarning,
			Message:  fmt.Sprintf("Certificate is about to expire. Expiration date: %s", validBefore.Format(time.RFC1123)),
		}
	}

	if validBefore.Before(now.Add(errorAfter)) {
		return uptimegopher.CheckResult{
			Success:  false,
			Severity: uptimegopher.SeverityError,
			Message:  fmt.Sprintf("Certificate is about to expire. Expiration date: %s", validBefore.Format(time.RFC1123)),
		}
	}

	return uptimegopher.CheckResult{
		Success: true,
	}
}

func SslCheck() uptimegopher.Check {
	checker := &CheckSsl{}

	return uptimegopher.Check{
		Key:  "ssl",
		Name: "Ssl Check",
		Run:  checker.Check,
		ValidateArgs: func(args map[string]string) error {
			notifyAfter, ok := args["notify_after"]
			if ok {
				_, err := time.ParseDuration(notifyAfter)
				if err != nil {
					return fmt.Errorf("notify_after must be a duration")
				}
			}

			errorAfter, ok := args["error_after"]
			if ok {
				_, err := time.ParseDuration(errorAfter)
				if err != nil {
					return fmt.Errorf("error_after must be a duration")
				}
			}

			return nil
		},
	}
}
