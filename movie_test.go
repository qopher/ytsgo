package ytsgo

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
)

func TestUnMarshalMovie(t *testing.T) {
	tMovie := loadTestData("movie.json", t)
	testData := []struct {
		desc    string
		data    []byte
		want    *Movie
		wantErr bool
	}{
		{
			desc: "success",
			data: tMovie,
			want: &Movie{
				ID:           10,
				URL:          mustURL("https://yts.lt/movie/13-2010", t),
				IMDBCode:     "tt0798817",
				Title:        "13",
				TitleEnglish: "13",
				Slug:         "13-2010",
				Year:         2010,
				Rating:       6.1,
				Runtime:      91,
				Genres: []string{
					"Action",
					"Drama",
					"Thriller",
				},
				DownloadCount:           235099,
				LikeCound:               254,
				DescriptionIntro:        "In Talbot, Ohio, a father's need for surgeries puts the family in a financial bind. His son Vince, an electrician, overhears a man talking about making a fortune in just a day. When the man overdoses on drugs, Vince finds instructions and a cell phone that the man has received and substitutes himself: taking a train to New York and awaiting contact. He has no idea what it's about. He ends up at a remote house where wealthy men bet on who will survive a complicated game of Russian roulette: he's number 13. In flashbacks we meet other contestants, including a man whose brother takes him out of a mental institution in order to compete. Can Vince be the last one standing?",
				DescriptionFull:         "In Talbot, Ohio, a father's need for surgeries puts the family in a financial bind. His son Vince, an electrician, overhears a man talking about making a fortune in just a day. When the man overdoses on drugs, Vince finds instructions and a cell phone that the man has received and substitutes himself: taking a train to New York and awaiting contact. He has no idea what it's about. He ends up at a remote house where wealthy men bet on who will survive a complicated game of Russian roulette: he's number 13. In flashbacks we meet other contestants, including a man whose brother takes him out of a mental institution in order to compete. Can Vince be the last one standing?",
				YouTubeTrailerCode:      "Y41fFj-P4jI",
				Language:                "English",
				MPARating:               "R",
				BackgroundImage:         mustURL("https://yts.lt/assets/images/movies/13_2010/background.jpg", t),
				BackgroundImageOriginal: mustURL("https://yts.lt/assets/images/movies/13_2010/background.jpg", t),
				SmallCoverImage:         mustURL("https://yts.lt/assets/images/movies/13_2010/small-cover.jpg", t),
				MediumCoverImage:        mustURL("https://yts.lt/assets/images/movies/13_2010/medium-cover.jpg", t),
				LargeCoverImage:         mustURL("https://yts.lt/assets/images/movies/13_2010/large-cover.jpg", t),
				DateUploaded:            time.Unix(1446320797, 0),
				DateUploadedUnix:        1446320797,
				Torrents: []*Torrent{{
					URL:              mustURL("https://yts.lt/torrent/download/BE046ED20B048C4FB86E15838DD69DADB27C5E8A", t),
					Hash:             "BE046ED20B048C4FB86E15838DD69DADB27C5E8A",
					Quality:          "720p",
					Type:             "bluray",
					Seeds:            19,
					Peers:            3,
					Size:             "946.49 MB",
					SizeBytes:        992466698,
					DateUploaded:     time.Unix(1446320797, 0),
					DateUploadedUnix: 1446320797,
					movieName:        "13",
				}},
				Cast: []*Cast{
					{
						Name:          "Jason Statham",
						CharacterName: "Jasper",
						URLSmallImage: mustURL("https://yts.lt/assets/images/actors/thumb/nm0005458.jpg", t),
						IMDBCode:      "0005458",
					},
					{
						Name:          "Michael Shannon",
						CharacterName: "Henry",
						URLSmallImage: mustURL("https://yts.lt/assets/images/actors/thumb/nm0788335.jpg", t),
						IMDBCode:      "0788335",
					},
					{
						Name:          "Alexander Skarsgård",
						CharacterName: "Jack",
						URLSmallImage: mustURL("https://yts.lt/assets/images/actors/thumb/nm0002907.jpg", t),
						IMDBCode:      "0002907",
					},
					{
						Name:          "Gaby Hoffmann",
						CharacterName: "Clara Ferro",
						URLSmallImage: mustURL("https://yts.lt/assets/images/actors/thumb/nm0000451.jpg", t),
						IMDBCode:      "0000451",
					},
				},
			},
		},
		{
			desc:    "no data",
			wantErr: true,
		},
		{
			desc:    "bad URL",
			data:    []byte(`{"url": ":"}`),
			wantErr: true,
		},
		{
			desc:    "bad time",
			data:    []byte(`{"date_uploaded_unix": 232.1313}`),
			wantErr: true,
		},
	}
	for _, tc := range testData {
		t.Run(tc.desc, func(t *testing.T) {
			m := &Movie{}
			err := json.Unmarshal(tc.data, m)
			if (err != nil) != tc.wantErr {
				t.Errorf("unexpected error: %v, want: %v", err, tc.wantErr)
			}
			if err != nil {
				return
			}
			if diff := cmp.Diff(tc.want, m, cmp.AllowUnexported(Torrent{})); diff != "" {
				t.Errorf("unexpected results, diff -want +got\n%s", diff)
			}
		})
	}
}

func TestMagnet(t *testing.T) {
	testData := []struct {
		desc     string
		trackers []string
		want     string
	}{
		{
			desc: "default trackers",
			want: "magnet:?xt=urn:btih:HASH123&dn=Name+of+cool+movie&tr=udp%3A%2F%2Fopen.demonii.com%3A1337%2Fannounce&tr=udp%3A%2F%2Ftracker.openbittorrent.com%3A80&tr=udp%3A%2F%2Ftracker.coppersurfer.tk%3A6969&tr=udp%3A%2F%2Fglotorrents.pw%3A6969%2Fannounce&tr=udp%3A%2F%2Ftracker.opentrackr.org%3A1337%2Fannounce&tr=udp%3A%2F%2Ftorrent.gresille.org%3A80%2Fannounce&tr=udp%3A%2F%2Fp4p.arenabg.com%3A1337&tr=udp%3A%2F%2Ftracker.leechers-paradise.org%3A6969",
		},
		{
			desc:     "custom trackers",
			trackers: []string{"udp://tracker1.com:1234", "http://tracker2.com:5678"},
			want:     "magnet:?xt=urn:btih:HASH123&dn=Name+of+cool+movie&tr=udp%3A%2F%2Ftracker1.com%3A1234&tr=http%3A%2F%2Ftracker2.com%3A5678",
		},
	}
	testMovie := &Movie{
		Torrents: []*Torrent{{
			movieName: "Name of cool movie",
			Hash:      "HASH123",
		}},
	}
	for _, tc := range testData {
		t.Run(tc.desc, func(t *testing.T) {
			got := testMovie.Torrents[0].Magnet(tc.trackers...)
			if got != tc.want {
				t.Errorf("Unexpected manget, got:\n%q\nwant:\n%q", got, tc.want)
			}
		})
	}
}
