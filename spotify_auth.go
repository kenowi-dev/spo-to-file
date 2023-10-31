package main

import (
	"context"
	"encoding/base64"
	"github.com/zmb3/spotify/v2"
	spotifyauth "github.com/zmb3/spotify/v2/auth"
	"golang.org/x/oauth2"
	"net/http"
	"os"
)

type SpotifyAuth interface {
	AuthUrl() string
	Callback(r *http.Request) (*spotify.Client, error)
	Client(accessToken string) *spotify.Client
}

type spotifyAuth struct {
	auth  *spotifyauth.Authenticator
	state string
}

func NewSpotifyAuth() SpotifyAuth {
	return &spotifyAuth{
		auth: spotifyauth.New(
			spotifyauth.WithRedirectURL("http://localhost:8000/callback"),
			spotifyauth.WithScopes(
				spotifyauth.ScopeUserReadPrivate,
				spotifyauth.ScopePlaylistReadPrivate,
				spotifyauth.ScopePlaylistReadCollaborative,
				spotifyauth.ScopeUserFollowRead,
				spotifyauth.ScopeUserLibraryRead,
			),
			spotifyauth.WithClientID(os.Getenv("SPOTIFY_ID")),
			spotifyauth.WithClientSecret(os.Getenv("SPOTIFY_SECRET")),
		),
		state: base64.StdEncoding.EncodeToString([]byte(os.Getenv("STATE_SALT"))),
	}
}

func (e *spotifyAuth) AuthUrl() string {
	return e.auth.AuthURL(e.state, spotifyauth.ShowDialog)
}

func (e *spotifyAuth) Callback(r *http.Request) (*spotify.Client, error) {

	tok, err := e.auth.Token(r.Context(), e.state, r)
	if err != nil {
		return nil, err
	}

	return spotify.New(e.auth.Client(r.Context(), tok)), nil
}

func (e *spotifyAuth) Client(accessToken string) *spotify.Client {
	tok := &oauth2.Token{AccessToken: accessToken}
	return spotify.New(e.auth.Client(context.Background(), tok))
}
