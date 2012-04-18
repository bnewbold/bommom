package main

// CLI for bommom tools. Also used to launch web interface.

import (
	"flag"
	"fmt"
	"log"
)

// Command line flags
var (
	templatePath  = flag.String("templatepath", "./templates", "path to template directory")
	fileStorePath = flag.String("path", "./filestore", "path to flat file data store top-level directory")
	verbose       = flag.Bool("verbose", false, "print extra info")
	helpFlag      = flag.Bool("help", false, "print full help info")
	outFormat     = flag.String("format", "text", "command output format (for 'dump' etc)")
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
	case "load", "serve":
		log.Fatal("Error: Unimplemented, sorry")
	case "init":
		log.Println("Initializing...")
        initCmd()
    case "dump":
        log.Println("Dumping...")
        dumpCmd()
    case "list":
        listCmd()
	}
}

func initCmd() {
    _, err := NewJSONFileBomStore(*fileStorePath)
    if err != nil {
        log.Fatal(err)
    }
}

func dumpCmd() {
    b := makeTestBom()
    b.Version = "v001"
    bs := &BomStub{Name: "widget",
                    Owner: "common",
                    Description: "fancy stuff",
                    HeadVersion: b.Version,
                    IsPublicView: true,
                    IsPublicEdit: true}
    jfbs, err := OpenJSONFileBomStore(*fileStorePath)
    if err != nil {
        log.Fatal(err)
    }
    jfbs.Persist(bs, b, "v001")
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
	fmt.Println("\tload <file>\t import a BOM")
	fmt.Println("\tdump <user> <name>\t dump a BOM to stdout")
	fmt.Println("\tserve\t\t serve up web interface over HTTP")
	fmt.Println("")
	fmt.Println("Extra command line options:")
	fmt.Println("")
	flag.PrintDefaults()
}
