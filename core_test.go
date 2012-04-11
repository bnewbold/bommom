package main

import (
	"encoding/json"
	//"fmt"
	"os"
	"testing"
)

func makeTestBom() *Bom {
	op1 := OfferPrice{Currency: "usd", Price: 1.0, MinQty: 1}
	op2 := OfferPrice{Currency: "usd", Price: 0.8, MinQty: 100}
	o := Offer{Sku: "A123", Distributor: "Acme", Prices: []OfferPrice{op1, op2}}
	//o.AddOfferPrice(op1)
	//o.AddOfferPrice(op2)
	li := LineItem{Mfg: "WidgetCo",
		Mpn:      "WIDG0001",
		Elements: []string{"W1", "W2"},
		Offers:   []Offer{o}}
	//li.AddOffer(o)
	b := NewBom("test01")
	b.AddLineItem(&li)
	return b
}

func TestNewBom(t *testing.T) {
	b := makeTestBom()
	if b == nil {
		t.Errorf("Something went wrong")
	}
}

func TestBomJSONDump(t *testing.T) {

	b := makeTestBom()
	enc := json.NewEncoder(os.Stdout)

	if err := enc.Encode(b); err != nil {
		t.Errorf("Error encoding: " + err.Error())
	}
}
