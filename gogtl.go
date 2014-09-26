package main

import (
	"bufio"
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/VonC/gogitolite/gitolite"
	"github.com/VonC/gogitolite/project"
	"github.com/VonC/gogitolite/reader"
)

type rdr struct {
	usersToReposOrGroup map[string][]gitolite.RepoOrGroup
	gtl                 *gitolite.Gitolite
	verbose             bool
	subconfs            map[string]*gitolite.Gitolite
	filename            string
}

var (
	args        []string
	r           *rdr
	fauditPtr   *bool = flag.Bool("audit", false, "print user access audit")
	flistPtr    *bool = flag.Bool("list", false, "list projects")
	fverbosePtr *bool = flag.Bool("v", false, "verbose, display filenames read")
	fprintPtr   *bool = flag.Bool("print", false, "print config")

	sout *bufio.Writer
	serr *bufio.Writer
)

func init() {
	flag.Usage = func() {
		fmt.Fprintf(oerr(),
			"Usage: gogitolite.exe [opts] gitolite.conf\n")
		fmt.Fprintf(oerr(), "Options:\n")
		flag.VisitAll(func(flag *flag.Flag) {
			format := "  -%s=%s: %s\n"
			if !strings.HasPrefix(flag.Name, "test.") {
				fmt.Fprintf(oerr(), format, flag.Name, flag.DefValue, flag.Usage)
			}
		})
	}
}

func out() io.Writer {
	if sout == nil {
		return os.Stdout
	}
	return sout
}

func oerr() io.Writer {
	if serr == nil {
		return os.Stderr
	}
	return serr
}

func main() {

	a := os.Args[1:]
	if args != nil {
		a = args
		if len(a) == 1 && a[0] == "-h" {
			flag.Usage()
			return
		}
	}
	flag.CommandLine.Parse(a)
	filenames := flag.Args()
	var filename string
	var err error
	if len(filenames) != 1 {
		fmt.Fprintf(serr, "%s", "One gitolite.conf file expected")
		goto eop
	}
	filename = filenames[0]
	r = &rdr{usersToReposOrGroup: make(map[string][]gitolite.RepoOrGroup),
		verbose:  *fverbosePtr,
		filename: filename,
		subconfs: make(map[string]*gitolite.Gitolite),
	}
	if r.verbose {
		fmt.Fprintf(out(), "Read file '%v'\n", filename)
	}

	r.gtl, err = r.process(filename, nil)
	if err == nil {
		r.processSubconfs()
		if *fauditPtr {
			r.printAudit()
		}
		if *flistPtr {
			r.listProjects()
		}
		if *fprintPtr {
			fmt.Fprintf(out(), "%v", r.gtl.Print())
		}
	}
eop:
}

func getGtlFromFile(filename string, gtl *gitolite.Gitolite) (*gitolite.Gitolite, error) {
	f, err := os.Open(filename)
	if err != nil {
		fmt.Fprintf(oerr(), "ERR %v\n", err.Error())
		return nil, err
	}
	defer f.Close()
	fr := bufio.NewReader(f)
	return getGtl2(fr, gtl)
}

func getGtl2(r io.Reader, gtl *gitolite.Gitolite) (*gitolite.Gitolite, error) {
	var err error
	if gtl == nil {
		gtl, err = reader.Read(r)
	} else {
		gtl, err = reader.Update(r, gtl)
	}
	if err != nil {
		fmt.Fprintf(oerr(), "ERR %v\n", err.Error())
		return nil, err
	}
	return gtl, nil
}

func addRogNoDup(rog gitolite.RepoOrGroup, rogs []gitolite.RepoOrGroup) []gitolite.RepoOrGroup {
	res := rogs
	seen := false
	for _, arog := range rogs {
		if arog.GetName() == rog.GetName() {
			seen = true
			break
		}
	}
	if !seen {
		res = append(rogs, rog)
	}
	return res
}

func (rdr *rdr) updateUsersToRepos(uog gitolite.UserOrGroup, config *gitolite.Config) {
	var rogs []gitolite.RepoOrGroup
	var ok bool
	if rogs, ok = rdr.usersToReposOrGroup[uog.GetName()]; !ok {
		rogs = []gitolite.RepoOrGroup{}
	}
	for _, cfgrog := range config.GetReposOrGroups() {
		rogs = addRogNoDup(cfgrog, rogs)
		if cfgrog.Group() != nil {
			cfggrp := cfgrog.Group()
			repos := cfggrp.GetAllRepos()
			for _, repo := range repos {
				rogs = addRogNoDup(repo, rogs)
			}
		}
	}
	rdr.usersToReposOrGroup[uog.GetName()] = rogs
}

func (rdr *rdr) process(filename string, parent *gitolite.Gitolite) (*gitolite.Gitolite, error) {
	gtl, err := getGtlFromFile(filename, parent)
	if err != nil {
		return nil, err
	}
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
	return gtl, nil
}

func (rdr *rdr) processSubconfs() {
	root := filepath.Dir(rdr.filename)
	//fmt.Println(root)
	filepath.Walk(root, func(path string, f os.FileInfo, err error) error {
		if !f.IsDir() {
			relname := strings.Replace(path, root+string(filepath.Separator), "", 1)
			relname = strings.Replace(relname, string(filepath.Separator), "/", -1)
			for _, subconfrx := range rdr.gtl.Subconfs() {
				//fmt.Printf("%v %v\n", subconfrx.String(), relname)
				if subconfrx.MatchString(relname) {
					if rdr.verbose {
						fmt.Fprintf(out(), "Visited: %s %s\n", relname, path)
					}
					subgtl, err := rdr.process(path, rdr.gtl)
					if err != nil {
						fmt.Fprintf(oerr(), "Ignore subconf file: %s %s because of err '%v'\n", relname, path, err)
					} else {
						rdr.subconfs[path] = subgtl
					}
				}
			}
		}
		return nil
	})
}

func (rdr *rdr) printAudit() {
	names := make([]string, 0, len(rdr.usersToReposOrGroup))
	for username := range rdr.usersToReposOrGroup {
		names = append(names, username)
	}
	sort.Strings(names)
	for _, username := range names {
		repos := rdr.usersToReposOrGroup[username]
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
			fmt.Fprintf(out(), "%v,,%v,%v\n", username, repo.GetName(), typeuser)
		}
	}

}

func (rdr *rdr) listProjects() {
	pm := project.NewManager(rdr.gtl, rdr.subconfs)
	fmt.Fprintf(out(), "NbProjects: %v\n", pm.NbProjects())
	for _, project := range pm.Projects() {
		fmt.Fprintf(out(), "%v\n", project)
	}
}
