package main

func main() {

	spotifyAuth := NewSpotifyAuth()

	router := InitRouter(spotifyAuth)

	router.SetupAndRun()
}
