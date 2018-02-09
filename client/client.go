package main

import (
	"context"
	"html/template"
	"log"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/gorilla/sessions"
	negronilogrus "github.com/meatballhat/negroni-logrus"
	"github.com/ory/common/env"
	"github.com/ory/hydra/sdk/go/hydra"
	"github.com/pkg/errors"
	"github.com/urfave/negroni"
	"golang.org/x/oauth2"
)

// This store will be used to save user authentication
var store = sessions.NewCookieStore([]byte("something-very-secret-keep-it-safe"))

// The session is a unique session identifier
const sessionName = "authentication"

// This is the Hydra SDK
var client hydra.SDK

// A state for performing the OAuth 2.0 flow. This is usually not part of a consent app, but in order for the demo
// to make sense, it performs the OAuth 2.0 authorize code flow.
var state = "demostatedemostatedemo2"

func main() {
	var err error

	client, err = hydra.NewSDK(&hydra.Configuration{
		ClientID:     env.Getenv("HYDRA_CLIENT_ID", "client-app"),
		ClientSecret: env.Getenv("HYDRA_CLIENT_SECRET", "client-secret"),
		EndpointURL:  env.Getenv("HYDRA_CLUSTER_URL", "http://localhost:4444"),
		Scopes:       []string{"demo"},
	})
	if err != nil {
		log.Fatalf("Unable to connect to the Hydra SDK because %s", err)
	}

	// Initialize the hydra SDK. The defaults work if you started hydra as described in the README.md
	// Set up a router and some routes
	r := mux.NewRouter()
	r.HandleFunc("/", handleHome)
	r.HandleFunc("/callback", handleCallback)

	// Set up a request logger, useful for debugging
	n := negroni.New()
	n.Use(negronilogrus.NewMiddleware())
	n.UseHandler(r)

	// Start http server
	log.Println("Listening on :" + env.Getenv("PORT", "3500"))
	http.ListenAndServe(":"+env.Getenv("PORT", "3500"), n)
}

// handles request at /home - a small page that let's you know what you can do in this app. Usually the first.
// page a user sees.
func handleHome(w http.ResponseWriter, _ *http.Request) {
	var config = client.GetOAuth2Config()
	config.RedirectURL = "http://localhost:3500/callback"
	config.Scopes = []string{"offline", "openid"}

	var authURL = client.GetOAuth2Config().AuthCodeURL(state) + "&nonce=" + state
	log.Println("AuthURL is " + authURL)
	renderTemplate(w, "home.html", authURL)
}

// Once the user has given their consent, we will hit this endpoint. Again,
// this is not something that would be included in a traditional consent app,
// but we added it so you can see the data once the consent flow is done.
func handleCallback(w http.ResponseWriter, r *http.Request) {
	// in the real world you should check the state query parameter, but this is omitted for brevity reasons.

	// Exchange the access code for an access (and optionally) a refresh token
	token, err := client.GetOAuth2Config().Exchange(context.Background(), r.URL.Query().Get("code"))
	if err != nil {
		http.Error(w, errors.Wrap(err, "Could not exhange token").Error(), http.StatusBadRequest)
		return
	}

	// Render the output
	renderTemplate(w, "callback.html", struct {
		*oauth2.Token
		IDToken interface{}
	}{
		Token:   token,
		IDToken: token.Extra("id_token"),
	})
}

// renderTemplate is a convenience helper for rendering templates.
func renderTemplate(w http.ResponseWriter, id string, d interface{}) bool {
	if t, err := template.New(id).ParseFiles("./templates/" + id); err != nil {
		http.Error(w, errors.Wrap(err, "Could not render template").Error(), http.StatusInternalServerError)
		return false
	} else if err := t.Execute(w, d); err != nil {
		http.Error(w, errors.Wrap(err, "Could not render template").Error(), http.StatusInternalServerError)
		return false
	}
	return true
}
