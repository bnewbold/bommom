package main

type Offer struct {
}

type LineItem struct {
}

type Element struct {
}

// The main anchor of a BOM as a cohesive whole, with a name and permissions.
// Multiple BOMs are associated with a single BomStub; the currently active one
// is the 'head'.
type BomStub struct {
	name                       *ShortName
	owner                      string
	description                string
	homepage                   *Url
	isPublicView, isPublicEdit bool
}

// An actual list of parts/elements. Intended to be immutable once persisted. 
type Bom struct {
	version   *ShortName
	date      uint64 // TODO: unix timestamp?
	progeny   string // where did this BOM come from?
	elements  []Element
	lineitems []LineItem
}
