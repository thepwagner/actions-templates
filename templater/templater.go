package templater

import (
	"fmt"
	"sync"
	"text/template"

	"github.com/google/go-github/v62/github"
)

type Repo struct {
	Owner string `json:"owner"`
	Name  string `json:"name"`
}

func (r Repo) String() string {
	return fmt.Sprintf("%s/%s", r.Owner, r.Name)
}

type Config struct {
	Repositories map[string]map[string]*RepositoryConfiguration `yaml:"repositories"`
	Secrets      map[string]string                              `yaml:"secrets"`

	Auth struct {
		GitHub string `yaml:"github"`
	} `yaml:"auth"`
	Registry string `yaml:"registry"`

	Committer *github.CommitAuthor `yaml:"committer"`
}

func (c Config) RepositoryCount() (ret int) {
	for _, repo := range c.Repositories {
		for range repo {
			ret++
		}
	}
	return
}

type RepositoryConfiguration struct {
	PreBuild  string `yaml:"prebuild"`
	PostBuild string `yaml:"postbuild"`
}

type Templater struct {
	config *Config
	client *github.Client

	loadTemplates sync.Once
	templates     map[string]*template.Template
}

func NewTemplater(client *github.Client, config *Config) *Templater {
	return &Templater{
		client: client,
		config: config,
	}
}
