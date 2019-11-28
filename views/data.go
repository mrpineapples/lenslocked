package views

const (
	AlertLevelError   = "danger"
	AlertLevelWarning = "warning"
	AlertLevelInfo    = "info"
	AlertLevelSuccess = "success"

	// AlertMsgGeneric is displayed when occur backend encounters an unexpected error.
	AlertMsgGeneric = "Something went wrong. Please try again, and contact us if the problem persists"
)

// Alert is used to render bootstrap alerts in templates
type Alert struct {
	Level   string
	Message string
}

// Data is the top level structure that views expect data to come in.
type Data struct {
	Alert *Alert
	Yield interface{}
}
