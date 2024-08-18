package checks

import (
	"fmt"
	"net/url"
	"strings"
	"time"
	uptimegopher "uptime-gopher/uptime-gopher"

	"github.com/likexian/whois"
	whoisparser "github.com/likexian/whois-parser"
)

type CheckDomain struct {
}

// TODO: Add support for RDAP
// RDAP: https://data.iana.org/rdap/dns.json and add /domain/github.com to the end
// func (c *CheckDomain) Check(address string, args map[string]string) uptimegopher.CheckResult {
// 	return uptimegopher.CheckResult{
// 		Success: true,
// 	}
// }

// Whois Check
func (c *CheckDomain) Check(address string, args map[string]string) uptimegopher.CheckResult {
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

	raw, err := whois.Whois(url.Hostname())
	if err != nil {
		return uptimegopher.CheckResult{
			Success:  false,
			Severity: uptimegopher.SeverityDown,
			Message:  err.Error(),
		}
	}

	result, err := whoisparser.Parse(raw)
	if err != nil {
		return uptimegopher.CheckResult{
			Success:  false,
			Severity: uptimegopher.SeverityDown,
			Message:  err.Error(),
		}
	}

	now := time.Now()
	validBefore, err := time.Parse(time.RFC3339, result.Domain.ExpirationDate)
	if err != nil {
		return uptimegopher.CheckResult{
			Success:  false,
			Severity: uptimegopher.SeverityError,
			Message:  fmt.Sprintf("Failed to parse expiration date: %s", err),
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

func DomainCheck() uptimegopher.Check {
	checker := &CheckDomain{}

	return uptimegopher.Check{
		Key:  "dns",
		Name: "Domain Check",
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
