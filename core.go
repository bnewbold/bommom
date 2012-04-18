package main

import (
	"time"
)

type OfferPrice struct {
	Currency string
	MinQty   uint32
	Price    float32
}

type Offer struct {
	Distributor, Sku, Url, Comment string
	Prices                         []OfferPrice
}

type LineItem struct {
	Mfg, Mpn, Description, Comment, Tag string
	Elements                            []string // TODO: add "circuit element" type
	Offers                              []Offer
}

func (li *LineItem) Id() string {
	return li.Mfg + "::" + li.Mpn
}

// The main anchor of a BOM as a cohesive whole, with a name and permissions.
// Multiple BOMs are associated with a single BomStub; the currently active one
// is the 'head'.
type BomStub struct {
	Name                       string
	Owner                      string
	Description                string
	HeadVersion                string
	Homepage                   *Url
	IsPublicView, IsPublicEdit bool
}

// An actual list of parts/elements. Intended to be immutable once persisted. 
type Bom struct {
	Version   string
	Created   time.Time // TODO: unix timestamp?
	Progeny   string    // where did this BOM come from?
	LineItems []LineItem
}

func NewBom(version string) *Bom {
	return &Bom{Version: version, Created: time.Now()}
}

func (b *Bom) GetLineItem(mfg, mpn string) *LineItem {
	for _, li := range b.LineItems {
		if li.Mfg == mfg && li.Mpn == mpn {
			return &li
		}
	}
	return nil
}

func (b *Bom) AddLineItem(li *LineItem) error {
	if eli := b.GetLineItem(li.Mfg, li.Mpn); eli != nil {
		return Error("This BOM already had an identical LineItem")
	}
	b.LineItems = append(b.LineItems, *li)
	return nil
}

// ---------- testing
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
