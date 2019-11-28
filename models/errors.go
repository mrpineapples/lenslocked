package models

import "strings"

const (
	// ErrNotFound is returned when a resource is not found in the database.
	ErrNotFound modelError = "models: resource not found"

	// ErrIDInvalid is returned when an invalid ID is provided.
	ErrIDInvalid modelError = "models: ID provided was invalid"

	// ErrPasswordIncorrect is returned when a password match is not found in the database.
	ErrPasswordIncorrect modelError = "models: incorrect password provided"

	// ErrEmailRequired is returned when an email address is not provided on user creation.
	ErrEmailRequired modelError = "models: email address is required"

	// ErrEmailInvalid is returned when an email does not match a valid email format.
	ErrEmailInvalid modelError = "models: email address is not valid"

	// ErrEmailTaken is returned when an email already exists in the database.
	ErrEmailTaken modelError = "models: email address is already taken"

	// ErrPasswordRequired is returned when a password is not provided.
	ErrPasswordRequired modelError = "models: password is required"

	// ErrPasswordTooShort is returned when a password is less than 8 characters.
	ErrPasswordTooShort modelError = "models: password must be at least 8 characters"

	// ErrRememberTooShort is returned when a remember token is less than 32 bytes.
	ErrRememberTooShort modelError = "models: remember token must be at least 32 bytes"

	// ErrRememberRequired is returned when a remember token is not provided.
	ErrRememberRequired modelError = "models: remember token is required"
)

type modelError string

func (e modelError) Error() string {
	return string(e)
}

func (e modelError) Public() string {
	s := strings.Replace(string(e), "models: ", "", 1)
	split := strings.Split(s, " ")
	split[0] = strings.Title(split[0])
	return strings.Join(split, " ") + "."
}
