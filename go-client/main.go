package main

import (
	"context"
	"encoding/json"
	"log"
	"net/http"

	oidc "github.com/coreos/go-oidc"
	oauth2 "golang.org/x/oauth2"
)


var (
	clientID     = "myclient"
	clientSecret = "osvD9ToRP0WaMlNkh5fsa51IMhBLhQsg"
)

func main() {
	ctx := context.Background()

	provider, err := oidc.NewProvider(ctx, "http://localhost:8080/auth/realms/myrealm")

	if err != nil {
		log.Fatalf("Failed to get provider: %v", err)
	}

	config := oauth2.Config{
		ClientID:     clientID,
		ClientSecret: clientSecret,
		Endpoint:     provider.Endpoint(),
		RedirectURL:  "http://localhost:8081/auth/callback",
		Scopes: []string{oidc.ScopeOpenID, "profile", "email", "roles"},
	}

	state := "123"

	http.HandleFunc("/", func(writer http.ResponseWriter, request *http.Request) {
		http.Redirect(writer, request, config.AuthCodeURL(state), http.StatusFound)
	})

	http.HandleFunc("/auth/callback", func(writer http.ResponseWriter, request *http.Request) {
		if request.URL.Query().Get("state") != state {
			http.Error(writer, "state did not match", http.StatusBadRequest)
			return
		}

		token, err := config.Exchange(ctx, request.URL.Query().Get("code"))
		if err != nil {
			http.Error(writer, "Failed to exchange token: "+err.Error(), http.StatusBadRequest)
			return
		}

		idToken, ok := token.Extra("id_token").(string)

		if !ok {
			http.Error(writer, "No id_token field in token", http.StatusBadRequest)
			return
		}

		userInfo, err := provider.UserInfo(ctx, oauth2.StaticTokenSource(token))

		if err != nil {
			http.Error(writer, "Failed to get user info: "+err.Error(), http.StatusBadRequest)
			return
		}

		resp := struct {
			AccessToken *oauth2.Token
			IDToken    string
			UserInfo   *oidc.UserInfo
		}{
			AccessToken: token,
			IDToken: idToken,
			UserInfo: userInfo,
		}

		data, err := json.Marshal(resp)

		if err != nil {
			http.Error(writer, "Failed to encode token to JSON", http.StatusInternalServerError)
			return
		}

		writer.Write(data)
	})

	log.Fatal(http.ListenAndServe(":8081", nil))
}