package cmd

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/config"
)

// Some of these constants are silly, but I like them
const (
	// need to fix '/' logic to be more consistent
	defaultTargetHostname = "github.com"
	defaultNewHostname    = "github.com"
	defaultPlanFile       = "grout-plan.json"
	defaultRemoteType     = https

	Darwin         = "darwin"
	Linux          = "linux"
	Windows        = "windows"
	WindowsSep     = "\\"
	DarwinLinuxSep = "/"

	dotGit    = ".git"
	justGit   = "git"
	http      = "http"
	https     = "https"
	twoSpaces = "  "
	sixSpaces = "  "

	Yes = "y"
)

// Global package variables
var cfgFile string
var targetDir string
var parentDir string
var remoteType string
var targetOrganization string
var newOrganization string
var newRemoteURL string
var targetRemoteURL string

var repoMap RepoMap
var changeSet ChangeSet
var errorBundle ErrorBundle

// Structs to aid in mapping, planning, and executing changes
type ErrorBundle struct {
	Count  int
	Errors []error
}

type SplitUrl struct {
	Type    string
	BaseURL string
	Org     string
	Repo    string
}

type Remote struct {
	Name string   `json:"name"`
	URLs []string `json:"urls"`
}

type LocalRepository struct {
	Name    string   `json:"name"`
	Path    string   `json:"path"`
	Remotes []Remote `json:"remotes"`
}

type RepoMap struct {
	Meta  map[string]string `json:"meta"`
	Repos []LocalRepository `json:"repos"`
}

// Change structs represent changes to a repo's remotes
type RemoteChange struct {
	Name         string   `json:"name"`
	Organization string   `json:"newOrganization"`
	CurrentURLs  []string `json:"current_urls"`
	NewURLs      []string `json:"new_urls"`
}

type RepoPlan struct {
	Repo       LocalRepository `json:"repo"`
	Changes    []RemoteChange  `json:"changes"`
	HasChanges bool            `json:"has_changes"`
}

type ChangeSet struct {
	Count int        `json:"count"`
	Plans []RepoPlan `json:"plans"`
}

// Build a new remote url string if the given remote matches our
// defaultTargetRemoteURL
func createNewRemoteURLs(urls []string, set *ChangeSet) []string {
	var newRemoteURLs []string

	for _, url := range urls {
		fields := strings.FieldsFunc(url, UrlSplit)
		if len(fields) != 4 {
			// Malformed Remote URL, skipping
			continue
		}
		splitUrl := SplitUrl{
			Type:    fields[0],
			BaseURL: fields[1],
			Org:     fields[2],
			Repo:    fields[3],
		}
		if splitUrl.BaseURL == targetRemoteURL {
			if len(targetOrganization) > 0 && targetOrganization != splitUrl.Org {
				newRemoteURLs = append(newRemoteURLs, url)
			} else {
				if len(newOrganization) > 0 {
					splitUrl.Org = newOrganization
				}

				var newRemote string

				if splitUrl.Type == justGit {
					newRemote = fmt.Sprintf("%s@%s:%s/%s", splitUrl.Type, newRemoteURL, splitUrl.Org, splitUrl.Repo)
				} else if strings.Contains(splitUrl.Type, http) {
					newRemote = fmt.Sprintf("%s://%s/%s/%s", remoteType, newRemoteURL, splitUrl.Org, splitUrl.Repo)
				}
				newRemoteURLs = append(newRemoteURLs, newRemote)

				set.Count++
			}
		} else {
			newRemoteURLs = append(newRemoteURLs, url)
		}
	}
	return newRemoteURLs
}

// Searches for a git repo in a given directory path
// If found, the repo is added to the RepoMap
func mapRepository(path string, info os.FileInfo, err error, repoMap *RepoMap) error {
	if err != nil {
		errorBundle.Count += 1
		errorBundle.Errors = append(errorBundle.Errors, err)
		return nil
	}
	if info.Name() == dotGit {
		r, err := git.PlainOpen(path)
		if err != nil {
			fmt.Printf("Error opening repo: %v\n", err)
			return err
		}

		remotes, err := r.Remotes()
		if err != nil {
			fmt.Println(err)
			return err
		}

		var mappedRemotes []Remote
		for _, remote := range remotes {
			mappedRemotes = append(mappedRemotes, Remote{
				Name: remote.Config().Name,
				URLs: remote.Config().URLs,
			})
		}

		currentRepo := LocalRepository{
			Name:    parentDir,
			Path:    path,
			Remotes: mappedRemotes,
		}
		repoMap.Repos = append(repoMap.Repos, currentRepo)

	} else if info.IsDir() {
		parentDir = info.Name()
	}
	return nil
}

// This generates a git remote change set for a given map of Repos
func createChangeSetFromMap(repoMap RepoMap) ChangeSet {
	for _, repo := range repoMap.Repos {

		var plan RepoPlan
		var changePlan RepoPlan

		changePlan.HasChanges = false
		plan.Repo = repo

		for _, remote := range repo.Remotes {
			currentURLs := remote.URLs
			newURLs := createNewRemoteURLs(currentURLs, &changeSet)
			for i := 0; i < len(newURLs); i++ {
				if currentURLs[i] == newURLs[i] {
					continue
				} else {
					change := RemoteChange{
						Name: remote.Name,
					}
					change.CurrentURLs = append(change.CurrentURLs, currentURLs[i])
					change.NewURLs = append(change.NewURLs, newURLs[i])

					changePlan.Changes = append(changePlan.Changes, change)
					changePlan.HasChanges = true
				}
			}
		}
		if changePlan.HasChanges {
			changePlan.Repo = plan.Repo
			changeSet.Plans = append(changeSet.Plans, changePlan)
		}
	}
	return changeSet
}

// Write all of our calculated changes to a json file in the current dir
func writeChangeSetToFile(changes ChangeSet, filename string) {
	jsonStr, err := json.MarshalIndent(changes, "", twoSpaces)
	if err != nil {
		fmt.Println(err)
		return
	}
	err = ioutil.WriteFile(filename, jsonStr, os.ModePerm)
	if err != nil {
		fmt.Println(err)
		return
	}
}

// Initialize json file into a ChangeSet
func initPlanFromFile(fd string) (ChangeSet, error) {
	fmt.Println("\nInitializing plan...")
	jsonFile, err := os.Open(fd)
	if err != nil {
		fmt.Printf("Unable to open plan file")
		return ChangeSet{}, err
	}
	byteValue, err := ioutil.ReadAll(jsonFile)
	if err != nil {
		fmt.Printf("Error reading json: %s\n", err)
	}
	byteValue = bytes.ReplaceAll(byteValue, []byte(twoSpaces), []byte(""))
	var set ChangeSet
	err = json.Unmarshal([]byte(byteValue), &set)
	if err != nil {
		fmt.Println("Error unmarshalling plan from file. Try recreating the plan before running update again.")
		return ChangeSet{}, err
	}
	return set, nil
}

func executeChanges(set ChangeSet) error {
	for _, plan := range set.Plans {
		repoPath := plan.Repo.Path
		gitRepo, err := git.PlainOpen(repoPath)
		if err != nil {
			fmt.Println(err)
			return err
		}
		for _, change := range plan.Changes {
			err := updateRemote(&change, gitRepo)
			if err != nil {
				fmt.Println(err)
				return err
			}
		}
	}
	return nil
}

func updateRemote(change *RemoteChange, gitRepo *git.Repository) error {
	remoteName := change.Name
	err := gitRepo.DeleteRemote(remoteName)
	if err != nil {
		fmt.Printf("Error deleting remote: %s\n", err)
		return err
	}
	remoteConfig := config.RemoteConfig{
		Name: remoteName,
		URLs: change.NewURLs,
	}
	if _, err = gitRepo.CreateRemote(&remoteConfig); err != nil {
		fmt.Printf("Error creating remote: %s\n", err)
		return err
	}
	return nil
}

func trimSlashSuffix(str string) string {
	str = strings.TrimSuffix(str, "/")
	return str
}

func removePathSeperators(str string) string {
	if runtime.GOOS == Windows {
		//fmt.Printf("arch: %s", Windows)
		str = strings.ReplaceAll(str, "\\", "")
	} else if runtime.GOOS == Darwin || runtime.GOOS == Linux {
		//fmt.Printf("arch: %s/%s", Darwin, Linux)
		str = strings.ReplaceAll(str, "/", "")
	}

	return str
}

func UrlSplit(r rune) bool {
	return r == ':' || r == '/' || r == '@'
}

func verifyTargetDirIsAbs() error {
	if !filepath.IsAbs(targetDir) {
		fmt.Printf("\nInvalid parameter: %s \n"+
			"Absolute path is required for search directory - Aborting\n", targetDir)
		return errors.New("invalid parameter")
	}
	return nil
}
