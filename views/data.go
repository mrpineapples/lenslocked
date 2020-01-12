package views

import (
	"log"
	"net/http"
	"time"

	"github.com/mrpineapples/lenslocked/models"
)

// Alert variables that are used for error handling
// or giving users visual cues.
const (
	AlertLevelError   = "danger"
	AlertLevelWarning = "warning"
	AlertLevelInfo    = "info"
	AlertLevelSuccess = "success"

	// AlertMsgGeneric is displayed when occur backend encounters an unexpected error.
	AlertMsgGeneric = "Something went wrong. Please try again, and contact us if the problem persists"

	// Cookie names that allow persistent alerts
	AlertLevelCookie   = "alert_level"
	AlertMessageCookie = "alert_message"
)

// Alert is used to render bootstrap alerts in templates
type Alert struct {
	Level   string
	Message string
}

// Data is the top level structure that views expect data to come in.
type Data struct {
	Alert *Alert
	User  *models.User
	Yield interface{}
}

func (d *Data) SetAlert(err error) {
	if pubErr, ok := err.(PublicError); ok {
		d.Alert = &Alert{
			Level:   AlertLevelError,
			Message: pubErr.Public(),
		}
	} else {
		log.Println(err)
		d.Alert = &Alert{
			Level:   AlertLevelError,
			Message: AlertMsgGeneric,
		}
	}
}

func (d *Data) AlertError(msg string) {
	d.Alert = &Alert{
		Level:   AlertLevelError,
		Message: msg,
	}
}

type PublicError interface {
	error
	Public() string
}

func persistAlert(w http.ResponseWriter, alert Alert) {
	expiresAt := time.Now().Add(2 * time.Minute)
	lvl := http.Cookie{
		Name:     AlertLevelCookie,
		Value:    alert.Level,
		Expires:  expiresAt,
		HttpOnly: true,
	}
	msg := http.Cookie{
		Name:     AlertMessageCookie,
		Value:    alert.Message,
		Expires:  expiresAt,
		HttpOnly: true,
	}

	http.SetCookie(w, &lvl)
	http.SetCookie(w, &msg)
}

func clearAlert(w http.ResponseWriter) {
	lvl := http.Cookie{
		Name:     AlertLevelCookie,
		Value:    "",
		Expires:  time.Now(),
		HttpOnly: true,
	}
	msg := http.Cookie{
		Name:     AlertMessageCookie,
		Value:    "",
		Expires:  time.Now(),
		HttpOnly: true,
	}

	http.SetCookie(w, &lvl)
	http.SetCookie(w, &msg)
}

func getAlert(r *http.Request) *Alert {
	lvl, err := r.Cookie(AlertLevelCookie)
	if err != nil {
		return nil
	}

	msg, err := r.Cookie(AlertMessageCookie)
	if err != nil {
		return nil
	}

	alert := Alert{
		Level:   lvl.Value,
		Message: msg.Value,
	}
	return &alert
}

func RedirectWithAlert(w http.ResponseWriter, r *http.Request, urlStr string, code int, alert Alert) {
	persistAlert(w, alert)
	http.Redirect(w, r, urlStr, code)
}
