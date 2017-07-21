package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"strings"
	"text/template"

	"github.com/go-redis/redis"
	"github.com/kizkoh/rca/rca"
	"github.com/pkg/errors"
)

// debug is extended bool and output debug message
type debug bool

func (debug debug) Printf(f string, v ...interface{}) {
	if debug {
		log.Printf(f, v...)
	}
}

// DEBUG is global debug type
var DEBUG debug

// ClusterNode is redis cluster node struct
type ClusterNode struct {
	id string
	// FIXME: ip addr
	ip      string
	host    string
	port    uint64
	states  []string
	slave   bool
	master  bool
	slaveof string
}

func main() {
	var help = false
	var verbose = false

	// parse args
	flags := flag.NewFlagSet(App.Name, flag.ContinueOnError)

	flags.BoolVar(&verbose, "verbose", verbose, "verbose")
	flags.BoolVar(&help, "h", help, "help")
	flags.BoolVar(&help, "help", help, "help")
	flags.BoolVar(&help, "version", help, "version")

	flags.Usage = func() { usage() }
	if err := flags.Parse(os.Args[1:]); err != nil {
		err = errors.Wrap(err, fmt.Sprintf("%v-%v failed: ", App.Name, App.Version))
		fmt.Printf("%v-%v failed: %v\n", App.Name, App.Version, err)
		os.Exit(1)
	}

	if help {
		usage()
		os.Exit(0)
	}

	DEBUG = debug(verbose)

	args := flags.Args()
	var arg string
	if len(args) == 0 {
		arg = "127.0.0.1:6379"
	} else {
		arg = args[len(args)-1]
	}

	client := redis.NewClient(&redis.Options{
		Addr: arg,
	})
	cluster, err := rca.ClusterNodes(client)
	if err != nil {
		err = errors.Wrap(err, fmt.Sprintf("%v-%v failed: ", App.Name, App.Version))
		fmt.Printf("%v-%v failed: %v\n", App.Name, App.Version, err)
		os.Exit(1)
	}

	var myself rca.ClusterNode
	for _, node := range cluster {
		for _, state := range node.States {
			if state == "myself" {
				myself = node
			}
		}
	}

	fmt.Printf("myself:\n")
	fmt.Printf("  id: %s\n", myself.ID)
	fmt.Printf("  host: %s\n", myself.Host)
	fmt.Printf("  port: %d\n", myself.Port)
	fmt.Printf("  state: ")
	for i, state := range myself.States {
		if len(myself.States)-1 != i {
			fmt.Printf("%s,", state)
		} else {
			fmt.Printf("%s\n", state)
		}
	}
	if myself.Master {
		fmt.Printf("  slaves:\n")
		for _, node := range cluster {
			if node.SlaveOf == myself.ID {
				fmt.Printf("  - id: %s\n", node.ID)
				fmt.Printf("    host: %s\n", node.Host)
				fmt.Printf("    port: %d\n", node.Port)
				fmt.Printf("    state: ")
				for i, state := range node.States {
					if len(node.States)-1 != i {
						fmt.Printf("%s,", state)
					} else {
						fmt.Printf("%s\n", state)
					}
				}
			}
		}
	}
	if myself.Slave {
		fmt.Printf("  slaveof:\n")
		for _, node := range cluster {
			if node.ID == myself.SlaveOf {
				fmt.Printf("  - id: %s\n", node.ID)
				fmt.Printf("    host: %s\n", node.Host)
				fmt.Printf("    port: %d\n", node.Port)
				fmt.Printf("    state: ")
				for i, state := range node.States {
					if len(node.States)-1 != i {
						fmt.Printf("%s,", state)
					} else {
						fmt.Printf("%s\n", state)
					}
				}
			}
		}
	}
}

func usage() {
	helpText := `
usage:
   {{.Name}} [command options]

version:
   {{.Version}}

author:
   kizkoh<GitHub: https://github.com/kizkoh>

options:
   --verbose                                    Print verbose messages
   --help, -h                                   Show help
   --version                                    Print the version
`
	t := template.New("usage")
	t, _ = t.Parse(strings.TrimSpace(helpText))
	t.Execute(os.Stdout, App)
	fmt.Println()
}
