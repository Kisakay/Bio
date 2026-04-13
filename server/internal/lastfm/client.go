package lastfm

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strings"
)

type Client struct {
	apiKey     string
	user       string
	httpClient *http.Client
}

type NowPlaying struct {
	Title     string  `json:"title"`
	Artist    string  `json:"artist"`
	Artwork   *string `json:"artwork"`
	Timestamp string  `json:"timestamp"`
	URL       string  `json:"url"`
	IsLive    bool    `json:"isLive"`
}

type recentTracksResponse struct {
	RecentTracks struct {
		Track json.RawMessage `json:"track"`
	} `json:"recenttracks"`
}

type trackPayload struct {
	Name   string            `json:"name"`
	URL    string            `json:"url"`
	Artist textPayload       `json:"artist"`
	Image  []textPayload     `json:"image"`
	Attr   nowPlayingPayload `json:"@attr"`
	Date   datePayload       `json:"date"`
}

type textPayload struct {
	Text string `json:"#text"`
	Name string `json:"name"`
}

type nowPlayingPayload struct {
	NowPlaying string `json:"nowplaying"`
}

type datePayload struct {
	Text string `json:"#text"`
}

func NewClient(apiKey, user string, httpClient *http.Client) *Client {
	return &Client{
		apiKey:     strings.TrimSpace(apiKey),
		user:       strings.TrimSpace(user),
		httpClient: httpClient,
	}
}

func (c *Client) HasCredentials() bool {
	return c.apiKey != ""
}

func (c *Client) FetchNowPlaying(ctx context.Context) (*NowPlaying, error) {
	endpoint := url.URL{
		Scheme: "https",
		Host:   "ws.audioscrobbler.com",
		Path:   "/2.0/",
	}

	query := endpoint.Query()
	query.Set("method", "user.getrecenttracks")
	query.Set("user", c.user)
	query.Set("api_key", c.apiKey)
	query.Set("format", "json")
	query.Set("limit", "1")
	query.Set("extended", "1")
	endpoint.RawQuery = query.Encode()

	request, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint.String(), nil)
	if err != nil {
		return nil, err
	}

	response, err := c.httpClient.Do(request)
	if err != nil {
		return nil, err
	}
	defer response.Body.Close()

	if response.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("last.fm returned %d", response.StatusCode)
	}

	var payload recentTracksResponse
	if err := json.NewDecoder(response.Body).Decode(&payload); err != nil {
		return nil, err
	}

	return parseTrack(payload.RecentTracks.Track, c.user)
}

func parseTrack(raw json.RawMessage, fallbackUser string) (*NowPlaying, error) {
	if len(raw) == 0 {
		return nil, nil
	}

	var tracks []trackPayload
	if err := json.Unmarshal(raw, &tracks); err != nil {
		var single trackPayload
		if errSingle := json.Unmarshal(raw, &single); errSingle != nil {
			return nil, errors.New("unable to decode last.fm track")
		}
		tracks = []trackPayload{single}
	}

	if len(tracks) == 0 {
		return nil, nil
	}

	track := tracks[0]
	title := strings.TrimSpace(track.Name)
	artist := strings.TrimSpace(firstNonEmpty(track.Artist.Name, track.Artist.Text))
	if title == "" || artist == "" {
		return nil, errors.New("last.fm track payload missing title or artist")
	}

	var artwork *string
	for index := len(track.Image) - 1; index >= 0; index-- {
		candidate := strings.TrimSpace(track.Image[index].Text)
		if candidate == "" {
			continue
		}

		artwork = &candidate
		break
	}

	isLive := strings.TrimSpace(track.Attr.NowPlaying) == "true"
	timestamp := strings.TrimSpace(track.Date.Text)
	if timestamp == "" {
		timestamp = "recent scrobble"
	}
	if isLive {
		timestamp = "live now"
	}

	trackURL := strings.TrimSpace(track.URL)
	if trackURL != "" && !strings.HasPrefix(trackURL, "http") {
		trackURL = "https://www.last.fm" + trackURL
	}
	if trackURL == "" {
		trackURL = "https://www.last.fm/user/" + fallbackUser
	}

	return &NowPlaying{
		Title:     title,
		Artist:    artist,
		Artwork:   artwork,
		Timestamp: timestamp,
		URL:       trackURL,
		IsLive:    isLive,
	}, nil
}

func firstNonEmpty(values ...string) string {
	for _, value := range values {
		trimmed := strings.TrimSpace(value)
		if trimmed != "" {
			return trimmed
		}
	}

	return ""
}
