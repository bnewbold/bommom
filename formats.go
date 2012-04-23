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
    "strconv"
    "text/tabwriter"
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
    tabWriter := tabwriter.NewWriter(out, 2, 4, 1, ' ', 0)
	// "by line item"
    // TODO: use text/tabwriter here to get proper column alignment
	fmt.Fprintf(tabWriter, "tag\tqty\tmanufacturer\tmpn\t\tfunction\t\tcomment\n")
	for _, li := range b.LineItems {
		fmt.Fprintf(tabWriter, "%s\t%d\t%s\t%s\t\t%s\t\t%s\n",
			li.Tag,
			len(li.Elements),
			li.Manufacturer,
			li.Mpn,
			li.Function,
			li.Comment)
	}
    tabWriter.Flush()
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
		"function",
		"form_factor",
		"specs",
		"comment"})
	for _, li := range b.LineItems {
		dumper.Write([]string{
			fmt.Sprint(len(li.Elements)),
			strings.Join(li.Elements, ","),
			li.Manufacturer,
			li.Mpn,
			li.Function,
			li.FormFactor,
			li.Specs,
			li.Comment})
	}
}

func appendField(existing, next *string) {
    if *existing == "" {
        *existing += " " + strings.TrimSpace(*next)
    }
    *existing = strings.TrimSpace(*next)
}

func LoadBomFromCSV(input io.Reader) (*Bom, error) {

	b := Bom{LineItems: []LineItem{} }
    reader := csv.NewReader(input) 
    reader.TrailingComma = true
    reader.TrimLeadingSpace = true

    header, err := reader.Read()
    if err != nil {
        log.Fatal(err)
    }
    var li *LineItem
    var el_count int
    var records []string
    var qty string
    for records, err = reader.Read(); err == nil; records, err = reader.Read() {
        qty = ""
        li = &LineItem{Elements: []string{}}
        for i, col := range header {
            switch strings.ToLower(col) {
                case "qty", "quantity":
                    appendField(&qty, &records[i])
                case "mpn", "manufacturer part number":
                    appendField(&li.Mpn, &records[i])
                case "mfg", "manufacturer":
                    appendField(&li.Manufacturer, &records[i])
                case "element", "id", "circuit element", "symbol_id", "symbol id":
                    for _, symb := range strings.Split(records[i], ",") {
                        symb = strings.TrimSpace(symb)
                        if !isShortName(symb) {
                            li.Elements = append(li.Elements, symb)
                        } else if *verbose {
                            log.Println("symbol not a ShortName, skipped: " + symb)
                        }
                    }
                case "function", "purpose", "role":
                    appendField(&li.Function, &records[i])
                case "form_factor", "form factor", "case/package", "package", "symbol", "footprint":
                    appendField(&li.FormFactor, &records[i])
                case "specs", "specifications", "properties", "attributes", "value":
                    appendField(&li.Specs, &records[i])
                case "comment", "comments", "note", "notes":
                    appendField(&li.Comment, &records[i])
                default:
                    // pass, no assignment
                    // TODO: should warn on this first time around?
            }
        }
        if qty != "" {
            if n, err := strconv.Atoi(qty); err == nil && n >= 0 {
                el_count = len(li.Elements)
                // XXX: kludge
                if n > 99999 || el_count > 99999 {
                    log.Fatal("too large a quantity of elements passed, crashing")
                } else if el_count > n {
                    if *verbose {
                        log.Println("more symbols than qty, taking all symbols")
                    }
                } else if el_count < n {
                    for j := 0; j < (n - el_count); j++ {
                        li.Elements = append(li.Elements, "")
                    }
                }
            } 
        }
        if len(li.Elements) == 0 {
            li.Elements = []string{"", }
        }
        b.LineItems = append(b.LineItems, *li)
    }
    if err.Error() != "EOF" {
        log.Fatal(err)
    }
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
