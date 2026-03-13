package cmd

import "github.com/spf13/cobra"

func NewRootCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "gmcl",
		Short: "gmcl is a minimal Minecraft launcher",
	}

	cmd.AddCommand(
		NewUpdateCmd(),
		NewListCmd(),
		NewInstallCmd(),
		NewRemoveCmd(),
		NewStartCmd(),
	)

	return cmd
}

func Execute() {
	if err := NewRootCmd().Execute(); err != nil {
		panic(err)
	}
}
