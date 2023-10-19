package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strings"

	// "github.com/google/go-github/v56/github"
	// git "github.com/go-git/go-git/v5"
	// "gopkg.in/src-d/go-git.v4/plumbing"
	git "gopkg.in/src-d/go-git.v4"
	"gopkg.in/src-d/go-git.v4/plumbing"
	"gopkg.in/yaml.v3"
)

func main(){

	// client := github.NewClient(nil).WithAuthToken(os.Getenv("TOKEN"))

	fmt.Println(os.Getenv("GITHUB_SERVER_URL"))
	fmt.Println(os.Getenv("GITHUB_REPOSITORY"))

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

