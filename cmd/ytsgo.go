package main

import (
	"flag"
	"fmt"
	"log"
	"sort"
	"strconv"
	"strings"

	"github.com/qopher/ytsgo"
)

var (
	ytsURL = flag.String("yts_url", ytsgo.DefaultBaseURL, "Base URL of yts.lt API")
)

func main() {
	flag.Parse()
	c, err := ytsgo.New(ytsgo.BaseURL(*ytsURL))
	if err != nil {
		log.Fatalf("Failed to create ytsgo client: %v", err)
	}
	if len(flag.CommandLine.Args()) != 2 {
		usage()
		return
	}
	switch flag.CommandLine.Arg(0) {
	case "movie":
		id, err := strconv.Atoi(flag.CommandLine.Arg(1))
		if err != nil {
			log.Fatalf("Failed to parse movie ID: %v", err)
		}
		m, err := c.Movie(id)
		if err != nil {
			log.Fatalf("Failed to fetch movie id:%v :%v", id, err)
		}
		fmt.Println(movieStr(m))
	case "list":
		mvs, err := c.ListMovies(ytsgo.LMSearch(flag.CommandLine.Arg(1)))
		if err != nil {
			log.Fatalf("Failed to search movies %q :%v", flag.CommandLine.Arg(1), err)
		}
		for _, m := range mvs.Movies {
			fmt.Println(movieStr(m))
		}
	default:
		usage()
		return
	}
}

func usage() {
	fmt.Printf(`Usage:
ytsgo movie [id]
ytsgo list "search term"
`)
}

func movieStr(m *ytsgo.Movie) string {
	ret := fmt.Sprintf("%q (%v)\n", m.Title, m.Year)
	var trts []string
	sort.Sort(sort.Reverse(ytsgo.TorrentsBySize(m.Torrents)))
	for _, t := range m.Torrents {
		trts = append(trts, fmt.Sprintf("\tSeeds: %v Peers: %v Size: %v\n\tMagnet: %s", t.Seeds, t.Peers, t.Size, t.Magnet()))
	}
	return ret + strings.Join(trts, "\n")
}
