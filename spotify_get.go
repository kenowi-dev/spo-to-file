package main

import (
	"context"
	"errors"
	"github.com/zmb3/spotify/v2"
	"sync"
)

type SpotifyGetter interface {
	SavedTacks() ([]spotify.SavedTrack, error)
	FollowedArtists() ([]spotify.FullArtist, error)
	Playlists() ([]PlaylistWithTracks, error)
	SavedAlbums() ([]AlbumWithTracks, error)
	PrivateUser() (*spotify.PrivateUser, error)
	Library() (*SpotifyLibrary, error)
}

type PlaylistWithTracks struct {
	spotify.FullPlaylist
	Items []spotify.PlaylistItem
}

type AlbumWithTracks struct {
	spotify.FullAlbum
	Items []spotify.SimpleTrack
}

type SpotifyLibrary struct {
	User      *spotify.PrivateUser
	Tracks    []spotify.SavedTrack
	Artists   []spotify.FullArtist
	Playlists []PlaylistWithTracks
	Albums    []AlbumWithTracks
}

type spotifyGetter struct {
	client *spotify.Client
}

func NewSpotifyGetter(client *spotify.Client) SpotifyGetter {
	return &spotifyGetter{
		client: client,
	}
}

func (s *spotifyGetter) Library() (*SpotifyLibrary, error) {

	wg := sync.WaitGroup{}
	library := SpotifyLibrary{}
	var userErr, tracksErr, artistsErr, playlistsErr, albumsErr error

	wg.Add(1)
	go func() {
		library.User, userErr = s.PrivateUser()
		wg.Done()
	}()

	wg.Add(1)
	go func() {
		library.Tracks, tracksErr = s.SavedTacks()
		wg.Done()
	}()

	wg.Add(1)
	go func() {
		library.Artists, artistsErr = s.FollowedArtists()
		wg.Done()
	}()

	wg.Add(1)
	go func() {
		library.Playlists, playlistsErr = s.Playlists()
		wg.Done()
	}()

	wg.Add(1)
	go func() {
		library.Albums, albumsErr = s.SavedAlbums()
		wg.Done()
	}()

	wg.Wait()

	return &library, errors.Join(userErr, tracksErr, artistsErr, playlistsErr, albumsErr)
}

func (s *spotifyGetter) PrivateUser() (*spotify.PrivateUser, error) {
	return s.client.CurrentUser(context.Background())
}

func (s *spotifyGetter) SavedTacks() ([]spotify.SavedTrack, error) {
	ctx := context.Background()

	allTracks := make([]spotify.SavedTrack, 0)

	tracks, err := s.client.CurrentUsersTracks(ctx, spotify.Limit(50))
	if err != nil {
		return nil, err
	}
	allTracks = append(allTracks, tracks.Tracks...)

	for {
		err = s.client.NextPage(ctx, tracks)
		if errors.Is(err, spotify.ErrNoMorePages) {
			break
		}
		if err != nil {
			return nil, err
		}
		allTracks = append(allTracks, tracks.Tracks...)
	}

	return allTracks, nil
}

func (s *spotifyGetter) FollowedArtists() ([]spotify.FullArtist, error) {
	ctx := context.Background()

	allArtists := make([]spotify.FullArtist, 0)
	after := ""

	for {

		opts := []spotify.RequestOption{spotify.Limit(50)}

		if after != "" {
			opts = append(opts, spotify.After(after))
		}

		artists, err := s.client.CurrentUsersFollowedArtists(
			ctx,
			opts...,
		)
		if err != nil {
			return nil, err
		}
		allArtists = append(allArtists, artists.Artists...)

		after = artists.Cursor.After
		if after == "" {
			break
		}

	}

	return allArtists, nil
}

func (s *spotifyGetter) Playlists() ([]PlaylistWithTracks, error) {
	ctx := context.Background()

	allPlaylists := make([]PlaylistWithTracks, 0)

	playlists, err := s.client.CurrentUsersPlaylists(ctx, spotify.Limit(50))
	if err != nil {
		return nil, err
	}

	for {

		for _, playlist := range playlists.Playlists {
			completePlaylist, err := s.completePlaylist(playlist.ID)
			if err != nil {
				return nil, err
			}
			allPlaylists = append(allPlaylists, *completePlaylist)
		}

		err = s.client.NextPage(ctx, playlists)
		if errors.Is(err, spotify.ErrNoMorePages) {
			break
		}
		if err != nil {
			return nil, err
		}
	}

	return allPlaylists, nil
}

func (s *spotifyGetter) SavedAlbums() ([]AlbumWithTracks, error) {
	ctx := context.Background()

	allAlbums := make([]AlbumWithTracks, 0)

	albums, err := s.client.CurrentUsersAlbums(ctx, spotify.Limit(50))
	if err != nil {
		return nil, err
	}

	for {

		for _, album := range albums.Albums {
			
			completeAlbum, err := s.completeAlbum(album.ID)
			if err != nil {
				return nil, err
			}
			allAlbums = append(allAlbums, *completeAlbum)
		}

		err = s.client.NextPage(ctx, albums)
		if errors.Is(err, spotify.ErrNoMorePages) {
			break
		}
		if err != nil {
			return nil, err
		}
	}

	return allAlbums, nil
}

func (s *spotifyGetter) completePlaylist(id spotify.ID) (*PlaylistWithTracks, error) {
	ctx := context.Background()

	playlist, err := s.client.GetPlaylist(ctx, id)
	if err != nil {
		return nil, err
	}

	playlistItems := make([]spotify.PlaylistItem, 0)

	items, err := s.client.GetPlaylistItems(ctx, id, spotify.Limit(50))
	if err != nil {
		return nil, err
	}

	playlistItems = append(playlistItems, items.Items...)

	for {
		err = s.client.NextPage(ctx, items)
		if errors.Is(err, spotify.ErrNoMorePages) {
			break
		}
		if err != nil {
			return nil, err
		}

		playlistItems = append(playlistItems, items.Items...)
	}

	playlistWithTracks := PlaylistWithTracks{
		FullPlaylist: *playlist,
		Items:        playlistItems,
	}

	return &playlistWithTracks, nil
}

func (s *spotifyGetter) completeAlbum(id spotify.ID) (*AlbumWithTracks, error) {
	ctx := context.Background()

	album, err := s.client.GetAlbum(ctx, id)
	if err != nil {
		return nil, err
	}

	albumTracks := make([]spotify.SimpleTrack, 0)

	tracks, err := s.client.GetAlbumTracks(ctx, id, spotify.Limit(50))
	if err != nil {
		return nil, err
	}

	albumTracks = append(albumTracks, tracks.Tracks...)

	for {
		err = s.client.NextPage(ctx, tracks)
		if errors.Is(err, spotify.ErrNoMorePages) {
			break
		}
		if err != nil {
			return nil, err
		}

		albumTracks = append(albumTracks, tracks.Tracks...)
	}

	playlistWithTracks := AlbumWithTracks{
		FullAlbum: *album,
		Items:     albumTracks,
	}

	return &playlistWithTracks, nil
}
