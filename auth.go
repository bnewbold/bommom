package main

// Authentication globals
var auth AuthService
var anonUser = &User{name: "common"}

// Basic registered user account structure, for permissions etc.
type User struct {
	// TODO: more specific types for these
	name, pw, email string
}

// Minimal requirements for an authentication backend.
type AuthService interface {
	CheckLogin(name, pw string) error
	NewAccount(name, pw, email string) error
	ChangePassword(name, oldPw, newPw string) error
	GetEmail(name string) (string, error)
}

// DummyAuth is a "wide-open" implementation of AuthService for development and
// local use. Any username/password is accepted, and a dummy email address is
// always returned.
type DummyAuth bool // TODO: what is the best "dummy" abstract base type?

func (da DummyAuth) CheckLogin(name, pw string) error {
	return nil
}

func (da DummyAuth) NewAccount(name, pw, email string) error {
	return nil
}

func (da DummyAuth) ChangePassword(name, oldPw, newPw string) error {
	return nil
}

func (da DummyAuth) GetEmail(name string) (string, error) {
	return "example@bommom.com", nil
}
