package views

// Alert variables that are used for error handling
// or giving users visual cues.
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

func (d *Data) SetAlert(err error) {
	if pubErr, ok := err.(PublicError); ok {
		d.Alert = &Alert{
			Level:   AlertLevelError,
			Message: pubErr.Public(),
		}
	} else {
		d.Alert = &Alert{
			Level:   AlertLevelError,
			Message: AlertMsgGeneric,
		}
	}
}

type PublicError interface {
	error
	Public() string
}
