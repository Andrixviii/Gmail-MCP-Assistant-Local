package main

import (
	"context"
	"crypto/rand"
	"crypto/subtle"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"net"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"time"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/gmail/v1"
	"google.golang.org/api/option"
)

// Mengambil direktori konfigurasi dari environment variable.
// Jika tidak tersedia, gunakan working directory saat ini.
func getConfigDir() (string, error) {
	dir := os.Getenv("GMAIL_MCP_CONFIG_DIR")

	if dir == "" {
		var err error
		dir, err = os.Getwd()
		if err != nil {
			return "", fmt.Errorf("gagal mendapatkan working directory: %w", err)
		}
	}

	absDir, err := filepath.Abs(dir)
	if err != nil {
		return "", fmt.Errorf("gagal menentukan direktori konfigurasi: %w", err)
	}

	return absDir, nil
}

// Mendapatkan lokasi credentials.json.
func getCredentialsPath() (string, error) {
	dir, err := getConfigDir()
	if err != nil {
		return "", err
	}

	return filepath.Join(dir, "credentials.json"), nil
}

// Mendapatkan lokasi token.json.
func getTokenPath() (string, error) {
	dir, err := getConfigDir()
	if err != nil {
		return "", err
	}

	return filepath.Join(dir, "token.json"), nil
}

func getGmailService() (*gmail.Service, error) {
	ctx := context.Background()

	// Membaca credentials.json.
	credentialsPath, err := getCredentialsPath()
	if err != nil {
		return nil, err
	}

	credentials, err := os.ReadFile(credentialsPath)
	if err != nil {
		return nil, fmt.Errorf(
			"gagal membaca credentials.json (%s): %w",
			credentialsPath,
			err,
		)
	}

	// Mengatur OAuth dengan akses read-only.
	config, err := google.ConfigFromJSON(
		credentials,
		gmail.GmailReadonlyScope,
	)
	if err != nil {
		return nil, fmt.Errorf("gagal membaca konfigurasi OAuth: %w", err)
	}

	// Membaca token.json.
	tokenPath, err := getTokenPath()
	if err != nil {
		return nil, err
	}

	token, err := loadToken(tokenPath)

	if errors.Is(err, os.ErrNotExist) {
		// Jika token belum tersedia, lakukan autentikasi browser.
		token, err = getTokenFromBrowser(config)
		if err != nil {
			return nil, fmt.Errorf("autentikasi OAuth gagal: %w", err)
		}

		if err := saveToken(tokenPath, token); err != nil {
			return nil, fmt.Errorf("gagal menyimpan token: %w", err)
		}
	} else if err != nil {
		return nil, fmt.Errorf("gagal membaca token: %w", err)
	}

	// Membuat HTTP client menggunakan token OAuth.
	client := config.Client(ctx, token)

	// Membuat Gmail API service.
	service, err := gmail.NewService(
		ctx,
		option.WithHTTPClient(client),
	)
	if err != nil {
		return nil, fmt.Errorf("gagal membuat Gmail service: %w", err)
	}

	return service, nil
}

func getTokenFromBrowser(config *oauth2.Config) (*oauth2.Token, error) {
	listener, err := net.Listen("tcp", "127.0.0.1:8765")
	if err != nil {
		return nil, fmt.Errorf("gagal membuka OAuth callback: %w", err)
	}
	defer listener.Close()

	config.RedirectURL = "http://127.0.0.1:8765/oauth2callback"

	// Membuat state acak untuk melindungi proses OAuth.
	stateBytes := make([]byte, 32)
	if _, err := rand.Read(stateBytes); err != nil {
		return nil, fmt.Errorf("gagal membuat OAuth state: %w", err)
	}

	state := base64.RawURLEncoding.EncodeToString(stateBytes)

	authURL := config.AuthCodeURL(
		state,
		oauth2.AccessTypeOffline,
		oauth2.SetAuthURLParam("prompt", "consent"),
	)

	type authResult struct {
		code string
		err  error
	}

	codeChannel := make(chan authResult, 1)

	// Menggunakan ServeMux lokal, bukan DefaultServeMux.
	mux := http.NewServeMux()

	mux.HandleFunc("/oauth2callback", func(w http.ResponseWriter, r *http.Request) {
		if subtle.ConstantTimeCompare(
			[]byte(r.URL.Query().Get("state")),
			[]byte(state),
		) != 1 {
			http.Error(w, "Invalid OAuth state", http.StatusBadRequest)
			return
		}

		if oauthErr := r.URL.Query().Get("error"); oauthErr != "" {
			http.Error(w, "Autentikasi ditolak: "+oauthErr, http.StatusBadRequest)

			codeChannel <- authResult{
				err: fmt.Errorf("Google OAuth error: %s", oauthErr),
			}
			return
		}

		code := r.URL.Query().Get("code")
		if code == "" {
			http.Error(w, "Authorization code tidak ditemukan", http.StatusBadRequest)

			codeChannel <- authResult{
				err: errors.New("authorization code tidak ditemukan"),
			}
			return
		}

		fmt.Fprintln(w, "Autentikasi berhasil. Silakan kembali ke terminal.")

		codeChannel <- authResult{
			code: code,
		}
	})

	server := &http.Server{
		Handler: mux,
	}

	go func() {
		if err := server.Serve(listener); err != nil && !errors.Is(err, http.ErrServerClosed) {
			fmt.Fprintln(os.Stderr, "OAuth callback server:", err)
		}
	}()

	fmt.Fprintln(os.Stderr, "Buka URL berikut untuk login:")
	fmt.Fprintln(os.Stderr, authURL)

	openBrowser(authURL)

	// Membatasi waktu tunggu autentikasi hingga 5 menit.
	timer := time.NewTimer(5 * time.Minute)
	defer timer.Stop()

	var result authResult

	select {
	case result = <-codeChannel:
	case <-timer.C:
		_ = server.Close()
		return nil, errors.New("waktu autentikasi OAuth habis")
	}

	shutdownCtx, cancel := context.WithTimeout(
		context.Background(),
		5*time.Second,
	)
	defer cancel()

	_ = server.Shutdown(shutdownCtx)

	if result.err != nil {
		return nil, result.err
	}

	token, err := config.Exchange(context.Background(), result.code)
	if err != nil {
		return nil, fmt.Errorf("gagal menukar authorization code: %w", err)
	}

	return token, nil
}

func openBrowser(url string) {
	var cmd *exec.Cmd

	switch runtime.GOOS {
	case "linux":
		cmd = exec.Command("xdg-open", url)
	case "darwin":
		cmd = exec.Command("open", url)
	case "windows":
		cmd = exec.Command("cmd", "/c", "start", "", url)
	}

	if cmd != nil {
		if err := cmd.Start(); err != nil {
			fmt.Fprintln(os.Stderr, "Tidak dapat membuka browser otomatis:", err)
		}
	} else {
		fmt.Fprintln(os.Stderr, "Silakan buka URL OAuth secara manual.")
	}
}

func loadToken(path string) (*oauth2.Token, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var token oauth2.Token

	if err := json.Unmarshal(data, &token); err != nil {
		return nil, fmt.Errorf("format token tidak valid: %w", err)
	}

	return &token, nil
}

func saveToken(path string, token *oauth2.Token) error {
	data, err := json.MarshalIndent(token, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(path, data, 0600)
}
