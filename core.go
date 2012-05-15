package main

import (
	"time"
)

type OfferPrice struct {
	Currency string  `json:"currency"`
	MinQty   uint32  `json:"min_qty"`
	Price    float32 `json:"price"`
}

type Offer struct {
	Distributor string       `json:"distributor_name"`
	Sku         string       `json:"sku"`
	Url         string       `json:"distributor_url"`
	Comment     string       `json:"comment"`
    Available       uint32   `json:"avail"`
	Prices      []OfferPrice `json:"prices"`
}

type LineItem struct {
	Manufacturer string   `json:"manufacturer"`
	Mpn          string   `json:"mpn"`
	Description  string   `json:"description"`
	FormFactor   string   `json:"form_factor"` // type:string
	Specs        string   `json:"specs"`       // comma seperated list
	Comment      string   `json:"comment"`
	Tag          string   `json:"tag"`      // comma seperated list
	Category     string   `json:"category"` // hierarchy as comma seperated list
	Elements     []string `json:"elements"`
	Offers       []Offer  `json:"offers"`
}

func (li *LineItem) Id() string {
	return li.Manufacturer + "::" + li.Mpn
}

// The main anchor of a BOM as a cohesive whole, with a name and permissions.
// Multiple BOMs are associated with a single BomMeta; the currently active one
// is the 'head'.
type BomMeta struct {
	Name         string `json:"name"`
	Owner        string `json:"owner_name"`
	Description  string `json:"description"`
	HeadVersion  string `json:"head_version"`
	Homepage     Url    `json:"homepage_url"`
	IsPublicView bool   `json:"is_publicview",omitempty`
	IsPublicEdit bool   `json:"is_publicedit",omitempty`
}

// An actual list of parts/elements. Intended to be immutable once persisted. 
type Bom struct {
	Version string `json:"version"`
	// TODO: unix timestamp?
	Created time.Time `json:"created_ts"`
	// "where did this BOM come from?"
	Progeny   string     `json:"progeny",omitifempty`
	LineItems []LineItem `json:"line_items"`
}

func NewBom(version string) *Bom {
	return &Bom{Version: version, Created: time.Now()}
}

func (b *Bom) GetLineItem(mfg, mpn string) *LineItem {
	for _, li := range b.LineItems {
		if li.Manufacturer == mfg && li.Mpn == mpn {
			return &li
		}
	}
	return nil
}

func (b *Bom) AddLineItem(li *LineItem) error {
	if eli := b.GetLineItem(li.Manufacturer, li.Mpn); eli != nil {
		return Error("This BOM already had an identical LineItem")
	}
	b.LineItems = append(b.LineItems, *li)
	return nil
}

func (b *Bom) Validate() error {
	if !isShortName(b.Version) {
		return Error("version not a ShortName: \"" + b.Version + "\"")
	}
	if b.Created.IsZero() {
		return Error("created timestamp not defined")
	}
	return nil
}

func (bm *BomMeta) Validate() error {
	if !isShortName(bm.Name) {
		return Error("name not a ShortName: \"" + bm.Name + "\"")
	}
	if !isShortName(bm.Owner) {
		return Error("owner name not a ShortName: \"" + bm.Owner + "\"")
	}
	return nil
}

// ---------- testing
func makeTestBom() *Bom {
	op1 := OfferPrice{Currency: "usd", Price: 1.0, MinQty: 1}
	op2 := OfferPrice{Currency: "usd", Price: 0.8, MinQty: 100}
	o := Offer{Sku: "A123", Distributor: "Acme", Prices: []OfferPrice{op1, op2}}
	//o.AddOfferPrice(op1)
	//o.AddOfferPrice(op2)
	li := LineItem{Manufacturer: "WidgetCo",
		Mpn:      "WIDG0001",
		Elements: []string{"W1", "W2"},
		Offers:   []Offer{o}}
	//li.AddOffer(o)
	b := NewBom("test01")
	b.AddLineItem(&li)
	return b
}
