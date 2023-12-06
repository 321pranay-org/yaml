package main

import (
	"context"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/google/go-github/v56/github"
	// git "github.com/go-git/go-git/v5"
	// "gopkg.in/src-d/go-git.v4/plumbing"
	"github.com/google/go-cmp/cmp"
	git "gopkg.in/src-d/go-git.v4"
	"gopkg.in/src-d/go-git.v4/plumbing"
	"gopkg.in/yaml.v3"
)

func main(){
	fmt.Println(os.Getenv("GITHUB_SERVER_URL"))
	fmt.Println(os.Getenv("GITHUB_REPOSITORY"))
	fmt.Println(os.Getenv("PR_NUMBER"))

	url := os.Getenv("GITHUB_SERVER_URL") + "/" + os.Getenv("GITHUB_REPOSITORY")
	
	_, err := git.PlainClone("master", false, &git.CloneOptions{
		URL:      strings.Replace(url, "https://", "https://" + os.Getenv("TOKEN") + "@", 1),
		Progress: os.Stdout,
	})

	if err != nil{
		fmt.Println(err)
	}

	// currDir, _ := os.Getwd()

	fmt.Println("********************")

	kongConfigMaster := getFromFileSystem("master/development/captain")

	fmt.Println(kongConfigMaster)

	fmt.Println(os.Getenv("GITHUB_HEAD_REF"))

	githubRef := os.Getenv("GITHUB_HEAD_REF")

	_, err2 := git.PlainClone("branch", false, &git.CloneOptions{
		URL:      strings.Replace(url, "https://", "https://" + os.Getenv("TOKEN") + "@", 1),
		Progress: os.Stdout,
		ReferenceName: plumbing.ReferenceName(fmt.Sprintf("refs/heads/%s", githubRef)),
	})

	if err2 != nil{
		fmt.Println(err2)
	}
	

	kongConfigBranch := getFromFileSystem("branch/development/captain")

	fmt.Println(kongConfigBranch)

	diff := cmp.Diff(kongConfigMaster, kongConfigBranch)

	fmt.Println(diff)


	commentBody := "This is first automated comment"
	commentPath := "README.md"
	commitId := os.Getenv("GITHUB_SHA")
	subjectType := "file"
	comment := &PullRequestComment{
		Body: &commentBody,
		Path: &commentPath,
		CommitID: &commitId,
		SubjectType: &subjectType,
	}

	err = createComment(comment)

	if err != nil{
		fmt.Println(err)
	}
}

type PullRequestComment struct {
	ID                  *int64     `json:"id,omitempty"`
	NodeID              *string    `json:"node_id,omitempty"`
	InReplyTo           *int64     `json:"in_reply_to_id,omitempty"`
	Body                *string    `json:"body,omitempty"`
	Path                *string    `json:"path,omitempty"`
	DiffHunk            *string    `json:"diff_hunk,omitempty"`
	PullRequestReviewID *int64     `json:"pull_request_review_id,omitempty"`
	Position            *int       `json:"position,omitempty"`
	OriginalPosition    *int       `json:"original_position,omitempty"`
	StartLine           *int       `json:"start_line,omitempty"`
	Line                *int       `json:"line,omitempty"`
	OriginalLine        *int       `json:"original_line,omitempty"`
	OriginalStartLine   *int       `json:"original_start_line,omitempty"`
	Side                *string    `json:"side,omitempty"`
	StartSide           *string    `json:"start_side,omitempty"`
	CommitID            *string    `json:"commit_id,omitempty"`
	OriginalCommitID    *string    `json:"original_commit_id,omitempty"`
	// AuthorAssociation is the comment author's relationship to the pull request's repository.
	// Possible values are "COLLABORATOR", "CONTRIBUTOR", "FIRST_TIMER", "FIRST_TIME_CONTRIBUTOR", "MEMBER", "OWNER", or "NONE".
	AuthorAssociation *string `json:"author_association,omitempty"`
	URL               *string `json:"url,omitempty"`
	HTMLURL           *string `json:"html_url,omitempty"`
	PullRequestURL    *string `json:"pull_request_url,omitempty"`
	// Can be one of: LINE, FILE from https://docs.github.com/en/rest/pulls/comments?apiVersion=2022-11-28#create-a-review-comment-for-a-pull-request
	SubjectType *string `json:"subject_type,omitempty"`
}

func createComment(comment *PullRequestComment) error{


	client := github.NewClient(nil).WithAuthToken(os.Getenv("TOKEN"))

	u := fmt.Sprintf("repos/%v/pulls/%v/comments", os.Getenv("GITHUB_REPOSITORY"), strings.Split(os.Getenv("GITHUB_REF"),"/")[2])
	req, err := client.NewRequest("POST", u, comment)
	if err != nil {
		return err
	}
	// TODO: remove custom Accept headers when their respective API fully launches.
	// acceptHeaders := []string{"application/vnd.github+json"}
	req.Header.Set("Accept", "application/vnd.github+json")

	c := new(PullRequestComment)
	_, err = client.Do(context.Background(), req, c)
	if err != nil {
		return  err
	}

	return nil
}


func getFromFileSystem(filePath string) []KongConfig {
	files := []string{}
	err := filepath.Walk(filePath, func(path string, f os.FileInfo, err error) error {
		if !f.IsDir() {
			files = append(files, path)
		}

		return err
	})

	if err != nil {
		log.Fatal(err)
		panic(err)
	}
	var kongConfig []KongConfig

	for _, f := range files {

		file, e := ioutil.ReadFile(f)
		if e != nil {
			log.Fatal(e)
		}

		data := KongConfig{}

		err = yaml.Unmarshal([]byte(file), &data)
		if err != nil {
			log.Fatalf("Error parsing file %v Reason: %v", f, err)
			panic(err)
		}

		kongConfig = append(kongConfig, data)
	}
	return kongConfig
}


type KongConfig struct {
	Resources Resources `yaml:"resources"`
}

type Resources struct {
	Services  []ServiceConfig  `yaml:"services"`
	Plugins   []PluginConfig   `yaml:"plugins"`
	Consumers []ConsumerConfig `yaml:"consumers"`
}

type ServiceConfig struct {
	Name           string         `yaml:"name"`
	Path           string         `yaml:"path"`
	Retries        int            `yaml:"retries"`
	ConnectTimeout int            `yaml:"connect_timeout"`
	WriteTimeout   int            `yaml:"write_timeout"`
	ReadTimeout    int            `yaml:"read_timeout"`
	Protocol       string         `yaml:"protocol"`
	Host           string         `yaml:"host"`
	Port           int            `yaml:"port"`
	Routes         []RoutesConfig `yaml:"routes,omitempty"`
	Plugins        []PluginConfig `yaml:"plugins,omitempty"`
}

type RoutesConfig struct {
	Name         string         `yaml:"name"`
	Protocols    []string       `yaml:"protocols"`
	Methods      []string       `yaml:"methods"`
	Hosts        []string       `yaml:"hosts"`
	Paths        []string       `yaml:"paths"`
	StripPath    bool           `yaml:"strip_path"`
	PreserveHost bool           `yaml:"preserve_host"`
	Plugins      []PluginConfig `yaml:"plugins"`
}

type PluginConfig struct {
	Name    string                 `yaml:"name" json:"name"`
	Config  map[string]interface{} `yaml:"config" json:"config"`
	Enabled bool                   `yaml:"enabled" json:"enabled"`
}

type ConsumerConfig struct {
	Username    string           `yaml:"username"`
	CustomID    string           `yaml:"customId"`
	Tags        []string         `yaml:"tags"`
	Credentials CredentialConfig `yaml:"credentials"`
}

type CredentialConfig struct {
	KeyAuth   []KeyAuthConfig   `yaml:"keyAuth"`
	BasicAuth []BasicAuthConfig `yaml:"basicAuth"`
	ACL       []ACLConfig       `yaml:"acl"`
}

type KeyAuthConfig struct {
	Key string `yaml:"key"`
}

type BasicAuthConfig struct {
	Username string `yaml:"username"`
	Password string `yaml:"password"`
	ID       string
}

type ACLConfig struct {
	Group string `yaml:"group"`
}

