package cmd

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strings"

	"github.com/spf13/cobra"
	"golang.org/x/term"
)

type authClient struct {
	serverURL *string
}

func newAuthCmd(serverURL *string) *cobra.Command {
	a := &authClient{serverURL: serverURL}
	cmd := &cobra.Command{Use: "auth", Short: "Authentication commands"}
	cmd.AddCommand(&cobra.Command{Use: "register", Short: "Register new user", RunE: a.register})
	cmd.AddCommand(&cobra.Command{Use: "login", Short: "Login and store token", RunE: a.login})
	return cmd
}

func (a *authClient) register(cmd *cobra.Command, args []string) error {
	reader := bufio.NewReader(os.Stdin)
	fmt.Fprint(cmd.OutOrStdout(), "Email: ")
	email, _ := reader.ReadString('\n')
	email = strings.TrimSpace(email)
	password, err := promptPassword(cmd, "Password: ")
	if err != nil {
		return err
	}
	body := map[string]string{"email": email, "password": string(password)}
	b, _ := json.Marshal(body)
	resp, err := http.Post(*a.serverURL+"/api/v1/auth/register", "application/json", bytes.NewReader(b))
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 300 {
		return fmt.Errorf("register failed: %s", resp.Status)
	}
	fmt.Fprintln(cmd.OutOrStdout(), "Registered")
	return nil
}

func (a *authClient) login(cmd *cobra.Command, args []string) error {
	reader := bufio.NewReader(os.Stdin)
	fmt.Fprint(cmd.OutOrStdout(), "Email: ")
	email, _ := reader.ReadString('\n')
	email = strings.TrimSpace(email)
	password, err := promptPassword(cmd, "Password: ")
	if err != nil {
		return err
	}
	body := map[string]string{"email": email, "password": string(password)}
	b, _ := json.Marshal(body)
	resp, err := http.Post(*a.serverURL+"/api/v1/auth/login", "application/json", bytes.NewReader(b))
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 300 {
		return fmt.Errorf("login failed: %s", resp.Status)
	}
	var result struct {
		AccessToken  string `json:"access_token"`
		RefreshToken string `json:"refresh_token"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return err
	}
	if err := saveToken(result.AccessToken); err != nil {
		return err
	}
	if result.RefreshToken != "" {
		_ = saveRefresh(result.RefreshToken)
	}
	fmt.Fprintln(cmd.OutOrStdout(), "Logged in")
	return nil
}

func promptPassword(cmd *cobra.Command, prompt string) ([]byte, error) {
	fmt.Fprint(cmd.OutOrStdout(), prompt)
	pass, err := term.ReadPassword(int(os.Stdin.Fd()))
	fmt.Fprintln(cmd.OutOrStdout())
	return pass, err
}

func tokenPath() string {
	home, _ := os.UserHomeDir()
	return home + string(os.PathSeparator) + ".gophkeeper_token"
}

func refreshPath() string {
	home, _ := os.UserHomeDir()
	return home + string(os.PathSeparator) + ".gophkeeper_refresh"
}

func saveToken(token string) error {
	return os.WriteFile(tokenPath(), []byte(token), 0600)
}

func loadToken() (string, error) {
	b, err := os.ReadFile(tokenPath())
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(b)), nil
}

func saveRefresh(token string) error { return os.WriteFile(refreshPath(), []byte(token), 0600) }
func loadRefresh() (string, error) {
	b, err := os.ReadFile(refreshPath())
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(b)), nil
}

func ensureAccessToken() (string, error) {
	tok, err := loadToken()
	if err == nil && tok != "" {
		return tok, nil
	}
	// try refresh
	r, err := loadRefresh()
	if err != nil || r == "" {
		return "", fmt.Errorf("no access token, please login")
	}
	body := map[string]string{"refresh_token": r}
	b, _ := json.Marshal(body)
	resp, err := http.Post(getServerURL()+"/api/v1/auth/refresh", "application/json", bytes.NewReader(b))
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 300 {
		return "", fmt.Errorf("refresh failed: %s", resp.Status)
	}
	var out struct {
		AccessToken string `json:"access_token"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
		return "", err
	}
	if out.AccessToken == "" {
		return "", fmt.Errorf("empty access token on refresh")
	}
	_ = saveToken(out.AccessToken)
	return out.AccessToken, nil
}

// getServerURL is a fallback for refresh call when we don't have cmd context; default to localhost.
func getServerURL() string {
	if v, ok := os.LookupEnv("GOPHKEEPER_SERVER_URL"); ok && v != "" {
		return v
	}
	return "http://localhost:8080"
}
