package main

import (
	"archive/zip"
	"bytes"
	"encoding/json"
	"github.com/zmb3/spotify/v2"
)

type Zipper interface {
	Json() (*bytes.Buffer, error)
}

type zipper struct {
	tracks    []spotify.SavedTrack
	artists   []spotify.FullArtist
	playlists []PlaylistWithTracks
	albums    []AlbumWithTracks
	user      *spotify.PrivateUser
}

func NewZipper(
	tracks []spotify.SavedTrack,
	artists []spotify.FullArtist,
	playlists []PlaylistWithTracks,
	albums []AlbumWithTracks,
	user *spotify.PrivateUser,
) Zipper {
	return &zipper{
		tracks:    tracks,
		artists:   artists,
		playlists: playlists,
		albums:    albums,
		user:      user,
	}
}

func (z *zipper) Json() (*bytes.Buffer, error) {

	userJson, err := json.MarshalIndent(z.user, "", "  ")
	if err != nil {
		return nil, err
	}

	tracksJson, err := json.MarshalIndent(z.tracks, "", "  ")
	if err != nil {
		return nil, err
	}

	playlistsJson, err := json.MarshalIndent(z.playlists, "", "  ")
	if err != nil {
		return nil, err
	}

	artistsJson, err := json.MarshalIndent(z.artists, "", "  ")
	if err != nil {
		return nil, err
	}

	albumsJson, err := json.MarshalIndent(z.albums, "", "  ")
	if err != nil {
		return nil, err
	}

	buf := new(bytes.Buffer)

	// Create a new zip archive.
	zipW := zip.NewWriter(buf)

	err = z.jsonToZip(userJson, zipW, "user.json")
	if err != nil {
		return nil, err
	}

	err = z.jsonToZip(tracksJson, zipW, "tracks.json")
	if err != nil {
		return nil, err
	}

	err = z.jsonToZip(playlistsJson, zipW, "playlists.json")
	if err != nil {
		return nil, err
	}

	err = z.jsonToZip(artistsJson, zipW, "artists.json")
	if err != nil {
		return nil, err
	}

	err = z.jsonToZip(albumsJson, zipW, "albums.json")
	if err != nil {
		return nil, err
	}

	err = zipW.Close()
	if err != nil {
		return nil, err
	}

	return buf, nil
}

func (z *zipper) jsonToZip(json []byte, w *zip.Writer, name string) error {
	f, err := w.Create(name)
	if err != nil {
		return err
	}

	_, err = f.Write(json)
	if err != nil {
		return err
	}
	return nil
}
