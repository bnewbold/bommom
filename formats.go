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
                    qty = strings.TrimSpace(records[i])
                case "mpn", "manufacturer part number":
                    li.Mpn = strings.TrimSpace(records[i])
                case "mfg", "manufacturer":
                    li.Manufacturer = strings.TrimSpace(records[i])
                case "symbol", "symbols":
                    for _, symb := range strings.Split(records[i], ",") {
                        symb = strings.TrimSpace(symb)
                        if !isShortName(symb) {
                            li.Elements = append(li.Elements, symb)
                        } else if *verbose {
                            log.Println("symbol not a ShortName, skipped: " + symb)
                        }
                    }
                case "description", "function":
                    li.Description = strings.TrimSpace(records[i])
                case "comment", "comments":
                    li.Comment = strings.TrimSpace(records[i])
                default:
                    // pass, no assignment
            }
        }
        if qty != "" {
            if n, err := strconv.Atoi(qty); err == nil && n >= 0 {
                el_count = len(li.Elements)
                // XXX: kludge, should handle this better
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
