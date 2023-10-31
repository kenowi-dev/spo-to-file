package main

import (
	"fmt"
	"log"
	"net/http"
	"time"
)

const CookieName = "spo-to-file"

type Router interface {
	SetupAndRun()
}

type router struct {
	spotifyAuth SpotifyAuth
}

func InitRouter(auth SpotifyAuth) Router {
	return &router{
		spotifyAuth: auth,
	}
}

func (ro *router) SetupAndRun() {

	http.HandleFunc("/", ro.index)

	http.HandleFunc("/login", ro.login)
	http.HandleFunc("/callback", ro.callback)

	http.HandleFunc("/download", ro.backup)
	http.HandleFunc("/download/yaml", ro.downloadJson)
	http.HandleFunc("/download/json", ro.downloadJson)

	log.Println("App running on 8000...")
	log.Fatal(http.ListenAndServe(":8000", nil))
}

func (ro *router) index(w http.ResponseWriter, r *http.Request) {
	http.ServeFile(w, r, "./index.html")
}

func (ro *router) login(w http.ResponseWriter, r *http.Request) {
	w.Header().Add("HX-Redirect", ro.spotifyAuth.AuthUrl())
}

func (ro *router) callback(w http.ResponseWriter, r *http.Request) {

	client, err := ro.spotifyAuth.Callback(r)
	if err != nil {
		log.Println(err)
		http.Error(w, err.Error(), http.StatusForbidden)
	}

	tok, err := client.Token()
	if err != nil {
		log.Println(err)
		http.Error(w, err.Error(), http.StatusForbidden)
	}

	cookie := http.Cookie{
		Name:     CookieName,
		Value:    tok.AccessToken,
		Expires:  time.Time{},
		MaxAge:   3600,
		Secure:   true,
		HttpOnly: true,
	}

	http.SetCookie(w, &cookie)
	http.Redirect(w, r, "/download", http.StatusSeeOther)
}

func (ro *router) backup(w http.ResponseWriter, r *http.Request) {
	http.ServeFile(w, r, "./download.html")
}

func (ro *router) downloadJson(w http.ResponseWriter, r *http.Request) {

	cookie, err := r.Cookie(CookieName)
	if err != nil {
		return
	}

	client := ro.spotifyAuth.Client(cookie.Value)

	getter := NewSpotifyGetter(client)

	user, err := getter.PrivateUser()
	if err != nil {
		log.Println(err)
		http.Redirect(w, r, "/", http.StatusInternalServerError)
	}

	tracks, err := getter.SavedTacks()
	if err != nil {
		log.Println(err)
		http.Redirect(w, r, "/", http.StatusInternalServerError)
	}

	artists, err := getter.FollowedArtists()
	if err != nil {
		log.Println(err)
		http.Redirect(w, r, "/", http.StatusInternalServerError)
	}

	playlists, err := getter.Playlists()
	if err != nil {
		log.Println(err)
		http.Redirect(w, r, "/", http.StatusInternalServerError)
	}

	albums, err := getter.SavedAlbums()
	if err != nil {
		log.Println(err)
		http.Redirect(w, r, "/", http.StatusInternalServerError)
	}

	zipper := NewZipper(tracks, artists, playlists, albums, user)

	buf, err := zipper.Json()
	if err != nil {
		log.Println(err)
		http.Redirect(w, r, "/", http.StatusInternalServerError)
	}

	w.Header().Set("Content-Type", "application/zip")
	w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=\"%s.zip\"", "library"))
	_, err = w.Write(buf.Bytes())
	if err != nil {
		log.Println(err)
	}
}
