package main

import (
    "os"

    "github.com/flarebyte/thoth-ostraca/cmd/thoth/root"
)

func main() {
    if err := root.Execute(os.Args[1:]); err != nil {
        os.Exit(1)
    }
}
