package cli

import (
	"github.com/Horlarhyinka/fs-organizer/internal/worker"
	"github.com/spf13/cobra"
)

var (
	outputDir string
	move	bool
	dryRun  bool
	onConflict string
	excluded []string
)

var rootCmd = cobra.Command{
	Aliases: []string{"fs-org", "fso", "fs-organizer"},
	Use: "Start fs-organizer cli",
	Example: "fs-org [dir] --mode",
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) < 1 {
			panic("file path must be provided as an argument")
		}
		if outputDir == "" {
			panic("output directory --output or -o must be provided")
		}
		opts := worker.Option{
			ConflictOpt: onConflict,

		}
		if err := worker.Organize(args[0], outputDir, opts); err != nil {
			panic(err)
		}
	},
}

func Init() {
	rootCmd.Flags().StringVarP(&outputDir, "output", "o", "", "provide output directory for organized files")
	rootCmd.Flags().BoolVarP(&move, "move", "m", false, "if true files are moved from source")
	rootCmd.Flags().BoolVar(&dryRun, "dry-run", false, "logs operations without file movement/copy")
	rootCmd.Flags().StringVar(&onConflict, "on-conflict", worker.ConflictOptSkip, "skip|rename|overwrite - defines filename conflict behaviour")
	rootCmd.RegisterFlagCompletionFunc("on-conflict", func (c *cobra.Command, args []string, toCOmplete string) ([]string, cobra.ShellCompDirective) {
		return []string{worker.ConflictOptSkip, worker.ConflictOptRename, worker.ConflictOptOverwrite}, cobra.ShellCompDirectiveNoFileComp
	})
	rootCmd.Flags().StringSliceVar(&excluded, "exclude", nil, "define list of excluded filepaths")
}