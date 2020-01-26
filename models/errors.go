package models

import "strings"

const (
	// ErrNotFound is returned when a resource is not found in the database.
	ErrNotFound modelError = "models: resource not found"

	// ErrPasswordIncorrect is returned when a password match is not found in the database.
	ErrPasswordIncorrect modelError = "models: incorrect password provided"

	// ErrEmailRequired is returned when an email address is not provided on user creation.
	ErrEmailRequired modelError = "models: email address is required"

	// ErrEmailInvalid is returned when an email does not match a valid email format.
	ErrEmailInvalid modelError = "models: email address is not valid"

	// ErrEmailTaken is returned when an email already exists in the database.
	ErrEmailTaken modelError = "models: email address is already taken"

	// ErrPasswordTooShort is returned when a password is less than 8 characters.
	ErrPasswordTooShort modelError = "models: password must be at least 8 characters"

	// ErrRememberRequired is returned when a remember token is not provided.
	ErrRememberRequired modelError = "models: remember token is required"

	// ErrTitleRequired is returned when a user does not provide a gallery title.
	ErrTitleRequired modelError = "models: title is required"

	// ErrTokenInvalid is returned when a provided token does not exist.
	ErrTokenInvalid modelError = "models: token provided is not valid"

	// ErrIDInvalid is returned when an invalid ID is provided.
	ErrIDInvalid privateError = "models: ID provided was invalid"

	// ErrPasswordRequired is returned when a password is not provided.
	ErrPasswordRequired privateError = "models: password is required"

	// ErrRememberTooShort is returned when a remember token is less than 32 bytes.
	ErrRememberTooShort privateError = "models: remember token must be at least 32 bytes"

	// ErrUserIDRequired is returned when a user ID is not provided.
	ErrUserIDRequired privateError = "models: user ID is required"

	// ErrServiceRequired is returned when a service is not provided.
	ErrServiceRequired privateError = "models: service is required"
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

type privateError string

func (e privateError) Error() string {
	return string(e)
}
