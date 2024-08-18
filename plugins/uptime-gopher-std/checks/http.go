package checks

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"sync/atomic"
	"time"

	uptimegopher "uptime-gopher/uptime-gopher"
)

type CheckHttp struct {
	failed atomic.Uint32
}

func (hc *CheckHttp) Check(address string, args map[string]string) uptimegopher.CheckResult {
	method, ok := args["method"]
	if !ok || method == "" {
		method = "GET"
	}

	successCode, ok := args["success_code"]
	if !ok || successCode == "" {
		successCode = "200"
	}

	retriesRaw, ok := args["retries"]
	if !ok || retriesRaw == "" {
		retriesRaw = "3"
	}

	retries, err := strconv.Atoi(retriesRaw)
	if err != nil {
		hc.failed.Add(1)

		return uptimegopher.CheckResult{
			Success:  false,
			Severity: uptimegopher.SeverityFatal,
			Message:  err.Error(),
		}
	}

	timeoutRaw, ok := args["timeout"]
	if !ok || timeoutRaw == "" {
		timeoutRaw = "5s"
	}

	timeout, err := time.ParseDuration(timeoutRaw)
	if err != nil {
		hc.failed.Add(1)

		return uptimegopher.CheckResult{
			Success:  false,
			Severity: uptimegopher.SeverityFatal,
			Message:  err.Error(),
		}
	}

	url, err := url.Parse(address)
	if err != nil {
		hc.failed.Add(1)

		return uptimegopher.CheckResult{
			Success:  false,
			Severity: uptimegopher.SeverityFatal,
			Message:  err.Error(),
		}
	}

	if url.Scheme == "" {
		url.Scheme = "https"
	}

	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, method, url.String(), nil)
	if err != nil {
		hc.failed.Add(1)

		return uptimegopher.CheckResult{
			Success:  false,
			Severity: uptimegopher.SeverityFatal,
			Message:  err.Error(),
		}
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		if hc.failed.Load() < uint32(retries) {
			hc.failed.Add(1)

			return uptimegopher.CheckResult{
				Success:  false,
				Severity: uptimegopher.SeverityDebug,
				Message:  err.Error(),
			}
		}

		hc.failed.Add(1)

		return uptimegopher.CheckResult{
			Success:  false,
			Severity: uptimegopher.SeverityError,
			Message:  err.Error(),
		}
	}

	defer resp.Body.Close()
	defer io.Copy(io.Discard, resp.Body)

	if strconv.Itoa(resp.StatusCode) != successCode {
		if hc.failed.Load() < uint32(retries) {
			hc.failed.Add(1)

			return uptimegopher.CheckResult{
				Success:  false,
				Severity: uptimegopher.SeverityDebug,
				Message:  fmt.Sprintf("status code is not as expected: %d", resp.StatusCode),
			}
		}

		hc.failed.Add(1)

		return uptimegopher.CheckResult{
			Success:  false,
			Severity: uptimegopher.SeverityError,
			Message:  fmt.Sprintf("status code is not as expected: %d", resp.StatusCode),
		}
	}

	hc.failed.Store(0)

	return uptimegopher.CheckResult{
		Success: true,
	}
}

func HttpCheck() uptimegopher.Check {
	checker := &CheckHttp{}

	return uptimegopher.Check{
		Key:  "http",
		Name: "Http Check",
		Run:  checker.Check,
		ValidateArgs: func(args map[string]string) error {
			method, ok := args["method"]
			if ok {
				if method != "GET" && method != "POST" {
					return fmt.Errorf("method must be GET or POST")
				}
			}

			successCode, ok := args["success_code"]
			if ok {
				num, err := strconv.Atoi(successCode)
				if err != nil {
					return fmt.Errorf("success_code must be a number")
				}

				if num < 100 || num > 599 {
					return fmt.Errorf("success_code must be between 100 and 599")
				}
			}

			retries, ok := args["retries"]
			if ok {
				num, err := strconv.Atoi(retries)
				if err != nil {
					return fmt.Errorf("retries must be a number")
				}

				if num < 1 || num > 10 {
					return fmt.Errorf("retries must be between 1 and 10")
				}
			}

			return nil
		},
	}
}
