package ytsgo

// File movie.go contains data structures and methods for movie parsing.

import (
	"encoding/json"
	"fmt"
	"net/url"
	"time"
)

// Movie contains information about a single movie from YTS.LT.
type Movie struct {
	ID                      uint       `json:"id"`
	URL                     *url.URL   `json:"-"`
	IMDBCode                string     `json:"imdb_code"`
	Title                   string     `json:"title"`
	TitleEnglish            string     `json:"title_english"`
	Slug                    string     `json:"slug"`
	Year                    uint       `json:"year"`
	Rating                  float32    `json:"rating"`
	Runtime                 uint       `json:"runtime"`
	Genres                  []string   `json:"genres"`
	DownloadCount           uint       `json:"download_count"`
	LikeCound               uint       `json:"like_count"`
	DescriptionIntro        string     `json:"description_intro"`
	DescriptionFull         string     `json:"description_full"`
	YouTubeTrailerCode      string     `json:"yt_trailer_code"`
	Language                string     `json:"language"`
	MPARating               string     `json:"mpa_rating"`
	BackgroundImage         *url.URL   `json:"-"`
	BackgroundImageOriginal *url.URL   `json:"-"`
	SmallCoverImage         *url.URL   `json:"-"`
	MediumCoverImage        *url.URL   `json:"-"`
	LargeCoverImage         *url.URL   `json:"-"`
	DateUploaded            time.Time  `json:"-"`
	DateUploadedUnix        int64      `json:"date_uploaded_unix"`
	Torrents                []*Torrent `json:"torrents"`
	Cast                    []*Cast    `json:"cast"`
}

// UnmarshalJSON unmarshals movie encoded as JSON.
func (m *Movie) UnmarshalJSON(data []byte) error {
	type mov Movie
	aux := &struct {
		URLRaw       string `json:"url"`
		BGImgURL     string `json:"background_image"`
		BGImgURLOrig string `json:"background_image_original"`
		SCoverImg    string `json:"small_cover_image"`
		MCoverImg    string `json:"medium_cover_image"`
		LCoverImg    string `json:"large_cover_image"`
		*mov
	}{
		mov: (*mov)(m),
	}
	if err := json.Unmarshal(data, aux); err != nil {
		return err
	}
	urls := []struct {
		dest **url.URL
		str  string
	}{
		{dest: &m.URL, str: aux.URLRaw},
		{dest: &m.BackgroundImage, str: aux.BGImgURL},
		{dest: &m.BackgroundImageOriginal, str: aux.BGImgURLOrig},
		{dest: &m.SmallCoverImage, str: aux.SCoverImg},
		{dest: &m.MediumCoverImage, str: aux.MCoverImg},
		{dest: &m.LargeCoverImage, str: aux.LCoverImg},
	}
	for _, u := range urls {
		if err := parseURL(u.dest, u.str); err != nil {
			return err
		}
	}
	parseTime(&m.DateUploaded, m.DateUploadedUnix)
	for _, t := range m.Torrents {
		t.movieName = m.Title
	}
	return nil
}

// Torrent contains information about torrent associated with the movie.
type Torrent struct {
	URL              *url.URL  `json:"-"`
	Hash             string    `json:"hash"`
	Quality          string    `json:"quality"`
	Type             string    `json:"type"`
	Seeds            uint      `json:"seeds"`
	Peers            uint      `json:"peers"`
	Size             string    `json:"size"`
	SizeBytes        uint      `json:"size_bytes"`
	DateUploaded     time.Time `json:"-"`
	DateUploadedUnix int64     `json:"date_uploaded_unix"`
	movieName        string
}

// UnmarshalJSON unmarshals Torrent encoded as JSON.
func (t *Torrent) UnmarshalJSON(data []byte) error {
	type tor Torrent
	aux := &struct {
		URLRaw string `json:"url"`
		*tor
	}{
		tor: (*tor)(t),
	}
	if err := json.Unmarshal(data, aux); err != nil {
		return err
	}
	if err := parseURL(&t.URL, aux.URLRaw); err != nil {
		return err
	}
	parseTime(&t.DateUploaded, t.DateUploadedUnix)
	return nil
}

// DefaultTackers is a default, recommended list of trackers.
var DefaultTackers = []string{
	"udp://open.demonii.com:1337/announce",
	"udp://tracker.openbittorrent.com:80",
	"udp://tracker.coppersurfer.tk:6969",
	"udp://glotorrents.pw:6969/announce",
	"udp://tracker.opentrackr.org:1337/announce",
	"udp://torrent.gresille.org:80/announce",
	"udp://p4p.arenabg.com:1337",
	"udp://tracker.leechers-paradise.org:6969",
}

// Magnet returns magnet link for provided torrent. If trackers are passed they will replace list of DefaultTrackers.
func (t *Torrent) Magnet(trackers ...string) string {
	v := url.Values{}
	v.Set("dn", t.movieName)
	trckrs := DefaultTackers
	if len(trackers) > 0 {
		trckrs = trackers
	}
	for _, tr := range trckrs {
		v.Add("tr", tr)
	}
	return fmt.Sprintf("magnet:?xt=urn:btih:%s&%s", t.Hash, v.Encode())
}

// Cast contais information about actors plaing in the movie.
type Cast struct {
	Name          string   `json:"name"`
	CharacterName string   `json:"character_name"`
	IMDBCode      string   `json:"imdb_code"`
	URLSmallImage *url.URL `json:"-"`
}

// UnmarshalJSON unmarshals Cast encoded as JSON.
func (c *Cast) UnmarshalJSON(data []byte) error {
	type cst Cast
	aux := &struct {
		SmallImageURL string `json:"url_small_image"`
		*cst
	}{
		cst: (*cst)(c),
	}
	if err := json.Unmarshal(data, aux); err != nil {
		return err
	}
	if err := parseURL(&c.URLSmallImage, aux.SmallImageURL); err != nil {
		return err
	}
	return nil
}

type TorrentsBySize []*Torrent

func (t TorrentsBySize) Len() int           { return len(t) }
func (t TorrentsBySize) Swap(i, j int)      { t[i], t[j] = t[j], t[i] }
func (t TorrentsBySize) Less(i, j int) bool { return t[i].SizeBytes < t[j].SizeBytes }

type TorrentsBySeeds []*Torrent

func (t TorrentsBySeeds) Len() int           { return len(t) }
func (t TorrentsBySeeds) Swap(i, j int)      { t[i], t[j] = t[j], t[i] }
func (t TorrentsBySeeds) Less(i, j int) bool { return t[i].Seeds < t[j].Seeds }

func parseTime(dest *time.Time, unix int64) {
	*dest = time.Unix(unix, 0)
}

func parseURL(dest **url.URL, str string) error {
	u, err := url.Parse(str)
	if err != nil {
		return err
	}
	*dest = u
	return nil
}
