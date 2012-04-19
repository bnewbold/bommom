package main

// Bom/BomStub conversion/dump/load routines

import (
	"encoding/csv"
	"encoding/json"
	"encoding/xml"
	"fmt"
	"io"
	"log"
	"strings"
)

// --------------------- text (CLI only ) -----------------------

func DumpBomAsText(bs *BomStub, b *Bom, out io.Writer) {
	fmt.Fprintln(out)
	fmt.Fprintf(out, "%s (version %s, created %s)\n", bs.Name, b.Version, b.Created)
	fmt.Fprintf(out, "Creator: %s\n", bs.Owner)
	if bs.Description != "" {
		fmt.Fprintf(out, "Description: %s\n", bs.Description)
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

func DumpBomAsJSON(bs *BomStub, b *Bom, out io.Writer) {

	obj := map[string]interface{}{
		"bom_meta": bs,
		"bom":      b,
	}

	enc := json.NewEncoder(out)
	if err := enc.Encode(&obj); err != nil {
		log.Fatal(err)
	}
}

func LoadBomFromJSON(input io.Reader) (*BomStub, *Bom, error) {

    bs := &BomStub{}
    b := &Bom{}

	obj := map[string]interface{}{
		"bom_meta": &bs,
		"bom":      &b,
	}

    fmt.Println(obj)

	enc := json.NewDecoder(input)
	if err := enc.Decode(&obj); err != nil {
		log.Fatal(err)
	}
    if &bs == nil || &b == nil {
        log.Fatal("didn't load successfully")
    }
    fmt.Println(bs)
    fmt.Println(b)
    return bs, b, nil
}

// --------------------- XML -----------------------

func DumpBomAsXML(bs *BomStub, b *Bom, out io.Writer) {

	enc := xml.NewEncoder(out)
	if err := enc.Encode(bs); err != nil {
		log.Fatal(err)
	}
	if err := enc.Encode(b); err != nil {
		log.Fatal(err)
	}
}

func LoadBomFromXML(input io.Reader) (*BomStub, *Bom, error) {

    bs := BomStub{}
    b := Bom{}

	enc := xml.NewDecoder(input)
	if err := enc.Decode(&bs); err != nil {
		log.Fatal(err)
	}
	if err := enc.Decode(&b); err != nil {
		log.Fatal(err)
	}
    return &bs, &b, nil
}
