package uptimegopher

// DO NOT EDIT!

type Severity int

const (
	SeverityDebug Severity = iota
	SeverityNotice
	SeverityWarning
	SeverityError
	SeverityDown
	SeverityFatal
)

type CheckResult struct {
	Success  bool
	Severity Severity
	Message  string
}

type Check struct {
	Key          string
	Name         string
	Run          func(string, map[string]string) CheckResult
	ValidateArgs func(map[string]string) error
}

type PluginCtx struct {}

func (p *PluginCtx) AddCheck(check Check) {}