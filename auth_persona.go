package main

import (
	"encoding/json"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
)

type PersonaResponse struct {
	Status, Email, Reason string
}

func (b PersonaResponse) Okay() bool {
	return b.Status == "okay"
}

func VerifyPersonaAssertion(assertion, audience string) PersonaResponse {
	resp, _ := http.PostForm(
		"https://browserid.org/verify",
		url.Values{
			"assertion": {assertion},
			"audience":  {audience},
		})
	response := personaResponseFromJson(resp.Body)
	resp.Body.Close()

	return response
}

func personaResponseFromJson(r io.Reader) (resp PersonaResponse) {
	body, err := ioutil.ReadAll(r)

	if err != nil {
		log.Fatal(err)
	}

	err = json.Unmarshal(body, &resp)

	if err != nil {
		log.Fatal(err)
	}

	return resp
}

type PersonaAuth bool

func (pa PersonaAuth) CheckLogin(name, pw string) error {
	return nil
}

func (pa PersonaAuth) NewAccount(name, pw, email string) error {
	return nil
}

func (pa PersonaAuth) ChangePassword(name, oldPw, newPw string) error {
	return nil
}

func (pa PersonaAuth) GetEmail(name string) (string, error) {
	return "example@localhost", nil
}

func (pa PersonaAuth) GetUserName(name string) (string, error) {
	return "common", nil
}
