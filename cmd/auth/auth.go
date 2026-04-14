package auth

import (
	"encoding/json"
	"fmt"

	"github.com/charmbracelet/huh"
	"github.com/eggnita/adapto_cms_cli/internal/client"
	"github.com/eggnita/adapto_cms_cli/internal/cmdutil"
	"github.com/eggnita/adapto_cms_cli/internal/config"
	"github.com/eggnita/adapto_cms_cli/internal/credentials"
	"github.com/eggnita/adapto_cms_cli/internal/httpclient"
	"github.com/eggnita/adapto_cms_cli/internal/output"
	"github.com/eggnita/adapto_cms_cli/internal/prompt"
	openapi_types "github.com/oapi-codegen/runtime/types"
	"github.com/spf13/cobra"
)

// Cmd is the root auth command.
var Cmd = &cobra.Command{
	Use:   "auth",
	Short: "Authentication commands",
}

func init() {
	Cmd.AddCommand(loginCmd)
	Cmd.AddCommand(registerCmd)
	Cmd.AddCommand(logoutCmd)
	Cmd.AddCommand(refreshCmd)
	Cmd.AddCommand(meCmd)
	Cmd.AddCommand(changePasswordCmd)
	Cmd.AddCommand(requestPasswordResetCmd)
	Cmd.AddCommand(resetPasswordCmd)
	Cmd.AddCommand(activateCmd)
	Cmd.AddCommand(resendActivationCmd)
	Cmd.AddCommand(loginGithubCmd)
	Cmd.AddCommand(callbackGithubCmd)
	Cmd.AddCommand(loginGoogleCmd)
	Cmd.AddCommand(switchTenantCmd)
	Cmd.AddCommand(orgsCmd)
}

var loginCmd = &cobra.Command{
	Use:   "login",
	Short: "Login with email and password",
	RunE: func(cmd *cobra.Command, args []string) error {
		email, _ := cmd.Flags().GetString("email")
		password, _ := cmd.Flags().GetString("password")

		var err error
		if email, err = prompt.RequireArg("email", email); err != nil {
			return err
		}
		if password, err = prompt.RequireArgSensitive("password", password); err != nil {
			return err
		}

		c, _, err := cmdutil.NewClient()
		if err != nil {
			return err
		}

		resp, err := c.LoginPasswordAuthLoginPostWithResponse(cmdutil.Ctx(), client.LoginRequest{
			Email:    openapi_types.Email(email),
			Password: password,
		})
		if err != nil {
			return err
		}
		if err := cmdutil.CheckErr(resp.StatusCode(), resp.Body); err != nil {
			return err
		}

		var data map[string]interface{}
		_ = json.Unmarshal(resp.Body, &data)

		accessToken, _ := data["access_token"].(string)
		refreshToken, _ := data["refresh_token"].(string)

		creds := &credentials.Credentials{
			AccessToken:  accessToken,
			RefreshToken: refreshToken,
		}

		// Resolve tenant via /auth/me
		tenantID, _ := selectTenant(accessToken)
		creds.TenantID = tenantID

		if err := credentials.Save(creds); err != nil {
			fmt.Fprintf(cmd.ErrOrStderr(), "Warning: could not save credentials: %v\n", err)
		}

		result := map[string]interface{}{
			"message": "Logged in successfully",
			"credentials_path": credentials.Path(),
		}
		if tenantID != "" {
			result["tenant_id"] = tenantID
		}
		output.Print(result, func(d interface{}) {
			fmt.Printf("Logged in. Credentials saved to %s\n", credentials.Path())
			if tenantID != "" {
				fmt.Printf("Active tenant: %s\n", tenantID)
			}
		})
		return nil
	},
}

var registerCmd = &cobra.Command{
	Use:   "register",
	Short: "Register a new account",
	RunE: func(cmd *cobra.Command, args []string) error {
		email, _ := cmd.Flags().GetString("email")
		password, _ := cmd.Flags().GetString("password")
		firstName, _ := cmd.Flags().GetString("first-name")
		lastName, _ := cmd.Flags().GetString("last-name")

		var err error
		if email, err = prompt.RequireArg("email", email); err != nil {
			return err
		}
		if password, err = prompt.RequireArgSensitive("password", password); err != nil {
			return err
		}

		c, _, err := cmdutil.NewClient()
		if err != nil {
			return err
		}

		body := client.RegisterRequest{
			Email:    openapi_types.Email(email),
			Password: password,
		}
		if firstName != "" {
			body.FirstName = &firstName
		}
		if lastName != "" {
			body.LastName = &lastName
		}

		resp, err := c.RegisterPasswordAuthRegisterPostWithResponse(cmdutil.Ctx(), body)
		if err != nil {
			return err
		}
		if err := cmdutil.CheckErr(resp.StatusCode(), resp.Body); err != nil {
			return err
		}

		var data map[string]interface{}
		_ = json.Unmarshal(resp.Body, &data)
		output.Print(data, func(d interface{}) {
			fmt.Println("Registration successful. Check your email to activate your account.")
		})
		return nil
	},
}

var logoutCmd = &cobra.Command{
	Use:   "logout",
	Short: "Logout (revoke refresh token)",
	RunE: func(cmd *cobra.Command, args []string) error {
		refreshToken, _ := cmd.Flags().GetString("refresh-token")

		// Fall back to stored refresh token
		if refreshToken == "" {
			if creds, err := credentials.Load(); err == nil {
				refreshToken = creds.RefreshToken
			}
		}
		if refreshToken == "" {
			return fmt.Errorf("no refresh token available: provide --refresh-token or login first")
		}

		c, _, err := cmdutil.NewClient()
		if err != nil {
			return err
		}

		resp, err := c.LogoutUserAuthLogoutPostWithResponse(cmdutil.Ctx(), client.RefreshTokenRequest{
			RefreshToken: refreshToken,
		})
		if err != nil {
			return err
		}
		if err := cmdutil.CheckErr(resp.StatusCode(), resp.Body); err != nil {
			return err
		}

		if err := credentials.Clear(); err != nil {
			fmt.Fprintf(cmd.ErrOrStderr(), "Warning: could not clear credentials: %v\n", err)
		}

		output.Success("Logged out successfully. Credentials cleared.")
		return nil
	},
}

var refreshCmd = &cobra.Command{
	Use:   "refresh",
	Short: "Refresh access token",
	RunE: func(cmd *cobra.Command, args []string) error {
		refreshToken, _ := cmd.Flags().GetString("refresh-token")

		// Fall back to stored refresh token
		if refreshToken == "" {
			if creds, err := credentials.Load(); err == nil {
				refreshToken = creds.RefreshToken
			}
		}
		if refreshToken == "" {
			return fmt.Errorf("no refresh token available: provide --refresh-token or login first")
		}

		c, _, err := cmdutil.NewClient()
		if err != nil {
			return err
		}

		resp, err := c.RefreshAccessTokenAuthRefreshPostWithResponse(cmdutil.Ctx(), client.RefreshTokenRequest{
			RefreshToken: refreshToken,
		})
		if err != nil {
			return err
		}
		if err := cmdutil.CheckErr(resp.StatusCode(), resp.Body); err != nil {
			return err
		}

		var data map[string]interface{}
		_ = json.Unmarshal(resp.Body, &data)

		// Update stored credentials
		if newToken, ok := data["access_token"].(string); ok {
			if creds, err := credentials.Load(); err == nil {
				creds.AccessToken = newToken
				if newRefresh, ok := data["refresh_token"].(string); ok && newRefresh != "" {
					creds.RefreshToken = newRefresh
				}
				if err := credentials.Save(creds); err != nil {
					fmt.Fprintf(cmd.ErrOrStderr(), "Warning: could not update credentials: %v\n", err)
				}
			}
		}

		output.Print(data, func(d interface{}) {
			fmt.Println("Token refreshed. Credentials updated.")
		})
		return nil
	},
}

var meCmd = &cobra.Command{
	Use:   "me",
	Short: "Get current user info",
	RunE: func(cmd *cobra.Command, args []string) error {
		c, _, err := cmdutil.NewClientWithAuth()
		if err != nil {
			return err
		}

		resp, err := c.GetMeAuthMeGetWithResponse(cmdutil.Ctx())
		if err != nil {
			return err
		}
		if err := cmdutil.CheckErr(resp.StatusCode(), resp.Body); err != nil {
			return err
		}

		if resp.JSON200 != nil {
			output.Print(resp.JSON200, func(d interface{}) {
				u := resp.JSON200
				pairs := [][2]string{
					{"ID", u.User.Id},
					{"Email", string(u.User.Email)},
					{"Status", u.User.Status},
					{"Verified", fmt.Sprintf("%v", u.User.IsEmailVerified)},
				}
				if u.User.FirstName != nil {
					pairs = append(pairs, [2]string{"First Name", *u.User.FirstName})
				}
				if u.User.LastName != nil {
					pairs = append(pairs, [2]string{"Last Name", *u.User.LastName})
				}
				output.KeyValue(pairs)
			})
		}
		return nil
	},
}

var changePasswordCmd = &cobra.Command{
	Use:   "change-password",
	Short: "Change your password",
	RunE: func(cmd *cobra.Command, args []string) error {
		currentPassword, _ := cmd.Flags().GetString("current-password")
		newPassword, _ := cmd.Flags().GetString("new-password")

		var err error
		if currentPassword, err = prompt.RequireArgSensitive("current-password", currentPassword); err != nil {
			return err
		}
		if newPassword, err = prompt.RequireArgSensitive("new-password", newPassword); err != nil {
			return err
		}

		c, _, err := cmdutil.NewClientWithAuth()
		if err != nil {
			return err
		}

		resp, err := c.ChangePasswordAuthChangePasswordPostWithResponse(cmdutil.Ctx(), client.ChangePasswordRequest{
			CurrentPassword: currentPassword,
			NewPassword:     newPassword,
		})
		if err != nil {
			return err
		}
		if err := cmdutil.CheckErr(resp.StatusCode(), resp.Body); err != nil {
			return err
		}

		output.Success("Password changed successfully.")
		return nil
	},
}

var requestPasswordResetCmd = &cobra.Command{
	Use:   "request-password-reset",
	Short: "Request a password reset email",
	RunE: func(cmd *cobra.Command, args []string) error {
		email, _ := cmd.Flags().GetString("email")
		var err error
		if email, err = prompt.RequireArg("email", email); err != nil {
			return err
		}

		c, _, err := cmdutil.NewClient()
		if err != nil {
			return err
		}

		resp, err := c.RequestPasswordResetAuthRequestPasswordResetPostWithResponse(cmdutil.Ctx(), client.RequestPasswordResetRequest{
			Email: openapi_types.Email(email),
		})
		if err != nil {
			return err
		}
		if err := cmdutil.CheckErr(resp.StatusCode(), resp.Body); err != nil {
			return err
		}

		output.Success("Password reset email sent. Check your inbox.")
		return nil
	},
}

var resetPasswordCmd = &cobra.Command{
	Use:   "reset-password",
	Short: "Reset password with token",
	RunE: func(cmd *cobra.Command, args []string) error {
		token, _ := cmd.Flags().GetString("token")
		newPassword, _ := cmd.Flags().GetString("new-password")

		var err error
		if token, err = prompt.RequireArg("token", token); err != nil {
			return err
		}
		if newPassword, err = prompt.RequireArgSensitive("new-password", newPassword); err != nil {
			return err
		}

		c, _, err := cmdutil.NewClient()
		if err != nil {
			return err
		}

		resp, err := c.ResetPasswordAuthResetPasswordPostWithResponse(cmdutil.Ctx(), client.ResetPasswordRequest{
			Token:       token,
			NewPassword: newPassword,
		})
		if err != nil {
			return err
		}
		if err := cmdutil.CheckErr(resp.StatusCode(), resp.Body); err != nil {
			return err
		}

		output.Success("Password reset successfully.")
		return nil
	},
}

var activateCmd = &cobra.Command{
	Use:   "activate",
	Short: "Activate account with token",
	RunE: func(cmd *cobra.Command, args []string) error {
		token, _ := cmd.Flags().GetString("token")
		var err error
		if token, err = prompt.RequireArg("token", token); err != nil {
			return err
		}

		c, _, err := cmdutil.NewClient()
		if err != nil {
			return err
		}

		resp, err := c.ActivateAccountAuthActivatePostWithResponse(cmdutil.Ctx(), &client.ActivateAccountAuthActivatePostParams{
			Token: token,
		})
		if err != nil {
			return err
		}
		if err := cmdutil.CheckErr(resp.StatusCode(), resp.Body); err != nil {
			return err
		}

		output.Success("Account activated successfully.")
		return nil
	},
}

var resendActivationCmd = &cobra.Command{
	Use:   "resend-activation",
	Short: "Resend activation email",
	RunE: func(cmd *cobra.Command, args []string) error {
		email, _ := cmd.Flags().GetString("email")
		var err error
		if email, err = prompt.RequireArg("email", email); err != nil {
			return err
		}

		c, _, err := cmdutil.NewClient()
		if err != nil {
			return err
		}

		resp, err := c.ResendActivationAuthResendActivationPostWithResponse(cmdutil.Ctx(), &client.ResendActivationAuthResendActivationPostParams{
			Email: openapi_types.Email(email),
		})
		if err != nil {
			return err
		}
		if err := cmdutil.CheckErr(resp.StatusCode(), resp.Body); err != nil {
			return err
		}

		output.Success("Activation email sent. Check your inbox.")
		return nil
	},
}

var loginGithubCmd = &cobra.Command{
	Use:   "login-github",
	Short: "Login via GitHub OAuth",
	RunE: func(cmd *cobra.Command, args []string) error {
		redirectURI, _ := cmd.Flags().GetString("redirect-uri")

		c, _, err := cmdutil.NewClient()
		if err != nil {
			return err
		}

		resp, err := c.LoginGithubAuthLoginGithubGetWithResponse(cmdutil.Ctx(), &client.LoginGithubAuthLoginGithubGetParams{
			RedirectUri: cmdutil.StringPtr(redirectURI),
		})
		if err != nil {
			return err
		}
		if err := cmdutil.CheckErr(resp.StatusCode(), resp.Body); err != nil {
			return err
		}

		var data map[string]interface{}
		_ = json.Unmarshal(resp.Body, &data)
		output.Print(data, func(d interface{}) {
			output.JSON(d)
		})
		return nil
	},
}

var callbackGithubCmd = &cobra.Command{
	Use:   "callback-github",
	Short: "Complete GitHub OAuth callback",
	RunE: func(cmd *cobra.Command, args []string) error {
		code, _ := cmd.Flags().GetString("code")
		redirectURI, _ := cmd.Flags().GetString("redirect-uri")

		var err error
		if code, err = prompt.RequireArg("code", code); err != nil {
			return err
		}

		c, _, err := cmdutil.NewClient()
		if err != nil {
			return err
		}

		resp, err := c.CallbackGithubAuthCallbackGithubGetWithResponse(cmdutil.Ctx(), &client.CallbackGithubAuthCallbackGithubGetParams{
			Code:        code,
			RedirectUri: cmdutil.StringPtr(redirectURI),
		})
		if err != nil {
			return err
		}
		if err := cmdutil.CheckErr(resp.StatusCode(), resp.Body); err != nil {
			return err
		}

		var data map[string]interface{}
		_ = json.Unmarshal(resp.Body, &data)
		output.Print(data, func(d interface{}) {
			output.JSON(d)
		})
		return nil
	},
}

var loginGoogleCmd = &cobra.Command{
	Use:   "login-google",
	Short: "Login via Google credential",
	RunE: func(cmd *cobra.Command, args []string) error {
		credential, _ := cmd.Flags().GetString("credential")
		var err error
		if credential, err = prompt.RequireArg("credential", credential); err != nil {
			return err
		}

		c, _, err := cmdutil.NewClient()
		if err != nil {
			return err
		}

		resp, err := c.LoginGoogleAuthLoginGooglePostWithResponse(cmdutil.Ctx(), client.GoogleIdTokenRequest{
			Credential: credential,
		})
		if err != nil {
			return err
		}
		if err := cmdutil.CheckErr(resp.StatusCode(), resp.Body); err != nil {
			return err
		}

		var data map[string]interface{}
		_ = json.Unmarshal(resp.Body, &data)
		output.Print(data, func(d interface{}) {
			output.JSON(d)
		})
		return nil
	},
}

func init() {
	loginCmd.Flags().String("email", "", "Email address")
	loginCmd.Flags().String("password", "", "Password")

	registerCmd.Flags().String("email", "", "Email address")
	registerCmd.Flags().String("password", "", "Password")
	registerCmd.Flags().String("first-name", "", "First name")
	registerCmd.Flags().String("last-name", "", "Last name")

	logoutCmd.Flags().String("refresh-token", "", "Refresh token to revoke")
	refreshCmd.Flags().String("refresh-token", "", "Refresh token")

	changePasswordCmd.Flags().String("current-password", "", "Current password")
	changePasswordCmd.Flags().String("new-password", "", "New password")

	requestPasswordResetCmd.Flags().String("email", "", "Email address")

	resetPasswordCmd.Flags().String("token", "", "Password reset token")
	resetPasswordCmd.Flags().String("new-password", "", "New password")

	activateCmd.Flags().String("token", "", "Activation token")
	resendActivationCmd.Flags().String("email", "", "Email address")

	loginGithubCmd.Flags().String("redirect-uri", "", "OAuth redirect URI")
	callbackGithubCmd.Flags().String("code", "", "OAuth authorization code")
	callbackGithubCmd.Flags().String("redirect-uri", "", "OAuth redirect URI")

	loginGoogleCmd.Flags().String("credential", "", "Google ID token")

	switchTenantCmd.Flags().String("tenant-id", "", "Tenant/organization ID to switch to")
}

var switchTenantCmd = &cobra.Command{
	Use:   "switch-tenant",
	Short: "Switch active tenant/organization",
	RunE: func(cmd *cobra.Command, args []string) error {
		tenantID, _ := cmd.Flags().GetString("tenant-id")

		creds, err := credentials.Load()
		if err != nil || creds.AccessToken == "" {
			return fmt.Errorf("not logged in: run 'adapto auth login' first")
		}

		if tenantID == "" {
			selected, err := selectTenant(creds.AccessToken)
			if err != nil {
				return err
			}
			tenantID = selected
		}

		if tenantID == "" {
			return fmt.Errorf("no tenant selected")
		}

		creds.TenantID = tenantID
		if err := credentials.Save(creds); err != nil {
			return fmt.Errorf("could not save credentials: %w", err)
		}

		output.Successf("Switched to tenant: %s", tenantID)
		return nil
	},
}

var orgsCmd = &cobra.Command{
	Use:   "orgs",
	Short: "List your organizations and their tenants",
	RunE: func(cmd *cobra.Command, args []string) error {
		c, _, err := cmdutil.NewClientWithAuth()
		if err != nil {
			return err
		}

		resp, err := c.ListMyOrgsOrgsGetWithResponse(cmdutil.Ctx())
		if err != nil {
			return err
		}
		if err := cmdutil.CheckErr(resp.StatusCode(), resp.Body); err != nil {
			return err
		}
		if resp.JSON200 == nil {
			fmt.Println("No organizations found.")
			return nil
		}

		orgs := *resp.JSON200

		activeTenant := ""
		if creds, err := credentials.Load(); err == nil {
			activeTenant = creds.TenantID
		}

		type tenantEntry struct {
			TenantID   string   `json:"tenant_id"`
			TenantName string   `json:"tenant_name"`
			Languages  []string `json:"languages"`
			Active     bool     `json:"active"`
		}
		type orgEntry struct {
			OrgID   string        `json:"org_id"`
			OrgName string        `json:"org_name"`
			Tenants []tenantEntry `json:"tenants"`
		}

		var entries []orgEntry
		for _, org := range orgs {
			entry := orgEntry{OrgID: org.Id, OrgName: org.Name}

			// Fetch tenants for this org
			tResp, err := c.ListOrgTenantsTenantsByOrgOrgIdGetWithResponse(cmdutil.Ctx(), org.Id)
			if err == nil && tResp.JSON200 != nil {
				for _, t := range *tResp.JSON200 {
					entry.Tenants = append(entry.Tenants, tenantEntry{
						TenantID:   t.Id,
						TenantName: t.Name,
						Languages:  t.EnabledLanguages,
						Active:     t.Id == activeTenant,
					})
				}
			}
			entries = append(entries, entry)
		}

		output.Print(entries, func(d interface{}) {
			if len(entries) == 0 {
				fmt.Println("No organizations found.")
				return
			}
			for _, e := range entries {
				fmt.Printf("Organization: %s (%s)\n", e.OrgName, e.OrgID)
				if len(e.Tenants) == 0 {
					fmt.Println("  No tenants")
				}
				for _, t := range e.Tenants {
					marker := ""
					if t.Active {
						marker = " (active)"
					}
					fmt.Printf("  Tenant: %s (%s)%s  [%s]\n", t.TenantName, t.TenantID, marker, fmt.Sprintf("%v", t.Languages))
				}
				fmt.Println()
			}
		})
		return nil
	},
}

// selectTenant fetches orgs and tenants, then prompts the user to pick one.
func selectTenant(accessToken string) (string, error) {
	cfg := config.Load()
	cfg.Token = accessToken
	c, err := httpclient.New(cfg)
	if err != nil {
		return "", err
	}

	// Fetch orgs
	orgsResp, err := c.ListMyOrgsOrgsGetWithResponse(cmdutil.Ctx())
	if err != nil {
		return "", nil // non-fatal
	}
	if orgsResp.JSON200 == nil || len(*orgsResp.JSON200) == 0 {
		return "", nil
	}

	orgs := *orgsResp.JSON200

	// Collect all tenants across all orgs
	type tenantInfo struct {
		ID      string
		Name    string
		OrgName string
	}
	var allTenants []tenantInfo

	for _, org := range orgs {
		tResp, err := c.ListOrgTenantsTenantsByOrgOrgIdGetWithResponse(cmdutil.Ctx(), org.Id)
		if err != nil || tResp.JSON200 == nil {
			continue
		}
		for _, t := range *tResp.JSON200 {
			allTenants = append(allTenants, tenantInfo{
				ID:      t.Id,
				Name:    t.Name,
				OrgName: org.Name,
			})
		}
	}

	if len(allTenants) == 0 {
		return "", nil
	}

	if len(allTenants) == 1 {
		fmt.Printf("Auto-selected tenant: %s (%s)\n", allTenants[0].Name, allTenants[0].OrgName)
		return allTenants[0].ID, nil
	}

	// Multiple tenants — prompt user
	if !prompt.IsTTY() {
		return "", fmt.Errorf("multiple tenants found but running non-interactively; use --tenant-id to specify one")
	}

	var options []huh.Option[string]
	for _, t := range allTenants {
		label := fmt.Sprintf("%s — %s", t.Name, t.OrgName)
		options = append(options, huh.NewOption(label, t.ID))
	}
	return prompt.AskSelect("Select a tenant:", options)
}
