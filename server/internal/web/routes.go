package web

import "net/http"

type routeDescriptor struct {
	Path        string   `json:"path"`
	Methods     []string `json:"methods"`
	Description string   `json:"description"`
}

type rootResponse struct {
	Name   string            `json:"name"`
	Routes []routeDescriptor `json:"routes"`
}

func apiRoutes() []routeDescriptor {
	return []routeDescriptor{
		{
			Path:        "/api/lastfm",
			Methods:     []string{http.MethodGet},
			Description: "Returns the current or most recent Last.fm track.",
		},
		{
			Path:        "/api/views/:username",
			Methods:     []string{http.MethodGet, http.MethodPost},
			Description: "Reads or increments the public view counter for a specific profile username.",
		},
	}
}
