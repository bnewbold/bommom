package main

// CLI for bommom tools. Also used to launch web interface.

import (
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"path"
	"time"
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
	inFormat      = ""
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
	case "serve":
		log.Fatal("Error: Unimplemented, sorry")
	case "init":
		log.Println("Initializing...")
		initCmd()
	case "dump":
		dumpCmd()
	case "load":
		loadCmd()
	case "convert":
		convertCmd()
	case "list":
		listCmd()
	}
}

func openBomStore() {
	// defaults to JSON file store
	var err error
	bomstore, err = OpenJSONFileBomStore(*fileStorePath)
	if err != nil {
		log.Fatal(err)
	}
}

func dumpOut(fname string, bs *BomStub, b *Bom) {
	var outFile io.Writer
	if fname == "" {
		outFile = os.Stdout
	} else {
		// if no outFormat defined, infer from file extension
		if *outFormat == "" {
			switch ext := path.Ext(fname); ext {
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
		f, err := os.Create(fname)
		if err != nil {
			log.Fatal(err)
		}
		defer f.Close()
		outFile = io.Writer(f)
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

func loadIn(fname string) (bs *BomStub, b *Bom) {

	if inFormat == "" {
		switch ext := path.Ext(fname); ext {
		case ".json", ".JSON":
			inFormat = "json"
		case ".csv", ".CSV":
			inFormat = "csv"
		case ".xml", ".XML":
			inFormat = "xml"
		default:
			log.Fatal("Unknown file extention (use -format): " + ext)
		}
	}

	infile, err := os.Open(fname)
	if err != nil {
		log.Fatal(err)
	}
	defer infile.Close()

	switch inFormat {
	case "json":
		bs, b, err = LoadBomFromJSON(infile)
	case "csv":
		b, err = LoadBomFromCSV(infile)
	case "xml":
		bs, b, err = LoadBomFromXML(infile)
	default:
		log.Fatal("Error: unknown/unimplemented format: " + *outFormat)
	}
	if err != nil {
		log.Fatal(err)
	}
	return bs, b
}

func initCmd() {

	openBomStore()
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

	if !isShortName(userStr) || !isShortName(nameStr) {
		log.Fatal("Error: not valid ShortName: " + userStr +
			" and/or " + nameStr)
	}

	var fname string
	if flag.NArg() == 4 {
		fname = flag.Arg(3)
	} else {
		fname = ""
	}

	openBomStore()

	if auth == nil {
		auth = DummyAuth(true)
	}
	bs, b, err := bomstore.GetHead(ShortName(userStr), ShortName(nameStr))
	if err != nil {
		log.Fatal(err)
	}

	dumpOut(fname, bs, b)
}

func loadCmd() {

	var userName, bomName, version string
	if flag.NArg() == 4 {
		userName = anonUser.name
		bomName = flag.Arg(2)
		version = flag.Arg(3)
	} else if flag.NArg() == 5 {
		userName = flag.Arg(2)
		bomName = flag.Arg(3)
		version = flag.Arg(4)
	} else {
		log.Fatal("Error: wrong number of arguments (expected input file, optional username, bomname)")
	}
	inFname := flag.Arg(1)

	if !(isShortName(userName) && isShortName(bomName) && isShortName(version)) {
		log.Fatal("user, name, and version must be ShortNames")
	}

	bs, b := loadIn(inFname)
	if inFormat == "csv" && bs == nil {
		// TODO: from inname? if ShortName?
		bs = &BomStub{}
	}

	bs.Owner = userName
	bs.Name = bomName
	b.Progeny = "File import from " + inFname + " (" + inFormat + ")"
	b.Created = time.Now()

	openBomStore()

	bomstore.Persist(bs, b, ShortName(version))
}

func convertCmd() {
	if flag.NArg() != 2 && flag.NArg() != 3 {
		log.Fatal("Error: wrong number of arguments (expected input and output files)")
	}

	// should refactor this to open both files first, then do processing? not
	// sure what best practice is.

	inFname := flag.Arg(1)
	outFname := flag.Arg(2)

	bs, b := loadIn(inFname)

	if b == nil {
		log.Fatal("null bom")
	}
	if inFormat == "csv" && bs == nil {
		// TODO: from inname? if ShortName?
		bs = &BomStub{Name: "untitled", Owner: anonUser.name}
	}

	if err := bs.Validate(); err != nil {
		log.Fatal("loaded bomstub not valid: " + err.Error())
	}
	if err := b.Validate(); err != nil {
		log.Fatal("loaded bom not valid: " + err.Error())
	}

	dumpOut(outFname, bs, b)
}

func listCmd() {

	openBomStore()
	var bomStubs []BomStub
	var err error
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
	fmt.Println("\tload <file.type> [user] <bom_name> <version>\t import a BOM")
	fmt.Println("\tdump <user> <name> [file.type]\t dump a BOM to stdout")
	fmt.Println("\tconvert <infile.type> <outfile.type>\t convert a BOM file")
	fmt.Println("\tserve\t\t serve up web interface over HTTP")
	fmt.Println("")
	fmt.Println("Extra command line options:")
	fmt.Println("")
	flag.PrintDefaults()
}
