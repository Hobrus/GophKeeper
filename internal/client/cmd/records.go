package cmd

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strings"

	"github.com/spf13/cobra"
	"gophkeeper/internal/client/vault"
	cryptohelper "gophkeeper/internal/shared/crypto"
)

type recordsClient struct{ serverURL *string }

func newRecordsCmd(serverURL *string) *cobra.Command {
	r := &recordsClient{serverURL: serverURL}
	cmd := &cobra.Command{Use: "records", Short: "Manage records"}
	cmd.AddCommand(&cobra.Command{Use: "list", Short: "List records", RunE: r.list})
	cmd.AddCommand(&cobra.Command{Use: "add-login", Short: "Add login/password record", RunE: r.addLogin})
	cmd.AddCommand(&cobra.Command{Use: "get", Short: "Get record by id", Args: cobra.ExactArgs(1), RunE: r.get})
	cmd.AddCommand(&cobra.Command{Use: "delete", Short: "Delete record by id", Args: cobra.ExactArgs(1), RunE: r.delete})
	cmd.AddCommand(&cobra.Command{Use: "add-text", Short: "Add text record", RunE: r.addText})
	cmd.AddCommand(&cobra.Command{Use: "add-file", Short: "Add binary file record", Args: cobra.ExactArgs(1), RunE: r.addFile})
	cmd.AddCommand(&cobra.Command{Use: "add-card", Short: "Add bank card record", RunE: r.addCard})
	return cmd
}

func (r *recordsClient) list(cmd *cobra.Command, args []string) error {
	token, err := ensureAccessToken()
	if err != nil {
		return err
	}
	req, _ := http.NewRequest("GET", *r.serverURL+"/api/v1/records", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 300 {
		return fmt.Errorf("list failed: %s", resp.Status)
	}
	var items []map[string]any
	if err := json.NewDecoder(resp.Body).Decode(&items); err != nil {
		return err
	}
	enc := json.NewEncoder(os.Stdout)
	enc.SetIndent("", "  ")
	return enc.Encode(items)
}

func (r *recordsClient) addLogin(cmd *cobra.Command, args []string) error {
	token, err := ensureAccessToken()
	if err != nil {
		return err
	}
	key, err := vault.Load()
	if err != nil {
		return err
	}
	// prompt simple stdin
	var site, login, password string
	fmt.Fprint(cmd.OutOrStdout(), "Site: ")
	fmt.Fscanln(os.Stdin, &site)
	fmt.Fprint(cmd.OutOrStdout(), "Login: ")
	fmt.Fscanln(os.Stdin, &login)
	fmt.Fprint(cmd.OutOrStdout(), "Password: ")
	fmt.Fscanln(os.Stdin, &password)
	plaintext := map[string]string{"login": login, "password": password}
	pbytes, _ := json.Marshal(plaintext)
	aad := []byte("login:" + site)
	ct, err := cryptohelper.EncryptAESGCM(key, pbytes, aad)
	if err != nil {
		return err
	}
	body := map[string]any{
		"type":    "login",
		"meta":    map[string]string{"site": site},
		"payload": ct,
	}
	b, _ := json.Marshal(body)
	req, _ := http.NewRequest("POST", *r.serverURL+"/api/v1/records", bytes.NewReader(b))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 300 {
		return fmt.Errorf("add-login failed: %s", resp.Status)
	}
	fmt.Fprintln(cmd.OutOrStdout(), "Record stored")
	return nil
}

func (r *recordsClient) get(cmd *cobra.Command, args []string) error {
	token, err := ensureAccessToken()
	if err != nil {
		return err
	}
	key, err := vault.Load()
	if err != nil {
		return err
	}
	id := args[0]
	req, _ := http.NewRequest("GET", *r.serverURL+"/api/v1/records/"+id, nil)
	req.Header.Set("Authorization", "Bearer "+token)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 300 {
		return fmt.Errorf("get failed: %s", resp.Status)
	}
	var rec struct {
		ID      string            `json:"id"`
		Type    string            `json:"type"`
		Meta    map[string]string `json:"meta"`
		Payload []byte            `json:"payload"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&rec); err != nil {
		return err
	}
	// Determine AAD based on record type and metadata to match encryption
	var aad []byte
	switch strings.TrimSpace(rec.Type) {
	case "login":
		if s, ok := rec.Meta["site"]; ok {
			aad = []byte("login:" + s)
		}
	case "text":
		if s, ok := rec.Meta["title"]; ok {
			aad = []byte("text:" + strings.TrimSpace(s))
		}
	case "binary":
		if s, ok := rec.Meta["name"]; ok {
			aad = []byte("binary:" + s)
		}
	case "bank_card":
		if s, ok := rec.Meta["bank"]; ok {
			aad = []byte("bank_card:" + s)
		}
	}
	if len(aad) == 0 {
		aad = []byte(strings.TrimSpace(rec.Type))
	}
	pt, err := cryptohelper.DecryptAESGCM(key, rec.Payload, aad)
	if err != nil {
		return err
	}
	var content map[string]string
	_ = json.Unmarshal(pt, &content)
	enc := json.NewEncoder(os.Stdout)
	enc.SetIndent("", "  ")
	return enc.Encode(map[string]any{"id": rec.ID, "type": rec.Type, "meta": rec.Meta, "content": content})
}

func (r *recordsClient) delete(cmd *cobra.Command, args []string) error {
	token, err := ensureAccessToken()
	if err != nil {
		return err
	}
	id := args[0]
	req, _ := http.NewRequest("DELETE", *r.serverURL+"/api/v1/records/"+id, nil)
	req.Header.Set("Authorization", "Bearer "+token)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusNoContent {
		return fmt.Errorf("delete failed: %s", resp.Status)
	}
	fmt.Fprintln(cmd.OutOrStdout(), "Deleted")
	return nil
}

func (r *recordsClient) addText(cmd *cobra.Command, args []string) error {
	token, err := ensureAccessToken()
	if err != nil {
		return err
	}
	key, err := vault.Load()
	if err != nil {
		return err
	}
	var title string
	var text string
	fmt.Fprint(cmd.OutOrStdout(), "Title: ")
	fmt.Fscanln(os.Stdin, &title)
	fmt.Fprintln(cmd.OutOrStdout(), "Enter text, end with EOF (Ctrl+Z then Enter on Windows):")
	buf := new(bytes.Buffer)
	if _, err := buf.ReadFrom(os.Stdin); err != nil { /* ignore */
	}
	text = buf.String()
	pbytes, _ := json.Marshal(map[string]string{"text": text})
	aad := []byte("text:" + strings.TrimSpace(title))
	ct, err := cryptohelper.EncryptAESGCM(key, pbytes, aad)
	if err != nil {
		return err
	}
	body := map[string]any{
		"type":    "text",
		"meta":    map[string]string{"title": title},
		"payload": ct,
	}
	b, _ := json.Marshal(body)
	req, _ := http.NewRequest("POST", *r.serverURL+"/api/v1/records", bytes.NewReader(b))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 300 {
		return fmt.Errorf("add-text failed: %s", resp.Status)
	}
	fmt.Fprintln(cmd.OutOrStdout(), "Record stored")
	return nil
}

func (r *recordsClient) addFile(cmd *cobra.Command, args []string) error {
	token, err := ensureAccessToken()
	if err != nil {
		return err
	}
	key, err := vault.Load()
	if err != nil {
		return err
	}
	path := args[0]
	data, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	name := path
	aad := []byte("binary:" + name)
	ct, err := cryptohelper.EncryptAESGCM(key, data, aad)
	if err != nil {
		return err
	}
	body := map[string]any{"type": "binary", "meta": map[string]string{"name": name}, "payload": ct}
	b, _ := json.Marshal(body)
	req, _ := http.NewRequest("POST", *r.serverURL+"/api/v1/records", bytes.NewReader(b))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 300 {
		return fmt.Errorf("add-file failed: %s", resp.Status)
	}
	fmt.Fprintln(cmd.OutOrStdout(), "Record stored")
	return nil
}

func (r *recordsClient) addCard(cmd *cobra.Command, args []string) error {
	token, err := ensureAccessToken()
	if err != nil {
		return err
	}
	key, err := vault.Load()
	if err != nil {
		return err
	}
	var bank, holder, number, exp, cvv string
	fmt.Fprint(cmd.OutOrStdout(), "Bank: ")
	fmt.Fscanln(os.Stdin, &bank)
	fmt.Fprint(cmd.OutOrStdout(), "Holder: ")
	fmt.Fscanln(os.Stdin, &holder)
	fmt.Fprint(cmd.OutOrStdout(), "Number: ")
	fmt.Fscanln(os.Stdin, &number)
	fmt.Fprint(cmd.OutOrStdout(), "Exp (MM/YY): ")
	fmt.Fscanln(os.Stdin, &exp)
	fmt.Fprint(cmd.OutOrStdout(), "CVV: ")
	fmt.Fscanln(os.Stdin, &cvv)
	content := map[string]string{"holder": holder, "number": number, "exp": exp, "cvv": cvv}
	pbytes, _ := json.Marshal(content)
	aad := []byte("bank_card:" + bank)
	ct, err := cryptohelper.EncryptAESGCM(key, pbytes, aad)
	if err != nil {
		return err
	}
	body := map[string]any{"type": "bank_card", "meta": map[string]string{"bank": bank}, "payload": ct}
	b, _ := json.Marshal(body)
	req, _ := http.NewRequest("POST", *r.serverURL+"/api/v1/records", bytes.NewReader(b))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 300 {
		return fmt.Errorf("add-card failed: %s", resp.Status)
	}
	fmt.Fprintln(cmd.OutOrStdout(), "Record stored")
	return nil
}
