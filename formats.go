package main

// Bom/BomMeta conversion/dump/load routines

import (
	"encoding/csv"
	"encoding/json"
	"encoding/xml"
	"fmt"
	"io"
	"log"
	"strings"
)

// This compound container struct is useful for serializing to XML and JSON
type BomContainer struct {
	BomMetadata *BomMeta `json:"metadata"`
	Bom         *Bom     `json:"bom"`
}

// --------------------- text (CLI only ) -----------------------

func DumpBomAsText(bm *BomMeta, b *Bom, out io.Writer) {
	fmt.Fprintln(out)
	fmt.Fprintf(out, "Name:\t\t%s\n", bm.Name)
	fmt.Fprintf(out, "Version:\t%s\n", b.Version)
	fmt.Fprintf(out, "Creator:\t%s\n", bm.Owner)
	fmt.Fprintf(out, "Timestamp:\t%s\n", b.Created)
	if bm.Homepage != "" {
		fmt.Fprintf(out, "Homepage:\t%s\n", bm.Homepage)
	}
	if b.Progeny != "" {
		fmt.Fprintf(out, "Source:\t\t%s\n", b.Progeny)
	}
	if bm.Description != "" {
		fmt.Fprintf(out, "Description:\t%s\n", bm.Description)
	}
	fmt.Println()
	// "by line item"
	fmt.Fprintf(out, "tag\tqty\tmanufacturer\tmpn\t\tdescription\t\tcomment\n")
	for _, li := range b.LineItems {
		fmt.Fprintf(out, "%s\t%d\t%s\t%s\t\t%s\t\t%s\n",
			li.Tag,
			len(li.Elements),
			li.Manufacturer,
			li.Mpn,
			li.Description,
			li.Comment)
	}
	/* // "by circuit element"
	   fmt.Fprintf(out, "tag\tsymbol\tmanufacturer\tmpn\t\tdescription\t\tcomment\n")
	   for _, li := range b.LineItems {
	       for _, elm := range li.Elements {
	           fmt.Fprintf(out, "%s\t%s\t%s\t%s\t\t%s\t\t%s\n",
	                      li.Tag,
	                      elm,
	                      li.Manufacturer,
	                      li.Mpn,
	                      li.Description,
	                      li.Comment)
	       }
	   }
	*/
}

// --------------------- csv -----------------------

func DumpBomAsCSV(b *Bom, out io.Writer) {
	dumper := csv.NewWriter(out)
	defer dumper.Flush()
	// "by line item"
	dumper.Write([]string{"qty",
		"symbols",
		"manufacturer",
		"mpn",
		"description",
		"comment"})
	for _, li := range b.LineItems {
		dumper.Write([]string{
			fmt.Sprint(len(li.Elements)),
			strings.Join(li.Elements, ","),
			li.Manufacturer,
			li.Mpn,
			li.Description,
			li.Comment})
	}
}

func LoadBomFromCSV(out io.Writer) (*Bom, error) {

	b := Bom{}

	return &b, nil
}

// --------------------- JSON -----------------------

func DumpBomAsJSON(bm *BomMeta, b *Bom, out io.Writer) {

	container := &BomContainer{BomMetadata: bm, Bom: b}

	enc := json.NewEncoder(out)
	if err := enc.Encode(&container); err != nil {
		log.Fatal(err)
	}
}

func LoadBomFromJSON(input io.Reader) (*BomMeta, *Bom, error) {

	container := &BomContainer{}

	enc := json.NewDecoder(input)
	if err := enc.Decode(&container); err != nil {
		log.Fatal(err)
	}
	return container.BomMetadata, container.Bom, nil
}

// --------------------- XML -----------------------

func DumpBomAsXML(bm *BomMeta, b *Bom, out io.Writer) {

	container := &BomContainer{BomMetadata: bm, Bom: b}
	enc := xml.NewEncoder(out)

	// generic XML header
	io.WriteString(out, xml.Header)

	if err := enc.Encode(container); err != nil {
		log.Fatal(err)
	}
}

func LoadBomFromXML(input io.Reader) (*BomMeta, *Bom, error) {

	container := &BomContainer{}
	enc := xml.NewDecoder(input)
	if err := enc.Decode(&container); err != nil {
		log.Fatal(err)
	}
	return container.BomMetadata, container.Bom, nil
}
