package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"sort"
	"time"

	"github.com/gorilla/mux"
	"github.com/mmcdole/gofeed"
	"github.com/sirupsen/logrus"
)

var feeds map[int]RSSFeed

type RSSFeed struct {
	ID       int    `json:"id,omitempty"`
	Name     string `json:"name"`
	URL      string `json:"url"`
	Category string `json:"category"`
	Provider string `json:"provider"`
}

func main() {
	feeds = defaultRSSFeeds()
	r := mux.NewRouter()
	r.HandleFunc("/api/feeds", FeedsHandler).Methods(http.MethodGet, http.MethodPut)
	r.HandleFunc("/api/articles", ArticlesHandler).Methods(http.MethodGet).Queries("sort", "{sort}")
	r.HandleFunc("/api/articles", ArticlesHandler).Methods(http.MethodGet)
	r.HandleFunc("/api/articles/{category}", ArticlesHandler).Methods(http.MethodGet).Queries("sort", "{sort}")
	r.HandleFunc("/api/articles/{category}", ArticlesHandler).Methods(http.MethodGet)
	r.HandleFunc("/api/articles/{category}/{provider}", ArticlesHandler).Methods(http.MethodGet).Queries("sort", "{sort}")
	r.HandleFunc("/api/articles/{category}/{provider}", ArticlesHandler).Methods(http.MethodGet)
	http.Handle("/", r)

	srv := &http.Server{
		Handler:      r,
		Addr:         "127.0.0.1:8000",
		WriteTimeout: 15 * time.Second,
		ReadTimeout:  15 * time.Second,
	}

	log.Fatal(srv.ListenAndServe())
}

func defaultRSSFeeds() map[int]RSSFeed {
	res := make(map[int]RSSFeed, 0)
	defaultFeeds := []RSSFeed{
		{
			ID:       1,
			Name:     "BBC News UK",
			Category: "general",
			Provider: "BBC",
			URL:      "http://feeds.bbci.co.uk/news/uk/rss.xml",
		},
		{
			ID:       2,
			Name:     "BBC News Technology",
			Category: "technology",
			Provider: "BBC",
			URL:      "http://feeds.bbci.co.uk/news/technology/rss.xml",
		},
		{
			ID:       3,
			Name:     "Reuters UK",
			Category: "general",
			Provider: "Reuters",
			URL:      "http://feeds.reuters.com/reuters/UKdomesticNews?format=xml",
		},
		{
			ID:       4,
			Name:     "Reuters Technology",
			Category: "technology",
			Provider: "Reuters",
			URL:      "http://feeds.reuters.com/reuters/technologyNews?format=xml",
		},
	}
	for _, feed := range defaultFeeds {
		res[feed.ID] = feed
	}
	return res
}

func FeedsHandler(rw http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		b, err := json.MarshalIndent(feeds, "", "  ")
		if err != nil {
			logrus.Error(err)
			rw.Write([]byte(fmt.Sprintf("%v", err)))
		}
		rw.Write(b)
	case http.MethodPut:
		b, err := ioutil.ReadAll(r.Body)
		if err != nil {
			logrus.Error(err)
			rw.WriteHeader(http.StatusInternalServerError)
			rw.Write([]byte(fmt.Sprintf("%v", err)))
		}
		var feed RSSFeed
		err = json.Unmarshal(b, &feed)
		if err != nil {
			logrus.Error(err)
			rw.WriteHeader(http.StatusInternalServerError)
			rw.Write([]byte(fmt.Sprintf("%v", err)))
		}
		if feed.ID == 0 {
			feed.ID = len(feeds) + 1
		}
		feeds[feed.ID] = feed

		b, err = json.MarshalIndent(feeds[feed.ID], "", "  ")
		if err != nil {
			logrus.Error(err)
			rw.WriteHeader(http.StatusInternalServerError)
			rw.Write([]byte(fmt.Sprintf("%v", err)))
		}
		rw.Write(b)
	default:
		rw.WriteHeader(http.StatusMethodNotAllowed)
	}
}

func ArticlesHandler(rw http.ResponseWriter, r *http.Request) {

	vars := mux.Vars(r)
	feedsToFetch := feeds
	if vars["category"] != "" {
		for _, f := range feedsToFetch {
			if f.Category != vars["category"] {
				delete(feedsToFetch, f.ID)
			}
		}
	}
	if vars["provider"] != "" {
		for _, f := range feedsToFetch {
			if f.Provider != vars["provider"] {
				delete(feedsToFetch, f.ID)
			}
		}
	}

	articleResults, err := GetArticles(feedsToFetch)
	combinedFeed := gofeed.Feed{}
	for _, result := range articleResults {
		combinedFeed.Items = append(combinedFeed.Items, result.Articles...)
	}

	if vars["sort"] != "date" {
		sort.Sort(combinedFeed)
	}

	b, err := json.MarshalIndent(combinedFeed.Items, "", "  ")
	if err != nil {
		logrus.Error(err)
		rw.Write([]byte(fmt.Sprintf("%v", err)))
	}
	rw.Write(b)
}

type ArticleResult struct {
	FeedID   int
	Articles []*gofeed.Item
	Error    error
}

func GetArticles(feeds map[int]RSSFeed) ([]ArticleResult, error) {
	var res = make([]ArticleResult, 0)
	var err error
	c := make(chan ArticleResult)

	for _, feed := range feeds {
		go GetArticlesForFeed(feed, c)
	}

	for i := 0; i < len(feeds); i++ {
		ar := <-c
		if ar.Error != nil {
			err = fmt.Errorf("error getting Articles from Feed %v", ar.FeedID)
		}
		res = append(res, ar)
	}
	return res, err
}

func GetArticlesForFeed(feed RSSFeed, c chan ArticleResult) {
	resp, err := http.Get(feed.URL)
	if err != nil {
		logrus.Error(err)
		c <- ArticleResult{FeedID: feed.ID, Articles: nil, Error: err}
		return
	}
	b, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		logrus.Error(err)
		c <- ArticleResult{FeedID: feed.ID, Articles: nil, Error: err}
		return
	}
	f, err := gofeed.NewParser().ParseString(string(b))
	if err != nil {
		logrus.Error(err)
		c <- ArticleResult{FeedID: feed.ID, Articles: nil, Error: err}
		return
	}
	c <- ArticleResult{FeedID: feed.ID, Articles: f.Items, Error: nil}
}
