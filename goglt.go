package main

import (
	"bufio"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/VonC/gogitolite/gitolite"
	"github.com/VonC/gogitolite/reader"
)

func main() {

	flag.Parse()
	filenames := flag.Args()
	if len(filenames) == 0 {
		fmt.Println("At least one gitolite.conf file expected")
		os.Exit(1)
	}

	var usersToRepos = make(map[string][]*gitolite.Repo)
	for _, filename := range filenames {
		fmt.Printf("Read file '%v'\n", filename)
		f := process(filename, usersToRepos)
		processSubconfs(f, usersToRepos)
	}
	// print(usersToRepos)
}

func getGtl(filename string) (*os.File, *gitolite.Gitolite) {
	f, err := os.Open(filename)
	if err != nil {
		fmt.Printf("ERR %v\n", err.Error())
		os.Exit(1)
	}
	defer f.Close()
	fr := bufio.NewReader(f)
	gtl, err := reader.Read(fr)
	if err != nil {
		fmt.Printf("ERR %v\n", err.Error())
		os.Exit(1)
	}
	return f, gtl
}

func updateUsersToRepos(user *gitolite.User, config *gitolite.Config, usersToRepos map[string][]*gitolite.Repo) {
	var repos []*gitolite.Repo
	var ok bool
	if repos, ok = usersToRepos[user.GetName()]; !ok {
		repos = []*gitolite.Repo{}
	}
	for _, cfgrepo := range config.GetRepos() {
		seen := false
		for _, repo := range repos {
			if repo.GetName() == cfgrepo.GetName() {
				seen = true
				break
			}
		}
		if !seen {
			repos = append(repos, cfgrepo)
		}
	}
	usersToRepos[user.GetName()] = repos
}

func process(filename string, usersToRepos map[string][]*gitolite.Repo) *os.File {
	f, gtl := getGtl(filename)
	// fmt.Println(gtl.String())
	for _, config := range gtl.Configs() {
		for _, rule := range config.Rules() {
			if strings.Contains(rule.Access(), "R") {
				for _, user := range rule.GetUsers() {
					updateUsersToRepos(user, config, usersToRepos)
				}
			}
		}
	}
	return f
}

func visit(path string, f os.FileInfo, err error) error {
	fmt.Printf("Visited: %s\n", path)
	return nil
}

func processSubconfs(f *os.File, usersToRepos map[string][]*gitolite.Repo) {
	root := filepath.Dir(f.Name())
	fmt.Println(root)
	filepath.Walk(root, visit)
}

func print(usersToRepos map[string][]*gitolite.Repo) {
	names := make([]string, 0, len(usersToRepos))
	for username := range usersToRepos {
		names = append(names, username)
	}
	sort.Strings(names)
	for _, username := range names {
		repos := usersToRepos[username]
		for _, repo := range repos {
			typeuser := "user"
			if strings.HasPrefix(username, "proj") {
				typeuser = "system"
			}
			if strings.HasPrefix(username, "HB") {
				typeuser = "system"
			}
			if strings.Contains(username, "dmin") {
				typeuser = "system"
			}
			fmt.Printf("%v,,%v,%v\n", username, repo.GetName(), typeuser)
		}
	}

}
