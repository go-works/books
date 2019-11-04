package main

import (
	"encoding/json"
	"io/ioutil"
	"net/http"

	"github.com/kjk/u"
)

const (
	githubServer = "https://api.github.com"
)

type Gist struct {
	URL         string               `json:"url"`
	ForksURL    string               `json:"forks_url"`
	CommitsURL  string               `json:"commits_url"`
	ID          string               `json:"id"`
	NodeID      string               `json:"node_id"`
	GitPullURL  string               `json:"git_pull_url"`
	GitPushURL  string               `json:"git_push_url"`
	HTMLURL     string               `json:"html_url"`
	Files       map[string]*GistFile `json:"files"`
	Public      bool                 `json:"public"`
	CreatedAt   string               `json:"created_at"`
	UpdatedAt   string               `json:"updated_at"`
	Description string               `json:"description"`
	Comments    int64                `json:"comments"`
	User        interface{}          `json:"user"`
	CommentsURL string               `json:"comments_url"`
	Owner       GitOwner             `json:"owner"`
	Forks       []interface{}        `json:"forks"`
	History     []GistHistory        `json:"history"`
	Truncated   bool                 `json:"truncated"`
}

type GistFile struct {
	Filename  string `json:"filename"`
	Type      string `json:"type"`
	Language  string `json:"language"`
	RawURL    string `json:"raw_url"`
	Size      int64  `json:"size"`
	Truncated bool   `json:"truncated"`
	Content   string `json:"content"`
}

type GistHistory struct {
	User         GitOwner        `json:"user"`
	Version      string          `json:"version"`
	CommittedAt  string          `json:"committed_at"`
	ChangeStatus GitChangeStatus `json:"change_status"`
	URL          string          `json:"url"`
}

type GitChangeStatus struct {
	Total     int64 `json:"total"`
	Additions int64 `json:"additions"`
	Deletions int64 `json:"deletions"`
}

type GitOwner struct {
	Login             string `json:"login"`
	ID                int64  `json:"id"`
	NodeID            string `json:"node_id"`
	AvatarURL         string `json:"avatar_url"`
	GravatarID        string `json:"gravatar_id"`
	URL               string `json:"url"`
	HTMLURL           string `json:"html_url"`
	FollowersURL      string `json:"followers_url"`
	FollowingURL      string `json:"following_url"`
	GistsURL          string `json:"gists_url"`
	StarredURL        string `json:"starred_url"`
	SubscriptionsURL  string `json:"subscriptions_url"`
	OrganizationsURL  string `json:"organizations_url"`
	ReposURL          string `json:"repos_url"`
	EventsURL         string `json:"events_url"`
	ReceivedEventsURL string `json:"received_events_url"`
	Type              string `json:"type"`
	SiteAdmin         bool   `json:"site_admin"`
}

func gistDecode(s string) *Gist {
	var res Gist
	err := json.Unmarshal([]byte(s), &res)
	must(err)
	return &res
}

func githubGet(endpoint string) string {
	uri := githubServer + endpoint
	resp, err := http.Get(uri)
	must(err)
	panicIf(resp.StatusCode != http.StatusOK, "http.Get('%s') failed with '%s'", uri, resp.Status)
	defer u.CloseNoError(resp.Body)
	d, err := ioutil.ReadAll(resp.Body)
	must(err)
	return string(d)
}

// TODO: download truncated files
func gistDownload(gistID string) string {
	endpoint := "/gists/" + gistID
	return githubGet(endpoint)
}
