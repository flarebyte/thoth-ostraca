package main

import (
    "os"
    "fmt"
)

// TODO: CLI Wiring
// - This is a minimal entry point without external deps.
// - For subcommands, switch on os.Args and call functions in internal/app.
// - Later, you can replace this with a CLI framework if desired.
func main() {
    args := os.Args[1:]
    if len(args) > 0 {
        switch args[0] {
        case "hello":
            fmt.Println("Hello World")
            return
        }
    }
    fmt.Println("Hello World")
}
