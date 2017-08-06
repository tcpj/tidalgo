package tidalgo

import (
	"net/http"
	"net/url"
)

type Session struct {
	Id          string
	UserId      int
	CountryCode string
	BaseUrl     *url.URL
	Client      *http.Client
}

type Artist struct {
	Id   int    `json:"id"`
	Name string `json:"name"`
}

type Album struct {
	AllowStreaming bool     `json:"allowStreaming"`
	Artist         Artist   `json:"artist"`
	Artists        []Artist `json:"artists"`
	Id             int      `json:"id"`
	NumberOfTracks int      `json:"numberOfTracks"`
	ReleaseDate    string   `json:"releaseDate"`
	Title          string   `json:"title"`
	Type           string   `json:"type"`
}

type Track struct {
	Album       Album    `json:"album"`
	Artist      Artist   `json:"artist"`
	Artists     []Artist `json:"artists"`
	Duration    int      `json:"duration"`
	Id          int      `json:"id"`
	Title       string   `json:"title"`
	Version     string   `json:"version"`
	TrackNumber int      `json:"trackNumber"`
}

type FavoriteResponse struct {
	Limit              int `json:"limit"`
	Offset             int `json:"offset"`
	TotalNumberOfItems int `json:"totalNumberOfItems"`
}

type FavoriteAlbumsResponse struct {
	FavoriteResponse
	Items []struct {
		Item Album  `json:"item"`
		Type string `json:"type"`
	} `json:"items"`
}

type FavoriteArtistsResponse struct {
	FavoriteResponse
	Items []struct {
		Item Artist `json:"item"`
		Type string `json:"type"`
	} `json:"items"`
}

type FavoriteTracksResponse struct {
	FavoriteResponse
	Items []struct {
		Item Track  `json:"item"`
		Type string `json:"type"`
	} `json:"items"`
}
