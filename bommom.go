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
	case "load", "dump", "serve":
		log.Fatal("Error: Unimplemented, sorry")
	case "init":
		log.Println("Initializing...")
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
	fmt.Println("\tload [file]\t import a BOM")
	fmt.Println("\tdump [name]\t dump a BOM to stdout")
	fmt.Println("\tserve\t\t serve up web interface over HTTP")
	fmt.Println("")
	fmt.Println("Extra command line options:")
	fmt.Println("")
	flag.PrintDefaults()
}
