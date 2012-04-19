package main

// CLI for bommom tools. Also used to launch web interface.

import (
	"flag"
	"fmt"
	"io"
	"log"
	"os"
    "path"
)

// Globals
var bomstore BomStore
var auth AuthService
var anonUser = &User{name: "common"}

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

func openBomStore() {

}

func initCmd() {
	err := NewJSONFileBomStore(*fileStorePath)
	if err != nil {
		log.Fatal(err)
	}
	bomstore, err = OpenJSONFileBomStore(*fileStorePath)
	if err != nil {
		log.Fatal(err)
	}
	bs, err := bomstore.GetStub(ShortName("common"), ShortName("gizmo"))
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
	bomstore.Persist(bs, b, "v001")
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
    var err error
	bomstore, err = OpenJSONFileBomStore(*fileStorePath)
	if err != nil {
		log.Fatal(err)
	}
	if auth == nil {
		auth = DummyAuth(true)
	}
	bs, b, err := bomstore.GetHead(ShortName(userStr), ShortName(nameStr))
	if err != nil {
		log.Fatal(err)
	}

	switch *outFormat {
	case "text", "":
		DumpBomAsText(bs, b, outFile)
	case "json":
		DumpBomAsJSON(bs, b, outFile)
	case "csv":
		DumpBomAsCSV(b, outFile)
	case "xml":
		DumpBomAsXML(bs, b, outFile)
	default:
		log.Fatal("Error: unknown/unimplemented format: " + *outFormat)
	}
}

func listCmd() {
	bomstore, err := OpenJSONFileBomStore(*fileStorePath)
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
		bomStubs, err = bomstore.ListBoms(ShortName(name))
		if err != nil {
			log.Fatal(err)
		}
	} else {
		// list all boms from all names
		// TODO: ERROR
		bomStubs, err = bomstore.ListBoms("")
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
	fmt.Println("Usage (flags must go first?):")
	fmt.Println("\tbommom [options] command [arguments]")
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

