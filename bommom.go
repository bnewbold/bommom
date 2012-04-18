package main

// CLI for bommom tools. Also used to launch web interface.

import (
	"encoding/csv"
	"encoding/json"
	"encoding/xml"
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"path"
	"strings"
)

// Command line flags
var (
	templatePath  = flag.String("templatepath", "./templates", "path to template directory")
	fileStorePath = flag.String("path", "./filestore", "path to flat file data store top-level directory")
	verbose       = flag.Bool("verbose", false, "print extra info")
	helpFlag      = flag.Bool("help", false, "print full help info")
	outFormat     = flag.String("format", "", "command output format (for 'dump' etc)")
)

func main() {

	// Parse configuration
	flag.Parse()
	if *verbose {
		log.Println("template dir:", *templatePath)
		log.Println("filestore dir:", *fileStorePath)
		log.Println("anon user:", anonUser.name)
	}

	// Process command
	if *helpFlag {
		printUsage()
		return
	}

	if flag.NArg() < 1 {
		printUsage()
		fmt.Println()
		log.Fatal("Error: No command specified")
	}

	switch flag.Arg(0) {
	default:
		log.Fatal("Error: unknown command: ", flag.Arg(0))
	case "load", "serve", "convert":
		log.Fatal("Error: Unimplemented, sorry")
	case "init":
		log.Println("Initializing...")
		initCmd()
	case "dump":
		dumpCmd()
	case "list":
		listCmd()
	}
}

func initCmd() {
	jfbs, err := NewJSONFileBomStore(*fileStorePath)
	if err != nil {
		log.Fatal(err)
	}
	jfbs, err = OpenJSONFileBomStore(*fileStorePath)
	if err != nil {
		log.Fatal(err)
	}
	bs, err := jfbs.GetStub(ShortName("common"), ShortName("gizmo"))
	if err == nil {
		// dummy BomStub already exists?
		return
	}
	b := makeTestBom()
	b.Version = "v001"
	bs = &BomStub{Name: "gizmo",
		Owner:        "common",
		Description:  "fancy stuff",
		HeadVersion:  b.Version,
		IsPublicView: true,
		IsPublicEdit: true}
	jfbs.Persist(bs, b, "v001")
}

func dumpCmd() {
	if flag.NArg() != 3 && flag.NArg() != 4 {
		log.Fatal("Error: wrong number of arguments (expected user and BOM name, optional file)")
	}
	userStr := flag.Arg(1)
	nameStr := flag.Arg(2)
	var outFile io.Writer
	outFile = os.Stdout
	if flag.NArg() == 4 {
		f, err := os.Create(flag.Arg(3))
		if err != nil {
			log.Fatal(err)
		}
		defer f.Close()
		outFile = io.Writer(f)
		// if no outFormat defined, infer from file extension
		if *outFormat == "" {
			switch ext := path.Ext(f.Name()); ext {
			case "", ".txt", ".text":
				// pass
			case ".json":
				*outFormat = "json"
			case ".csv":
				*outFormat = "csv"
			case ".xml":
				*outFormat = "xml"
			default:
				log.Fatal("Unknown file extention (use -format): " + ext)
			}
		}
	}

	if !isShortName(userStr) || !isShortName(nameStr) {
		log.Fatal("Error: not valid ShortName: " + userStr +
			" and/or " + nameStr)
	}
	jfbs, err := OpenJSONFileBomStore(*fileStorePath)
	if err != nil {
		log.Fatal(err)
	}
	if auth == nil {
		auth = DummyAuth(true)
	}
	bs, b, err := jfbs.GetHead(ShortName(userStr), ShortName(nameStr))
	if err != nil {
		log.Fatal(err)
	}

	switch *outFormat {
	case "text", "":
		DumpBomAsText(bs, b, outFile)
	case "json":
		DumpBomAsJSON(bs, b, outFile)
	case "csv":
		DumpBomAsCSV(bs, b, outFile)
	case "xml":
		DumpBomAsXML(bs, b, outFile)
	default:
		log.Fatal("Error: unknown/unimplemented format: " + *outFormat)
	}
}

func listCmd() {
	jfbs, err := OpenJSONFileBomStore(*fileStorePath)
	if err != nil {
		log.Fatal(err)
	}
	var bomStubs []BomStub
	if flag.NArg() > 2 {
		log.Fatal("Error: too many arguments...")
	}
	if flag.NArg() == 2 {
		name := flag.Arg(1)
		if !isShortName(name) {
			log.Fatal("Error: not a possible username: " + name)
		}
		bomStubs, err = jfbs.ListBoms(ShortName(name))
		if err != nil {
			log.Fatal(err)
		}
	} else {
		// list all boms from all names
		// TODO: ERROR
		bomStubs, err = jfbs.ListBoms("")
		if err != nil {
			log.Fatal(err)
		}
	}
	for _, bs := range bomStubs {
		fmt.Println(bs.Owner + "/" + bs.Name)
	}
}

func printUsage() {
	fmt.Println("bommom is a tool for managing and publishing electronics BOMs")
	fmt.Println("")
	fmt.Println("Usage:")
	fmt.Println("\tbommom command [options]")
	fmt.Println("")
	fmt.Println("Commands:")
	fmt.Println("")
	fmt.Println("\tinit \t\t initialize BOM and authentication datastores")
	fmt.Println("\tlist [user]\t\t list BOMs, optionally filtered by user")
	fmt.Println("\tload <file.type> [user] [bom_name]\t import a BOM")
	fmt.Println("\tdump <user> <name> [file.type]\t dump a BOM to stdout")
	fmt.Println("\tconvert <infile.type> [outfile.type]\t convert a BOM file")
	fmt.Println("\tserve\t\t serve up web interface over HTTP")
	fmt.Println("")
	fmt.Println("Extra command line options:")
	fmt.Println("")
	flag.PrintDefaults()
}

// -------- conversion/dump/load routines

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

func DumpBomAsCSV(bs *BomStub, b *Bom, out io.Writer) {
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

func DumpBomAsXML(bs *BomStub, b *Bom, out io.Writer) {

	/*
	   obj := map[string] interface{} {
	       "BomMeta": bs,
	       "Bom": b, 
	   }
	*/

	enc := xml.NewEncoder(out)
	if err := enc.Encode(bs); err != nil {
		log.Fatal(err)
	}
	if err := enc.Encode(b); err != nil {
		log.Fatal(err)
	}
}
