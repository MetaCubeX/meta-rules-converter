package main

import (
	"log"
	"runtime"

	"github.com/metacubex/meta-rules-converter/input"
	F "github.com/sagernet/sing/common/format"
	"github.com/spf13/cobra"
)

const (
	Version = "0.1"
)

var (
	inPath  string
	outType string
	outDir  string
)

var mainCommand = &cobra.Command{
	Use:  "convert",
	Long: F.ToString("convert v", Version, " (", runtime.Version(), ", ", runtime.GOOS, "/", runtime.GOARCH, ")"),
}

var commandSite = &cobra.Command{
	Use: "geosite",
	RunE: func(cmd *cobra.Command, args []string) error {
		return input.ConvertSite(cmd, inPath, outType, outDir)
	},
}

var commandIP = &cobra.Command{
	Use: "geoip",
	RunE: func(cmd *cobra.Command, args []string) error {
		return input.ConvertIP(cmd, inPath, outType, outDir)
	},
}

var commandClash = &cobra.Command{
	Use: "clash",
	RunE: func(cmd *cobra.Command, args []string) error {
		return input.ConvertClash(cmd, inPath, outType, outDir)
	},
}

var commandASN = &cobra.Command{
	Use: "asn",
	RunE: func(cmd *cobra.Command, args []string) error {
		return input.ConvertASN(cmd, inPath, outType, outDir)
	},
}

func init() {
	mainCommand.PersistentFlags().StringVarP(&inPath, "file", "f", "", "Input File Path")
	mainCommand.PersistentFlags().StringVarP(&outType, "type", "t", "", "Output Type")
	mainCommand.PersistentFlags().StringVarP(&outDir, "out", "o", "", "Output Path")

	mainCommand.AddCommand(commandSite)
	mainCommand.AddCommand(commandASN)
	mainCommand.AddCommand(commandIP)
	mainCommand.AddCommand(commandClash)
}

func main() {
	if err := mainCommand.Execute(); err != nil {
		log.Fatal(err)
	}
}
