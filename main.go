package main

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net/http"
	"sync"

	"golang.org/x/oauth2"
)

type Scopes []string

func (s *Scopes) String() string {
	return fmt.Sprintf("%s", *s)
}

func (s *Scopes) Set(value string) error {
	*s = append(*s, value)
	return nil
}

func randString() string {
	buf := make([]byte, 32)
	rand.Read(buf)
	return base64.StdEncoding.EncodeToString(buf)
}

func main() {
	var (
		port         = flag.Int("port", 8080, "Callback port")
		path         = flag.String("path", "/oauth/callback", "Callback path")
		clientID     = flag.String("id", "", "Client ID")
		clientSecret = flag.String("secret", "", "Client secret")
		authURL      = flag.String("auth", "https://localhost/oauth/authorize", "Authorization URL")
		tokenURL     = flag.String("token", "https://localhost/oauth/token", "Token URL")
		scopes       Scopes
	)
	flag.Var(&scopes, "scope", "oAuth scopes to authorize (can be specified multiple times")
	flag.Parse()

	config := &oauth2.Config{
		ClientID:     *clientID,
		ClientSecret: *clientSecret,
		Scopes:       scopes,
		RedirectURL:  fmt.Sprintf("http://127.0.0.1:%d%s", *port, *path),
		Endpoint: oauth2.Endpoint{
			AuthURL:  *authURL,
			TokenURL: *tokenURL,
		},
	}

	state := randString()
	url := config.AuthCodeURL(state, oauth2.AccessTypeOffline)
	fmt.Printf("Visit this URL in your browser:\n\n%s\n\n", url)

	ctx := context.Background()
	var wg sync.WaitGroup
	wg.Add(1)

	http.HandleFunc(*path, func(w http.ResponseWriter, r *http.Request) {
		defer wg.Done()

		if s := r.URL.Query().Get("state"); s != state {
			http.Error(w, fmt.Sprintf("Invalid state: %s", s), http.StatusUnauthorized)
			return
		}

		code := r.URL.Query().Get("code")
		token, err := config.Exchange(ctx, code)
		if err != nil {
			http.Error(w, fmt.Sprintf("Exchange error: %s", err), http.StatusServiceUnavailable)
			return
		}

		tokenJSON, err := json.MarshalIndent(token, "", "  ")
		if err != nil {
			http.Error(w, fmt.Sprintf("Token parse error: %s", err), http.StatusServiceUnavailable)
			return
		}

		w.Write(tokenJSON)
	})

	server := http.Server{
		Addr: fmt.Sprintf(":%d", *port),
	}

	go func() {
		if err := server.ListenAndServe(); err != http.ErrServerClosed {
			log.Fatalln(err)
		}
	}()

	wg.Wait()
	if err := server.Shutdown(ctx); err != nil {
		log.Fatalln(err)
	}
}
