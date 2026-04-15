package lastfm

import (
	"context"
	"encoding/base64"
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
	return c.FetchNowPlayingForUser(ctx, c.user)
}

func (c *Client) FetchNowPlayingForUser(ctx context.Context, user string) (*NowPlaying, error) {
	username := strings.TrimSpace(user)
	if username == "" {
		return nil, errors.New("last.fm username is required")
	}

	endpoint := url.URL{
		Scheme: "https",
		Host:   "ws.audioscrobbler.com",
		Path:   "/2.0/",
	}

	query := endpoint.Query()
	query.Set("method", "user.getrecenttracks")
	query.Set("user", username)
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

	return parseTrack(payload.RecentTracks.Track, username)
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

	artwork := fallbackArtworkDataURL()
	for index := len(track.Image) - 1; index >= 0; index-- {
		candidate := strings.TrimSpace(track.Image[index].Text)
		if !isUsableArtworkURL(candidate) {
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

func isUsableArtworkURL(candidate string) bool {
	if candidate == "" {
		return false
	}

	// Last.fm sometimes returns a non-empty placeholder image instead of real artwork.
	return !strings.Contains(strings.ToLower(candidate), "2a96cbd8b46e442fc41c2b86b821562f")
}

func fallbackArtworkDataURL() *string {
	svg := `<svg fill="#000000" viewBox="0 0 32.00 32.00" id="icon" xmlns="http://www.w3.org/2000/svg" stroke="#000000" transform="rotate(0)" stroke-width="0.00032"><g id="SVGRepo_bgCarrier" stroke-width="0"></g><g id="SVGRepo_tracerCarrier" stroke-linecap="round" stroke-linejoin="round" stroke="#CCCCCC" stroke-width="0.704"></g><g id="SVGRepo_iconCarrier"><defs><style>.cls-1{fill:none;}</style></defs><title>unknown--filled</title><path d="M29.4163,14.5906,17.41,2.5842a1.9937,1.9937,0,0,0-2.8191,0L2.5837,14.5906a1.994,1.994,0,0,0,0,2.8193L14.5906,29.4163a1.9937,1.9937,0,0,0,2.8191,0L29.4163,17.41A1.994,1.994,0,0,0,29.4163,14.5906ZM16,24a1.5,1.5,0,1,1,1.5-1.5A1.5,1.5,0,0,1,16,24Zm1.125-6.7519v1.8769h-2.25V15H17a1.875,1.875,0,0,0,0-3.75H15a1.8771,1.8771,0,0,0-1.875,1.875v.5h-2.25v-.5A4.13,4.13,0,0,1,15,9h2a4.125,4.125,0,0,1,.125,8.2481Z"></path><path id="inner-path" class="cls-1" d="M16,21a1.5,1.5,0,1,1-1.5,1.5A1.5,1.5,0,0,1,16,21Zm1.125-3.752A4.1249,4.1249,0,0,0,17,9H15a4.13,4.13,0,0,0-4.125,4.125v.5h2.25v-.5A1.8772,1.8772,0,0,1,15,11.25h2A1.875,1.875,0,0,1,17,15H14.875v4.125h2.25Z"></path><rect id="_Transparent_Rectangle_" data-name="&lt;Transparent Rectangle&gt;" class="cls-1" width="32" height="32"></rect></g></svg>`
	data := "data:image/svg+xml;base64," + base64.StdEncoding.EncodeToString([]byte(svg))
	return &data
}
