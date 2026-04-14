package cmd

import (
	"fmt"

	"github.com/eggnita/adapto_cms_cli/cmd/articles"
	"github.com/eggnita/adapto_cms_cli/cmd/auth"
	"github.com/eggnita/adapto_cms_cli/cmd/categories"
	"github.com/eggnita/adapto_cms_cli/cmd/collections"
	"github.com/eggnita/adapto_cms_cli/cmd/files"
	"github.com/eggnita/adapto_cms_cli/cmd/microcopy"
	"github.com/eggnita/adapto_cms_cli/cmd/status"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var rootCmd = &cobra.Command{
	Use:   "adapto",
	Short: "Adapto CMS CLI",
	Long:  "Command-line interface for the Adapto CMS management API.",
	SilenceUsage:  true,
	SilenceErrors: true,
}

func init() {
	cobra.OnInitialize(initConfig)

	rootCmd.AddCommand(auth.Cmd)
	rootCmd.AddCommand(articles.Cmd)
	rootCmd.AddCommand(categories.Cmd)
	rootCmd.AddCommand(collections.Cmd)
	rootCmd.AddCommand(files.Cmd)
	rootCmd.AddCommand(microcopy.Cmd)
	rootCmd.AddCommand(status.Cmd)

	rootCmd.PersistentFlags().String("api-url", "", "Adapto API base URL (env: ADAPTO_API_URL)")
	rootCmd.PersistentFlags().String("token", "", "Bearer token (env: ADAPTO_TOKEN)")
	rootCmd.PersistentFlags().String("tenant-id", "", "Tenant ID (env: ADAPTO_TENANT_ID)")
	rootCmd.PersistentFlags().Bool("json", false, "Output as JSON instead of table")
	rootCmd.PersistentFlags().Bool("verbose", false, "Show HTTP request/response details")

	_ = viper.BindPFlag("api_url", rootCmd.PersistentFlags().Lookup("api-url"))
	_ = viper.BindPFlag("token", rootCmd.PersistentFlags().Lookup("token"))
	_ = viper.BindPFlag("tenant_id", rootCmd.PersistentFlags().Lookup("tenant-id"))
	_ = viper.BindPFlag("json", rootCmd.PersistentFlags().Lookup("json"))
	_ = viper.BindPFlag("verbose", rootCmd.PersistentFlags().Lookup("verbose"))
}

func initConfig() {
	viper.SetEnvPrefix("ADAPTO")
	viper.AutomaticEnv()

	// Map flag names to env var names
	_ = viper.BindEnv("api_url", "ADAPTO_API_URL")
	_ = viper.BindEnv("token", "ADAPTO_TOKEN")
	_ = viper.BindEnv("tenant_id", "ADAPTO_TENANT_ID")

	if viper.GetString("api_url") == "" {
		viper.SetDefault("api_url", "https://api.adaptocms.com")
	}
}

// Execute runs the root command.
func Execute() error {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(rootCmd.ErrOrStderr(), err)
		return err
	}
	return nil
}

// Root returns the root command for subcommand registration.
func Root() *cobra.Command {
	return rootCmd
}
