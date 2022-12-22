package main

import (
	"fmt"
	"os"
)

func report(args ...string) {
}

func test(args ...string) {
}

func spice(args ...string) {
}

func help(args ...string) {
	fmt.Println("Production Rule: A self-timed circuit verification tool")
	fmt.Println("usage: pr <command> <flags...>")
	fmt.Println("")
	fmt.Println("Commands:")
	fmt.Println("  report - generate an aggregate performance report from the architectural simulation")
	fmt.Println("  test   - use the architectural simulation to create inject and expect files for the digital simulator")
	fmt.Println("  spice  - generate a spice simulation from that digital simulation for a particular process")
}

func main() {
	if len(os.Args) > 1 {
		switch os.Args[1] {
		case "report": report(os.Args[2:len(os.Args)]...)
		case "test": test(os.Args[2:len(os.Args)]...)
		case "spice": spice(os.Args[2:len(os.Args)]...)
		case "help": help(os.Args[2:len(os.Args)]...)
		default: fmt.Printf("error: unrecognized command '%s'\n", os.Args[1])
		}
	} else { 
		help()
	}	
}
