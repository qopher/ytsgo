// Package ytsgo is a client for YTS.LT API.
// Details can be found at https://yts.lt/api
package ytsgo

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"time"
)

const (
	// DefaultBaseURL is a default URL used for queries.
	DefaultBaseURL = "https://yts.lt/api/v2/"
	statusOK       = "ok"
)

var (
	// DefaultTimeout is a default timeout used for queries.
	DefaultTimeout = time.Second * 10
	urls           = map[string]string{
		"movieURL":       "movie_details.json",
		"listMoviesURL":  "list_movies.json",
		"suggestionsURL": "movie_suggestions.json",
	}
)

// ClientOption modify default behavior of the Client.
type ClientOption func(c *Client)

// BaseURL overrides DefaultBaseURL value.
func BaseURL(url string) ClientOption {
	return func(c *Client) {
		c.baseURLStr = url
	}
}

// HTTPTimeout overrides default HTTP client timeout.
func HTTPTimeout(d time.Duration) ClientOption {
	return func(c *Client) {
		c.httpClient.Timeout = d
	}
}

// UserAgent sets the User-Agent header. By default this header is not set.
func UserAgent(ua string) ClientOption {
	return func(c *Client) {
		c.userAgent = ua
	}
}

// Client implements yts.lt API client.
type Client struct {
	baseURLStr string
	baseURL    *url.URL
	userAgent  string
	httpClient *http.Client
	urls       map[string]*url.URL
}

// New creates a new Client.
func New(opts ...ClientOption) (*Client, error) {
	c := &Client{
		baseURLStr: DefaultBaseURL,
		httpClient: &http.Client{
			Timeout: DefaultTimeout,
		},
		urls: make(map[string]*url.URL),
	}
	for _, o := range opts {
		o(c)
	}
	var err error
	c.baseURL, err = url.Parse(c.baseURLStr)
	if err != nil {
		return nil, err
	}
	for k, u := range urls {
		ur, err := url.Parse(u)
		if err != nil {
			return nil, err
		}
		c.urls[k] = ur
	}
	return c, nil
}

// MovieOption changes the default behavior of Movie call.
type MovieOption func(url.Values)

// MovieWithImages if true will return additional image URLs in the response.
func MovieWithImages(b bool) MovieOption {
	return func(v url.Values) {
		v.Add("with_images", fmt.Sprintf("%v", b))
	}
}

// MovieWithCast if true will return cast information for the movie.
func MovieWithCast(b bool) MovieOption {
	return func(v url.Values) {
		v.Add("with_cast", fmt.Sprintf("%v", b))
	}
}

// Movie returns movie details based on provided ID and options.
func (c *Client) Movie(id int, opts ...MovieOption) (*Movie, error) {
	u := c.baseURL.ResolveReference(c.urls["movieURL"])
	params := u.Query()
	params.Set("movie_id", fmt.Sprintf("%v", id))
	for _, o := range opts {
		o(params)
	}
	req, err := c.newRequest(u, params)
	if err != nil {
		return nil, err
	}
	rsp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer rsp.Body.Close()
	if rsp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("server returned code %v: %s", rsp.StatusCode, rsp.Status)
	}
	var data movieDetailsResponse
	if err := json.NewDecoder(rsp.Body).Decode(&data); err != nil {
		return nil, err
	}
	if data.status.Status != statusOK {
		return nil, fmt.Errorf("api returned incorrect status %s: %s", data.status.Status, data.status.StatusMessage)
	}
	return data.Data.Movie, nil
}

// ListMoviesOption configures behavior of ListMovies. Limits, pages, quality and more can be set.
type ListMoviesOption func(url.Values)

// LMLimit is used limit of results per page that has been set (max 50, default 20).
func LMLimit(l uint) ListMoviesOption {
	return func(v url.Values) {
		if l > 50 {
			l = 50
		}
		v.Set("limit", fmt.Sprintf("%v", l))
	}
}

// LMPage is used to see the next page of movies, eg limit=15 and page=2 will show you movies 15-30.
func LMPage(p uint) ListMoviesOption {
	return func(v url.Values) {
		v.Set("page", fmt.Sprintf("%v", p))
	}
}

// LMQuality is used to filter by a given quality. Possible values: 720p, 1080p, 3D.
func LMQuality(q string) ListMoviesOption {
	return func(v url.Values) {
		v.Set("quality", q)
	}
}

// LMMinimumRating is used to filter movie by a given minimum IMDb rating. Allowed values (0-9).
func LMMinimumRating(r uint) ListMoviesOption {
	return func(v url.Values) {
		if r > 9 {
			r = 9
		}
		v.Set("minimum_rating", fmt.Sprintf("%v", r))
	}
}

// LMSearch is used to for movie search, matching on: Movie Title/IMDb Code, Actor Name/IMDb Code, Director Name/IMDb Code
func LMSearch(q string) ListMoviesOption {
	return func(v url.Values) {
		v.Set("query_term", q)
	}
}

// LMGenre is used to filter by a given genre (See http://www.imdb.com/genre/ for full list).
func LMGenre(g string) ListMoviesOption {
	return func(v url.Values) {
		v.Set("genre", g)
	}
}

// LMSortBy sorts the results by choosen value. Possible values (title, year, rating, peers, seeds, download_count, like_count, date_added).
func LMSortBy(s string) ListMoviesOption {
	return func(v url.Values) {
		v.Set("sort_by", s)
	}
}

// LMOrderBy orders the results by either Ascending (asc) or Descending (desc) order
func LMOrderBy(s string) ListMoviesOption {
	return func(v url.Values) {
		v.Set("order_by", s)
	}
}

// Movies contain data returned by ListMovies and MovieSuggestions.
type Movies struct {
	// MovieCount is a total movie count results for your query.
	MovieCount uint `json:"movie_count"`
	// Page is a current page number you are viewing.
	Page uint `json:"page_number"`
	// Limit of results per page that has been set.
	Limit  uint     `json:"limit"`
	Movies []*Movie `json:"movies"`
}

// ListMovies is used to list and search through out all the available movies. Can sort, filter, search and order the results.
func (c *Client) ListMovies(opts ...ListMoviesOption) (*Movies, error) {
	u := c.baseURL.ResolveReference(c.urls["listMoviesURL"])
	params := u.Query()
	for _, o := range opts {
		o(params)
	}
	req, err := c.newRequest(u, params)
	if err != nil {
		return nil, err
	}
	rsp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer rsp.Body.Close()
	if rsp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("server returned code %v: %s", rsp.StatusCode, rsp.Status)
	}
	var data listMoviesResponse
	if err := json.NewDecoder(rsp.Body).Decode(&data); err != nil {
		return nil, err
	}
	if data.status.Status != statusOK {
		return nil, fmt.Errorf("api returned incorrect status %s: %s", data.status.Status, data.status.StatusMessage)
	}
	return data.Data, nil
}

// Suggestions returns 4 related movies as suggestions for the user.
func (c *Client) Suggestions(id int) ([]*Movie, error) {
	u := c.baseURL.ResolveReference(c.urls["suggestionsURL"])
	params := u.Query()
	params.Set("movie_id", fmt.Sprintf("%v", id))
	req, err := c.newRequest(u, params)
	if err != nil {
		return nil, err
	}
	rsp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer rsp.Body.Close()
	if rsp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("server returned code %v: %s", rsp.StatusCode, rsp.Status)
	}
	var data suggestionsResponse
	if err := json.NewDecoder(rsp.Body).Decode(&data); err != nil {
		return nil, err
	}
	if data.status.Status != statusOK {
		return nil, fmt.Errorf("api returned incorrect status %s: %s", data.status.Status, data.status.StatusMessage)
	}
	return data.Data.Movies, nil
}

func (c *Client) newRequest(u *url.URL, params url.Values) (*http.Request, error) {
	u.RawQuery = params.Encode()
	req, err := http.NewRequest("GET", u.String(), nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Accept", "application/json")
	if c.userAgent != "" {
		req.Header.Set("User-Agent", c.userAgent)
	}
	return req, nil
}

type status struct {
	Status        string `json:"status"`
	StatusMessage string `json:"status_message"`
}

type movieDetailsData struct {
	Movie *Movie `json:"movie"`
}

type movieDetailsResponse struct {
	status
	Data movieDetailsData `json:"data"`
}

type listMoviesResponse struct {
	status
	Data *Movies `json:"data"`
}

type suggestionsData struct {
	MovieCount uint     `json:"movie_count"`
	Movies     []*Movie `json:"movies"`
}

type suggestionsResponse struct {
	status
	Data suggestionsData `json:"data"`
}
