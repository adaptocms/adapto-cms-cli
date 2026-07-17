package cmd

import (
	"errors"
	"fmt"
	"os"

	"github.com/adaptocms/adapto-cms-cli/cmd/apikey"
	"github.com/adaptocms/adapto-cms-cli/cmd/articles"
	"github.com/adaptocms/adapto-cms-cli/cmd/auth"
	"github.com/adaptocms/adapto-cms-cli/cmd/categories"
	"github.com/adaptocms/adapto-cms-cli/cmd/collections"
	"github.com/adaptocms/adapto-cms-cli/cmd/files"
	"github.com/adaptocms/adapto-cms-cli/cmd/llminfo"
	"github.com/adaptocms/adapto-cms-cli/cmd/microcopy"
	"github.com/adaptocms/adapto-cms-cli/cmd/onboard"
	"github.com/adaptocms/adapto-cms-cli/cmd/org"
	"github.com/adaptocms/adapto-cms-cli/cmd/pages"
	"github.com/adaptocms/adapto-cms-cli/cmd/project"
	"github.com/adaptocms/adapto-cms-cli/cmd/status"
	"github.com/adaptocms/adapto-cms-cli/internal/httpclient"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var rootCmd = &cobra.Command{
	Use:           "adapto",
	Short:         "Adapto CMS CLI",
	Long:          "Command-line interface for the Adapto CMS Management API.",
	SilenceUsage:  true,
	SilenceErrors: true,
}

func init() {
	cobra.OnInitialize(initConfig)

	rootCmd.AddCommand(auth.Cmd)
	rootCmd.AddCommand(onboard.Cmd)
	rootCmd.AddCommand(org.Cmd)
	rootCmd.AddCommand(project.Cmd)
	rootCmd.AddCommand(apikey.Cmd)
	rootCmd.AddCommand(articles.Cmd)
	rootCmd.AddCommand(categories.Cmd)
	rootCmd.AddCommand(collections.Cmd)
	rootCmd.AddCommand(files.Cmd)
	rootCmd.AddCommand(llminfo.Cmd)
	rootCmd.AddCommand(microcopy.Cmd)
	rootCmd.AddCommand(pages.Cmd)
	rootCmd.AddCommand(status.Cmd)

	rootCmd.PersistentFlags().String("api-url", "", "Adapto Management API base URL (env: ADAPTO_CLI_API_URL)")
	rootCmd.PersistentFlags().String("token", "", "Bearer token (env: ADAPTO_CLI_TOKEN)")
	rootCmd.PersistentFlags().String("tenant-id", "", "Tenant ID (env: ADAPTO_CLI_TENANT_ID)")
	rootCmd.PersistentFlags().Bool("json", false, "Output as JSON instead of table")
	rootCmd.PersistentFlags().Bool("verbose", false, "Show HTTP request/response details")

	_ = viper.BindPFlag("api_url", rootCmd.PersistentFlags().Lookup("api-url"))
	_ = viper.BindPFlag("token", rootCmd.PersistentFlags().Lookup("token"))
	_ = viper.BindPFlag("tenant_id", rootCmd.PersistentFlags().Lookup("tenant-id"))
	_ = viper.BindPFlag("json", rootCmd.PersistentFlags().Lookup("json"))
	_ = viper.BindPFlag("verbose", rootCmd.PersistentFlags().Lookup("verbose"))
}

func initConfig() {
	// A bare ADAPTO prefix would resolve api_url from ADAPTO_API_URL, the client-site Public API var.
	viper.SetEnvPrefix("ADAPTO_CLI")
	viper.AutomaticEnv()

	_ = viper.BindEnv("api_url", "ADAPTO_CLI_API_URL")
	_ = viper.BindEnv("token", "ADAPTO_CLI_TOKEN")
	_ = viper.BindEnv("tenant_id", "ADAPTO_CLI_TENANT_ID")

	warnLegacyEnv()

	if viper.GetString("api_url") == "" {
		viper.SetDefault("api_url", "https://api.adaptocms.com")
	}
}

func warnLegacyEnv() {
	if os.Getenv("ADAPTO_API_URL") != "" && os.Getenv("ADAPTO_CLI_API_URL") == "" {
		fmt.Fprintln(os.Stderr, "Warning: ADAPTO_API_URL is ignored by the CLI (it configures client-site Public API access); use ADAPTO_CLI_API_URL for the Management API.")
	}
	for _, v := range [][2]string{{"ADAPTO_TOKEN", "ADAPTO_CLI_TOKEN"}, {"ADAPTO_TENANT_ID", "ADAPTO_CLI_TENANT_ID"}} {
		if os.Getenv(v[0]) != "" && os.Getenv(v[1]) == "" {
			fmt.Fprintf(os.Stderr, "Warning: %s is ignored by the CLI; use %s.\n", v[0], v[1])
		}
	}
}

// Execute runs the root command.
func Execute() error {
	if err := rootCmd.Execute(); err != nil {
		if errors.Is(err, httpclient.ErrSessionExpired) {
			fmt.Fprintln(rootCmd.ErrOrStderr(), httpclient.ErrSessionExpired)
		} else {
			fmt.Fprintln(rootCmd.ErrOrStderr(), err)
		}
		return err
	}
	return nil
}

// Root returns the root command for subcommand registration.
func Root() *cobra.Command {
	return rootCmd
}
