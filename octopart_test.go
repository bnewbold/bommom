package main

import (
	"log"
	"os"
	"testing"
)

func TestApiCall(t *testing.T) {
	oc := NewOctopartClient("")
	result, err := oc.apiCall("parts/search", map[string]string{"q": "ne555", "limit": "2"})
	if err != nil {
		t.Errorf("Error with api call: " + err.Error())
	}
	log.Println(result["hits"])
}

func TestGetMarketInfoList(t *testing.T) {
	oc := NewOctopartClient("")
	log.Println("Running the first time...")
	result, err := oc.GetMarketInfoList([]string{"ti", "atmel", "atmel"}, []string{"ne555", "attiny*", "avrtiny123qqq?"})
	if err != nil {
		t.Errorf("Error with api call: " + err.Error())
	}
	for i, r := range result {
		if r == nil {
			log.Printf("\t%d: %s", i, "nil")
		} else {
			log.Printf("\t%d: %s", i, r.(map[string]interface{})["detail_url"])
		}
	}
	log.Println("Running a second time, results should be cached...")
	result, err = oc.GetMarketInfoList([]string{"ti", "atmel", "atmel"}, []string{"ne555", "attiny*", "avrtiny123qqq?"})
	if err != nil {
		t.Errorf("Error with api call: " + err.Error())
	}
	for i, r := range result {
		if r == nil {
			log.Printf("\t%d: %s", i, "nil")
		} else {
			log.Printf("\t%d: %s", i, r.(map[string]interface{})["detail_url"])
		}
	}
	log.Println("Running in single mode, result should be cached...")
	result_single, err := oc.GetMarketInfo("ti", "ne555")
	if err != nil {
		t.Errorf("Error with api call: " + err.Error())
	}
    if result_single == nil {
        log.Printf("\t%d: %s", "nil")
    } else {
        log.Printf("\t%d: %s", result_single["detail_url"])
    }
}

func TestAttachInfo(t *testing.T) {
	_, b := makeTestBom()
	bm := &BomMeta{}
	oc := NewOctopartClient("")
	oc.AttachMarketInfoBom(b)
	t.Errorf("unimplemented")
	DumpBomAsText(bm, b, os.Stdout)
}
