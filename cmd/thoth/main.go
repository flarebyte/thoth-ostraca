package main

import (
    "flag"
    "fmt"
    "os"
)

const helpText = `thoth - minimal CLI

Usage:
  thoth [command]

Available Commands:
  version    Print version

Flags:
  -h, --help    Show help
`

func main() {
    // Support -h and --help on the root command.
    fs := flag.NewFlagSet("thoth", flag.ContinueOnError)
    // Discard default output; we print our own help text.
    fs.SetOutput(os.Stdout)
    help := fs.Bool("help", false, "show help")
    fs.BoolVar(help, "h", false, "show help")

    // Parse all args; subcommand and its args remain in fs.Args().
    _ = fs.Parse(os.Args[1:])

    if *help {
        fmt.Print(helpText)
        return
    }

    args := fs.Args()
    if len(args) == 0 {
        fmt.Print(helpText)
        return
    }

    switch args[0] {
    case "version":
        fmt.Println("thoth dev")
        return
    case "help":
        fmt.Print(helpText)
        return
    default:
        // Unknown command: show help and non-zero exit.
        fmt.Print(helpText)
        os.Exit(2)
    }
}
