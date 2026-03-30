package cmd

import (
	"github.com/boobam22/gmcl/cli"
	"github.com/spf13/cobra"
)

func NewRootCmd(g *cli.Gmcl) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "gmcl",
		Short: "Gmcl is a minimal Minecraft launcher",

		SilenceUsage:  true,
		SilenceErrors: true,
	}

	cmd.AddCommand(
		NewUpdateCmd(g),
		NewListCmd(g),
		NewInstallCmd(g),
		NewRemoveCmd(g),
		NewStartCmd(g),
	)

	return cmd
}
