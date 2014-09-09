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

type rdr struct {
	usersToRepos map[string][]*gitolite.Repo
	gtl          *gitolite.Gitolite
	f            *os.File
}

func main() {

	fauditPtr := flag.Bool("audit", false, "print user access audit")
	flag.Parse()
	filenames := flag.Args()
	if len(filenames) == 0 {
		fmt.Println("At least one gitolite.conf file expected")
		os.Exit(1)
	}

	rdr := &rdr{usersToRepos: make(map[string][]*gitolite.Repo)}
	for _, filename := range filenames {
		fmt.Printf("Read file '%v'\n", filename)
		rdr.f, rdr.gtl = rdr.process(filename, nil)
		rdr.processSubconfs()
	}
	if *fauditPtr {
		rdr.printAudit()
	}
}

func getGtl(filename string, gtl *gitolite.Gitolite) (*os.File, *gitolite.Gitolite) {
	f, err := os.Open(filename)
	if err != nil {
		fmt.Printf("ERR %v\n", err.Error())
		os.Exit(1)
	}
	defer f.Close()
	fr := bufio.NewReader(f)
	if gtl == nil {
		gtl, err = reader.Read(fr)
	} else {
		gtl, err = reader.Update(fr, gtl)
	}
	if err != nil {
		fmt.Printf("ERR %v\n", err.Error())
		os.Exit(1)
	}
	return f, gtl
}

func (rdr *rdr) updateUsersToRepos(uog gitolite.UserOrGroup, config *gitolite.Config) {
	var repos []*gitolite.Repo
	var ok bool
	if repos, ok = rdr.usersToRepos[uog.GetName()]; !ok {
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
	rdr.usersToRepos[uog.GetName()] = repos
}

func (rdr *rdr) process(filename string, parent *gitolite.Gitolite) (*os.File, *gitolite.Gitolite) {
	f, gtl := getGtl(filename, parent)
	// fmt.Println(gtl.String())
	for _, config := range gtl.Configs() {
		for _, rule := range config.Rules() {
			if strings.Contains(rule.Access(), "R") {
				for _, uog := range rule.GetUsersFirstOrGroups() {
					rdr.updateUsersToRepos(uog, config)
				}
			}
		}
	}
	return f, gtl
}

func (rdr *rdr) processSubconfs() {
	root := filepath.Dir(rdr.f.Name())
	fmt.Println(root)
	filepath.Walk(root, func(path string, f os.FileInfo, err error) error {
		if !f.IsDir() {
			relname := strings.Replace(path, root+string(filepath.Separator), "", 1)
			relname = strings.Replace(relname, string(filepath.Separator), "/", -1)
			for _, subconfrx := range rdr.gtl.Subconfs() {
				//fmt.Printf("%v %v\n", subconfrx.String(), relname)
				if subconfrx.MatchString(relname) {
					fmt.Printf("Visited: %s %s\n", relname, path)
					rdr.process(path, rdr.gtl)
				}
			}
		}
		return nil
	})
}

func (rdr *rdr) printAudit() {
	names := make([]string, 0, len(rdr.usersToRepos))
	for username := range rdr.usersToRepos {
		names = append(names, username)
	}
	sort.Strings(names)
	for _, username := range names {
		repos := rdr.usersToRepos[username]
		for _, repo := range repos {
			typeuser := "user"
			if strings.HasPrefix(username, "@") {
				typeuser = "system"
			} else if strings.HasPrefix(username, "proj") {
				typeuser = "system"
			} else if strings.HasPrefix(username, "HB") {
				typeuser = "system"
			} else if strings.Contains(username, "dmin") {
				typeuser = "system"
			}
			fmt.Printf("%v,,%v,%v\n", username, repo.GetName(), typeuser)
		}
	}

}
