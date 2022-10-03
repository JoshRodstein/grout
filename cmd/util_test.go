package cmd

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"reflect"
	"runtime"
	"strings"
	"testing"

	"github.com/go-git/go-git/v5"
)

var mockRemoteURLs []string
var mockRepo LocalRepository
var mockChangeSet ChangeSet
var remoteURL1 string
var remoteURL2 string
var remoteURL3 string

func TestMain(m *testing.M) {
	log.SetOutput(ioutil.Discard)
	targetRemoteURL = "github.com"
	newRemoteURL = "gitlab.com"
	mockRepo.Name = "mockRepo"
	remoteURL1 = "https://" + targetRemoteURL + "/OldUsername/" + mockRepo.Name + dotGit
	remoteURL2 = "https://" + targetRemoteURL + "/JoshRodstein/" + mockRepo.Name + dotGit
	remoteURL3 = "https://github.com" + "/SomeUsername/" + mockRepo.Name + dotGit
	mockRepo.Path = fmt.Sprintf("/Users/mockUser/%s", mockRepo.Name)

	os.Exit(m.Run())
}

func TestCreateNewRemoteURLsNotChange(t *testing.T) {
	// reset test variables
	var noMatchRemoteURLS []string
	mockRepo.Remotes = nil
	mockChangeSet.Count = 0

	// Set remoteURL to NON matching URL (url != targetRemoteURL)
	noMatchRemoteURLS = append(noMatchRemoteURLS, remoteURL3)
	mockRepo.Remotes = append(mockRepo.Remotes, Remote{URLs: noMatchRemoteURLS})
	newURLs := createNewRemoteURLs(mockRepo.Remotes[0].URLs, &mockChangeSet)

	if len(newURLs) != 1 {
		fmt.Println(newURLs)
		t.Error()
	}
	if mockChangeSet.Count != 0 {
		t.Error()
	}
	unchangedURL := newURLs[0] != newRemoteURL+"/"+newOrganization+"/"+mockRepo.Name+dotGit
	if unchangedURL != true {
		t.Error()
	}
}

func TestCreateNewRemoteURLsNoORG(t *testing.T) {
	mockChangeSet.Count = 0
	mockRepo.Remotes = nil
	mockRemoteURLs = append(mockRemoteURLs, remoteURL1)
	mockRemoteURLs = append(mockRemoteURLs, remoteURL2)
	mockRepo.Remotes = append(mockRepo.Remotes, Remote{Name: "origin", URLs: mockRemoteURLs})

	newURLs := createNewRemoteURLs(mockRepo.Remotes[0].URLs, &mockChangeSet)
	if len(newURLs) != 2 {
		t.Error()
	}
	if mockChangeSet.Count != 2 {
		t.Error()
	}
	validNewURL1 := newURLs[0] == "https://"+newRemoteURL+"/OldOrganization/"+mockRepo.Name+dotGit
	validNewURL2 := newURLs[1] == "https://"+newRemoteURL+"/jrodstein2/"+mockRepo.Name+dotGit
	if validNewURL1 != true {
		t.Error()
	}
	if validNewURL2 != true {
		t.Error()
	}
}

func TestCreateNewRemoteURLsNewORG(t *testing.T) {
	mockChangeSet.Count = 0
	newOrganization = "NewOrg"
	mockRepo.Remotes = nil
	mockRemoteURLs = nil

	mockRemoteURLs = append(mockRemoteURLs, remoteURL1)
	mockRemoteURLs = append(mockRemoteURLs, remoteURL2)
	mockRepo.Remotes = append(mockRepo.Remotes, Remote{Name: "origin", URLs: mockRemoteURLs})
	newURLs := createNewRemoteURLs(mockRepo.Remotes[0].URLs, &mockChangeSet)

	if len(newURLs) != 2 {
		t.Error()
	}
	if mockChangeSet.Count != 2 {
		t.Error()
	}
	validNewURL1 := newURLs[0] == "https://"+newRemoteURL+"/"+newOrganization+"/"+mockRepo.Name+dotGit
	validNewURL2 := newURLs[1] == "https://"+newRemoteURL+"/"+newOrganization+"/"+mockRepo.Name+dotGit
	if validNewURL1 != true {
		t.Error()
	}
	if validNewURL2 != true {
		t.Error()
	}
}

func TestCreateChangeSetFromMap(t *testing.T) {
	var testMap RepoMap

	// re init globals
	changeSet = ChangeSet{}
	mockRepo.Remotes = nil
	testMap.Repos = nil
	mockRemoteURLs = nil

	mockRemoteURLs = append(mockRemoteURLs, remoteURL1)
	mockRemoteURLs = append(mockRemoteURLs, remoteURL2)
	mockRepo.Remotes = append(mockRepo.Remotes, Remote{Name: "origin", URLs: mockRemoteURLs})
	testMap.Repos = append(testMap.Repos, mockRepo)

	newTestChangeSet := createChangeSetFromMap(testMap)

	if newTestChangeSet.Count != 2 {
		t.Errorf("newTestChangeSet.Count: Expected 2, Got %d", newTestChangeSet.Count)
	}

}

func TestWriteChangeSetToFile(t *testing.T) {
	var testMap RepoMap

	// re init globals
	changeSet = ChangeSet{}
	mockRepo.Remotes = nil
	mockRemoteURLs = nil

	filename := "test-" + defaultPlanFile
	mockRemoteURLs = append(mockRemoteURLs, remoteURL1)
	mockRemoteURLs = append(mockRemoteURLs, remoteURL2)
	mockRepo.Remotes = append(mockRepo.Remotes, Remote{Name: "origin", URLs: mockRemoteURLs})
	testMap.Repos = append(testMap.Repos, mockRepo)

	diffTestChangeSet := createChangeSetFromMap(testMap)

	if diffTestChangeSet.Count != 2 {
		t.Errorf("diffTestChangeSet.Count: Expected 2, Got %d", diffTestChangeSet.Count)
	}
	writeChangeSetToFile(diffTestChangeSet, filename)
	_, err := os.Open(filename)
	if err != nil {
		t.Error(err)
	}
	err = os.Remove(filename)
	if err != nil {
		t.Error("Error cleaning up " + filename + " after test")
	}
}

func TestInitPlanFromFile(t *testing.T) {
	filename := "../test/test-" + defaultPlanFile
	_, err := os.Open(filename)
	if err != nil {
		t.Error(err)
	}
	result, err := initPlanFromFile(filename)
	if err != nil {
		t.Error(err)
	}
	if reflect.TypeOf(result) != reflect.TypeOf(ChangeSet{}) {
		t.Error(err)
	}
	if result.Count != 1 {
		t.Error(err)
	}
}

func TestInitPlanFromBlankFile(t *testing.T) {
	filename := "../test/test-blank-file.json"
	_, err := os.Open(filename)
	if err != nil {
		t.Error(err)
	}
	result, err := initPlanFromFile(filename)
	if err == nil {
		t.Error("Expected ERROR when initializing blank file")
	}
	if reflect.TypeOf(result) != reflect.TypeOf(ChangeSet{}) {
		t.Error(err)
	}
}

func TestExecuteChangeNoChange(t *testing.T) {
	directoryName, err := os.Getwd()
	if err != nil {
		t.Errorf("error getting current working directory: %v", err)
	}

	currentPath, err := filepath.Abs(directoryName)
	if err != nil {
		if err != nil {
			t.Errorf("error building abs path for current dir: %v", err)
		}
	}

	_, _ = git.PlainInit(currentPath, true)

	localRepo := LocalRepository{
		Name:    directoryName,
		Path:    currentPath,
		Remotes: nil,
	}

	repoPlan := RepoPlan{
		Repo:       localRepo,
		Changes:    nil,
		HasChanges: false,
	}

	var plans []RepoPlan
	plans = append(plans, repoPlan)

	set := ChangeSet{
		Count: 1,
		Plans: plans,
	}

	err = executeChanges(set)
	if err != nil {
		t.Errorf("error executing ChangeSet w/ no changes in plan: %v", err)
	}
}

// TODO: fix workspace cleanup to avoid failures on subsequent runs requiring mock repo creation
//func TestExecuteChange(t *testing.T) {
//	directoryName, err := os.Getwd()
//	if err != nil {
//		t.Errorf("error getting current working directory: %v", err)
//	}
//
//	currentPath, err := filepath.Abs(directoryName)
//	if err != nil {
//		if err != nil {
//			t.Errorf("error building abs path for current dir: %v", err)
//		}
//	}
//
//	repo, _ := git.PlainInit(currentPath, false)
//
//	var urls []string
//	urls = append(urls, remoteURL1)
//
//	remoteConfig := config.RemoteConfig{
//		Name: "origin",
//		URLs: urls,
//	}
//
//	_, err = repo.CreateRemote(&remoteConfig)
//	if err != nil {
//		t.Errorf("error creating remote for test repo: %v", err)
//	}
//
//	remote := Remote{
//		Name: remoteConfig.Name,
//		URLs: urls,
//	}
//
//	var newUrls []string
//	newUrls = append(newUrls, remoteURL2)
//
//	var remotes []Remote
//	remotes = append(remotes, remote)
//
//	localRepo := LocalRepository{
//		Name: directoryName,
//		Path: "/"+directoryName,
//		Remotes: remotes,
//	}
//
//	change := RemoteChange{
//		Name: remotes[0].Name,
//		Organization: "OldOrganization",
//		CurrentURLs: remotes[0].URLs,
//		NewURLs: newUrls,
//	}
//
//	var changes []RemoteChange
//	changes = append(changes, change)
//
//	repoPlan := RepoPlan{
//		Repo: localRepo,
//		Changes: changes,
//		HasChanges: false,
//	}
//
//	var plans []RepoPlan
//	plans = append(plans, repoPlan)
//
//	set := ChangeSet{
//		Count: 1,
//		Plans: plans,
//	}
//
//	err = executeChanges(set)
//	if err != nil {
//		t.Errorf("error executing ChangeSet in plan: %v", err)
//	}
//
//	err = repo.DeleteRemote(remoteConfig.Name)
//	if err != nil {
//		t.Errorf("error deleting test remote: %v", err)
//	}
//}

func TestTimeSlashURL(t *testing.T) {
	url := "/this/is/a/path/"
	url = trimSlashSuffix(url)
	if strings.HasSuffix(url, "/") {
		t.Error()
	}
	url = trimSlashSuffix(url)
	if !strings.HasSuffix(url, "h") {
		t.Error()
	}
}

func TestCleanPathElement(t *testing.T) {
	if runtime.GOOS == Darwin || runtime.GOOS == Linux {
		element := "/OldOrganization/"
		element = removePathSeperators(element)
		if strings.Contains(element, DarwinLinuxSep) {
			t.Error()
		}
	} else if runtime.GOOS == Windows {
		element := "\\OldOrganization\\"
		element = removePathSeperators(element)
		if strings.Contains(element, WindowsSep) {
			t.Error()
		}
	}

}

func TestVerifyTargetDirIsAbs(t *testing.T) {
	targetDir = "/this/is/an/abs/path"
	err := verifyTargetDirIsAbs()
	if err != nil {
		t.Error()
	}
}

func TestVerifyTargetDirIsNotAbs(t *testing.T) {
	targetDir = "this/is/NOT/an/abs/path"
	err := verifyTargetDirIsAbs()
	if err == nil {
		t.Error()
	}
}
