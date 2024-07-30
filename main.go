package main

import (
	"log"
	"runtime"

	F "github.com/sagernet/sing/common/format"
	"github.com/spf13/cobra"
)

const (
	Version = "0.1"
)

var mainCommand = &cobra.Command{
	Use:  "convert",
	Long: F.ToString("convert v", Version, " (", runtime.Version(), ", ", runtime.GOOS, "/", runtime.GOARCH, ")"),
}

func main() {
	if err := mainCommand.Execute(); err != nil {
		log.Fatal(err)
	}
}
