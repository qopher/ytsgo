package ytsgo

import (
	"errors"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"net/url"
	"path/filepath"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
)

func loadTestData(file string, t *testing.T) []byte {
	t.Helper()
	path := filepath.Join("testdata", file)
	bytes, err := ioutil.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	return bytes
}

func mustURL(str string, t *testing.T) *url.URL {
	t.Helper()
	u, err := url.Parse(str)
	if err != nil {
		t.Fatal(err)
	}
	return u
}

type fakeYTSServer struct {
	err  error
	data []byte
	req  *http.Request
}

func (f *fakeYTSServer) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if f.err != nil {
		http.Error(w, f.err.Error(), http.StatusInternalServerError)
		return
	}
	f.req = r
	w.Write(f.data)
}

func TestMovie(t *testing.T) {
	testData := []struct {
		desc      string
		id        int
		opts      []MovieOption
		respFile  string
		err       error
		wantQuery url.Values
		wantErr   bool
	}{
		{
			desc:     "no options",
			id:       1,
			respFile: "matrix.json",
			wantQuery: func() url.Values {
				v := url.Values{}
				v.Set("movie_id", "1")
				return v
			}(),
		},
		{
			desc:     "with images",
			id:       1,
			opts:     []MovieOption{MovieWithImages(true)},
			respFile: "matrix.json",
			wantQuery: func() url.Values {
				v := url.Values{}
				v.Set("movie_id", "1")
				v.Set("with_images", "true")
				return v
			}(),
		},
		{
			desc:     "with cast",
			id:       1,
			opts:     []MovieOption{MovieWithCast(true)},
			respFile: "matrix.json",
			wantQuery: func() url.Values {
				v := url.Values{}
				v.Set("movie_id", "1")
				v.Set("with_cast", "true")
				return v
			}(),
		},
		{
			desc:     "unmarshal error",
			id:       1,
			respFile: "bad_json.json",
			wantErr:  true,
		},
		{
			desc:     "API error",
			id:       1,
			respFile: "error.json",
			wantErr:  true,
		},
		{
			desc:     "error",
			id:       1,
			respFile: "matrix.json",
			err:      errors.New("some error"),
			wantErr:  true,
		},
	}
	f := &fakeYTSServer{
		data: loadTestData("matrix.json", t),
	}
	ts := httptest.NewServer(f)
	defer ts.Close()
	c, err := New(BaseURL(ts.URL), HTTPTimeout(time.Second*5), UserAgent("test"))
	if err != nil {
		t.Fatalf("Failed to connect to test server: %v", err)
	}
	for _, tc := range testData {
		t.Run(tc.desc, func(t *testing.T) {
			f.err = tc.err
			f.data = loadTestData(tc.respFile, t)

			_, err := c.Movie(tc.id, tc.opts...)
			if (err != nil) != tc.wantErr {
				t.Errorf("Unexpected error, got %v want %v", err, tc.wantErr)
			}
			if err != nil {
				return
			}
			if diff := cmp.Diff(tc.wantQuery, f.req.URL.Query()); diff != "" {
				t.Errorf("Unexpected query, diff -want +got\n%s", diff)
			}
			if got, want := f.req.URL.Path, "/movie_details.json"; got != want {
				t.Errorf("Unexpected path, got %q want %q", got, want)
			}
		})
	}
}

func TestListMovies(t *testing.T) {
	testData := []struct {
		desc      string
		opts      []ListMoviesOption
		respFile  string
		err       error
		wantQuery url.Values
		wantErr   bool
	}{
		{
			desc:      "no options",
			respFile:  "matrixes.json",
			wantQuery: url.Values{},
		},
		{
			desc:     "with limit",
			opts:     []ListMoviesOption{LMLimit(45)},
			respFile: "matrixes.json",
			wantQuery: func() url.Values {
				v := url.Values{}
				v.Set("limit", "45")
				return v
			}(),
		},
		{
			desc:     "with too large limit",
			opts:     []ListMoviesOption{LMLimit(450)},
			respFile: "matrixes.json",
			wantQuery: func() url.Values {
				v := url.Values{}
				v.Set("limit", "50")
				return v
			}(),
		},
		{
			desc:     "with page",
			opts:     []ListMoviesOption{LMPage(12)},
			respFile: "matrixes.json",
			wantQuery: func() url.Values {
				v := url.Values{}
				v.Set("page", "12")
				return v
			}(),
		},
		{
			desc:     "with quality",
			opts:     []ListMoviesOption{LMQuality("1080p")},
			respFile: "matrixes.json",
			wantQuery: func() url.Values {
				v := url.Values{}
				v.Set("quality", "1080p")
				return v
			}(),
		},
		{
			desc:     "with minimum rating",
			opts:     []ListMoviesOption{LMMinimumRating(7)},
			respFile: "matrixes.json",
			wantQuery: func() url.Values {
				v := url.Values{}
				v.Set("minimum_rating", "7")
				return v
			}(),
		},
		{
			desc:     "with too large minimum rating",
			opts:     []ListMoviesOption{LMMinimumRating(70)},
			respFile: "matrixes.json",
			wantQuery: func() url.Values {
				v := url.Values{}
				v.Set("minimum_rating", "9")
				return v
			}(),
		},
		{
			desc:     "with query",
			opts:     []ListMoviesOption{LMSearch("some title")},
			respFile: "matrixes.json",
			wantQuery: func() url.Values {
				v := url.Values{}
				v.Set("query_term", "some title")
				return v
			}(),
		},
		{
			desc:     "with genre",
			opts:     []ListMoviesOption{LMGenre("drama")},
			respFile: "matrixes.json",
			wantQuery: func() url.Values {
				v := url.Values{}
				v.Set("genre", "drama")
				return v
			}(),
		},
		{
			desc:     "with sort by",
			opts:     []ListMoviesOption{LMSortBy("title")},
			respFile: "matrixes.json",
			wantQuery: func() url.Values {
				v := url.Values{}
				v.Set("sort_by", "title")
				return v
			}(),
		},
		{
			desc:     "with order",
			opts:     []ListMoviesOption{LMOrderBy("asc")},
			respFile: "matrixes.json",
			wantQuery: func() url.Values {
				v := url.Values{}
				v.Set("order_by", "asc")
				return v
			}(),
		},
		{
			desc:     "unmarshal error",
			respFile: "bad_json.json",
			wantErr:  true,
		},
		{
			desc:     "API error",
			respFile: "error.json",
			wantErr:  true,
		},
		{
			desc:     "server error",
			respFile: "matrixes.json",
			err:      errors.New("some error"),
			wantErr:  true,
		},
	}
	f := &fakeYTSServer{
		data: loadTestData("matrix.json", t),
	}
	ts := httptest.NewServer(f)
	defer ts.Close()
	c, err := New(BaseURL(ts.URL), HTTPTimeout(time.Second*5), UserAgent("test"))
	if err != nil {
		t.Fatalf("Failed to connect to test server: %v", err)
	}
	for _, tc := range testData {
		t.Run(tc.desc, func(t *testing.T) {
			f.err = tc.err
			f.data = loadTestData(tc.respFile, t)

			_, err := c.ListMovies(tc.opts...)
			if (err != nil) != tc.wantErr {
				t.Errorf("Unexpected error, got %v want %v", err, tc.wantErr)
			}
			if err != nil {
				return
			}
			if diff := cmp.Diff(tc.wantQuery, f.req.URL.Query()); diff != "" {
				t.Errorf("Unexpected query, diff -want +got\n%s", diff)
			}
			if got, want := f.req.URL.Path, "/list_movies.json"; got != want {
				t.Errorf("Unexpected path, got %q want %q", got, want)
			}
		})
	}
}

func TestSuggestions(t *testing.T) {
	testData := []struct {
		desc      string
		id        int
		respFile  string
		err       error
		wantQuery url.Values
		wantErr   bool
	}{
		{
			desc:     "success",
			id:       1,
			respFile: "suggestions.json",
			wantQuery: func() url.Values {
				v := url.Values{}
				v.Set("movie_id", "1")
				return v
			}(),
		},
		{
			desc:     "unmarshal error",
			id:       1,
			respFile: "bad_json.json",
			wantErr:  true,
		},
		{
			desc:     "API error",
			id:       1,
			respFile: "error.json",
			wantErr:  true,
		},
		{
			desc:     "error",
			id:       1,
			respFile: "matrix.json",
			err:      errors.New("some error"),
			wantErr:  true,
		},
	}
	f := &fakeYTSServer{
		data: loadTestData("matrix.json", t),
	}
	ts := httptest.NewServer(f)
	defer ts.Close()
	c, err := New(BaseURL(ts.URL), HTTPTimeout(time.Second*5), UserAgent("test"))
	if err != nil {
		t.Fatalf("Failed to connect to test server: %v", err)
	}
	for _, tc := range testData {
		t.Run(tc.desc, func(t *testing.T) {
			f.err = tc.err
			f.data = loadTestData(tc.respFile, t)

			_, err := c.Suggestions(tc.id)
			if (err != nil) != tc.wantErr {
				t.Errorf("Unexpected error, got %v want %v", err, tc.wantErr)
			}
			if err != nil {
				return
			}
			if diff := cmp.Diff(tc.wantQuery, f.req.URL.Query()); diff != "" {
				t.Errorf("Unexpected query, diff -want +got\n%s", diff)
			}
			if got, want := f.req.URL.Path, "/movie_suggestions.json"; got != want {
				t.Errorf("Unexpected path, got %q want %q", got, want)
			}
		})
	}
}
