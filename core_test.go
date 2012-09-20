package main

import (
	"encoding/json"
	//"fmt"
	"os"
	"testing"
)

func TestNewBom(t *testing.T) {
	b, _ := makeTestBom()
	if b == nil {
		t.Errorf("Something went wrong")
	}
}

func TestBomJSONDump(t *testing.T) {

	b, _ := makeTestBom()
	enc := json.NewEncoder(os.Stdout)

	if err := enc.Encode(b); err != nil {
		t.Errorf("Error encoding: " + err.Error())
	}
}
