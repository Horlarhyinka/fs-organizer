package main

import "github.com/spf13/cobra"

var RootCmd = cobra.Command{
	Aliases: []string{"fs-org", "fso", "fs-organizer"},
	Use: "Start fs-organizer cli",
	Example: "fs-org [dir] --mode",
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) < 2 {
			panic("file path must be provided as an argument")
		}
		flags := cmd.Flags()
		excludes, err := flags.GetStringArray("exclude"); if err != nil {
			panic(err)
		}
		if err := RunWorker(args[1], "", excludes); err != nil {
			panic(err)
		}
	},
}