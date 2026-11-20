package main

import (
	"context"
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"os"
	"os/exec"
	"runtime"
	"time"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/gmail/v1"
	"google.golang.org/api/option"
)

const tokenFile = "token.json"

func getGmailService() (*gmail.Service, error) {
	ctx := context.Background()

	credentials, err := os.ReadFile("credentials.json")
	if err != nil {
		return nil, fmt.Errorf("credentials.json tidak ditemukan: %w", err)
	}

	config, err := google.ConfigFromJSON(
		credentials,
		gmail.GmailReadonlyScope,
	)
	if err != nil {
		return nil, err
	}

	token, err := loadToken(tokenFile)

	if err != nil {
		token, err = getTokenFromBrowser(config)
		if err != nil {
			return nil, err
		}

		if err := saveToken(tokenFile, token); err != nil {
			return nil, err
		}
	}

	client := config.Client(ctx, token)

	service, err := gmail.NewService(
		ctx,
		option.WithHTTPClient(client),
	)
	if err != nil {
		return nil, err
	}

	return service, nil
}

func getTokenFromBrowser(config *oauth2.Config) (*oauth2.Token, error) {
	listener, err := net.Listen("tcp", "127.0.0.1:8765")
	if err != nil {
		return nil, err
	}

	defer listener.Close()

	config.RedirectURL = "http://127.0.0.1:8765/oauth2callback"

	state := "gmail-mcp-auth-state"

	authURL := config.AuthCodeURL(
		state,
		oauth2.AccessTypeOffline,
		oauth2.SetAuthURLParam("prompt", "consent"),
	)

	fmt.Fprintln(os.Stderr, "Buka URL berikut untuk login:")
	fmt.Fprintln(os.Stderr, authURL)

	openBrowser(authURL)

	codeChannel := make(chan string, 1)

	server := &http.Server{}

	http.HandleFunc("/oauth2callback", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Query().Get("state") != state {
			http.Error(w, "Invalid state", http.StatusBadRequest)
			return
		}

		code := r.URL.Query().Get("code")

		if code == "" {
			http.Error(w, "Authorization code tidak ditemukan", http.StatusBadRequest)
			return
		}

		fmt.Fprintln(w, "Autentikasi berhasil. Silakan kembali ke terminal.")

		codeChannel <- code
	})

	go server.Serve(listener)

	code := <-codeChannel

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	server.Shutdown(ctx)

	token, err := config.Exchange(context.Background(), code)
	if err != nil {
		return nil, err
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
		cmd = exec.Command("cmd", "/c", "start", url)
	}

	if cmd != nil {
		_ = cmd.Start()
	}
}

func loadToken(path string) (*oauth2.Token, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var token oauth2.Token

	err = json.Unmarshal(data, &token)
	if err != nil {
		return nil, err
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
