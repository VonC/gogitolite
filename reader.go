package gogitolite

import (
	"bufio"
	"fmt"
	"io"
	"regexp"
	"strings"
)

// Gitolite config decoded
type Gitolite struct {
	groups         []*Group
	repos          []*Repo
	users          []*User
	namesToGroups  map[string][]*Group
	repoGroups     []*Group
	userGroups     []*Group
	configs        []*Config
	reposToConfigs map[string][]*Config
}

type content struct {
	s   *bufio.Scanner
	l   int
	gtl *Gitolite
}

type stateFn func(*content) (stateFn, error)

// Group (of repo or resources, ie people)
type Group struct {
	name      string
	members   []string
	kind      kind
	container container
}

func (g *Group) String() string {
	res := fmt.Sprintf("group '%v'(%v): %+v", g.name, g.kind, g.members)
	return res
}

type container interface {
	addReposGroup(grp *Group)
	addUsersGroup(grp *Group)
}

type kind int

const (
	undefined = iota
	users
	repos
)

// Repo (single or group name)
type Repo struct {
	name string
}

// Read a gitolite config file
func Read(r io.Reader) (*Gitolite, error) {
	res := &Gitolite{namesToGroups: make(map[string][]*Group), reposToConfigs: make(map[string][]*Config)}
	if r == nil {
		return res, nil
	}
	s := bufio.NewScanner(r)
	s.Scan()
	c := &content{s: s, gtl: res, l: 1}
	var state stateFn
	var err error
	for state, err = readEmptyOrCommentLines(c); state != nil && err == nil; {
		state, err = state(c)
	}
	fmt.Printf("\nGitolite res='%v'\n", res)
	return res, err
}

func (gtl *Gitolite) String() string {
	res := fmt.Sprintf("NbGroups: %v [", len(gtl.groups))
	for i, group := range gtl.groups {
		if i > 1 {
			res = res + ", "
		}
		res = res + group.name
	}
	res = res + "]\n"

	res = res + fmt.Sprintf("NbRepoGroups: %v [", len(gtl.repoGroups))
	for i, repogrp := range gtl.repoGroups {
		if i > 0 {
			res = res + ", "
		}
		res = res + repogrp.name
	}
	res = res + "]\n"

	res = res + fmt.Sprintf("NbRepos: %v %+v\n", len(gtl.repos), gtl.repos)
	res = res + fmt.Sprintf("NbUsers: %v %+v\n", len(gtl.users), gtl.users)
	res = res + fmt.Sprintf("NbUserGroups: %v [", len(gtl.userGroups))
	for i, usergrp := range gtl.userGroups {
		if i > 1 {
			res = res + ", "
		}
		res = res + usergrp.name
	}
	res = res + "]\n"
	res = res + fmt.Sprintf("NbConfigs: %v [", len(gtl.configs))
	for i, config := range gtl.configs {
		if i > 0 {
			res = res + ", "
		}
		res = res + config.String()
	}
	res = res + "]\n"
	res = res + fmt.Sprintf("namesToGroups: %v [", len(gtl.namesToGroups))
	first := true
	for name, groups := range gtl.namesToGroups {
		if !first {
			res = res + ", "
		}
		first = false
		res = res + fmt.Sprintf("%v => %+v", name, groups)
	}
	res = res + "]\n"
	res = res + fmt.Sprintf("reposToConfigs: %v [", len(gtl.reposToConfigs))
	first = true
	for reponame, config := range gtl.reposToConfigs {
		if !first {
			res = res + ", "
		}
		first = false
		res = res + fmt.Sprintf("%v => %+v", reponame, config)
	}
	res = res + "]\n"

	return res
}

// IsEmpty checks if config includes any repo or groups
func (gtl *Gitolite) IsEmpty() bool {
	return (gtl.groups == nil || len(gtl.groups) == 0) && (gtl.repos == nil || len(gtl.repos) == 0)
}

// ParseError indicates gitolite.conf parsing error
type ParseError struct {
	msg string
}

func (pe ParseError) Error() string {
	return fmt.Sprintf("Parse Error: %s", pe.msg)
}

var readEmptyOrCommentLinesRx = regexp.MustCompile(`(?m)^\s*?$|^\s*?#(.*?)$`)

func readEmptyOrCommentLines(c *content) (stateFn, error) {
	t := strings.TrimSpace(c.s.Text())
	for keepReading := true; keepReading; {
		res := readEmptyOrCommentLinesRx.FindStringSubmatchIndex(t)
		//fmt.Println(res, ">'"+t+"'")
		if res == nil {
			return readRepoOrGroup, nil
		}
		if !c.s.Scan() {
			keepReading = false
		} else {
			c.l = c.l + 1
			t = strings.TrimSpace(c.s.Text())
		}
	}
	if c.gtl.IsEmpty() {
		return nil, ParseError{msg: fmt.Sprintf("comment, group or repo expected at line %v ('%v')", c.l, t)}
	}
	return nil, nil
}

var readRepoOrGroupRx = regexp.MustCompile(`^\s*?(repo |@)`)

func readRepoOrGroup(c *content) (stateFn, error) {
	t := strings.TrimSpace(c.s.Text())
	res := readRepoOrGroupRx.FindStringSubmatchIndex(t)
	if res == nil {
		return nil, ParseError{msg: fmt.Sprintf("group or repo expected after line %v ('%v')", c.l, t)}
	}
	prefix := t[res[2]:res[3]]
	if prefix == "@" {
		return readGroup, nil
	}
	return readRepo, nil
}

var readGroupRx = regexp.MustCompile(`(?m)^\s*?(@[a-zA-Z0-9_-]+)\s*?=\s*?((?:[a-zA-Z0-9_-]+\s*?)+)$`)

func readGroup(c *content) (stateFn, error) {
	t := strings.TrimSpace(c.s.Text())
	res := readGroupRx.FindStringSubmatchIndex(t)
	//fmt.Println(res, "'"+t+"'")
	if len(res) == 0 {
		return nil, ParseError{msg: fmt.Sprintf("Incorrect group declaration at line %v ('%v')", c.l, t)}
	}
	//fmt.Println(res, "'"+c.s+"'", "'"+c.s[res[2]:res[3]]+"'", "'"+c.s[res[4]:res[5]]+"'")
	grpname := t[res[2]:res[3]]
	grpmembers := strings.Split(strings.TrimSpace(t[res[4]:res[5]]), " ")
	grp := &Group{name: grpname, members: grpmembers, container: c.gtl}
	for _, g := range c.gtl.groups {
		if g.name == grpname {
			return nil, ParseError{msg: fmt.Sprintf("Duplicate group name '%v' at line %v ('%v')", grpname, c.l, t)}
		}
	}
	// http://cats.groups.google.com.meowbify.com/forum/#!topic/golang-nuts/-pqkICuokio
	//fmt.Printf("'%v'\n", grpmembers)
	seen := map[string]bool{}
	for _, val := range grpmembers {
		if _, ok := seen[val]; !ok {
			seen[val] = true
		} else {
			return nil, ParseError{msg: fmt.Sprintf("Duplicate group element name '%v' at line %v ('%v')", val, c.l, t)}
		}
		c.gtl.namesToGroups[val] = append(c.gtl.namesToGroups[val], grp)
	}
	c.gtl.groups = append(c.gtl.groups, grp)
	c.gtl.namesToGroups[grpname] = append(c.gtl.namesToGroups[grpname], grp)
	// fmt.Println("'" + c.s + "'")
	if !c.s.Scan() {
		return nil, nil
	}
	c.l = c.l + 1
	return readEmptyOrCommentLines, nil
}

// NbGroup returns the number of groups (people or repos)
func (gtl *Gitolite) NbGroup() int {
	return len(gtl.groups)
}

// NbRepos returns the number of repos (single or groups)
func (gtl *Gitolite) NbRepos() int {
	return len(gtl.repos)
}

var readRepoRx = regexp.MustCompile(`(?m)^\s*?repo\s*?((?:@?[a-zA-Z0-9_-]+\s*?)+)$`)

func readRepo(c *content) (stateFn, error) {
	t := strings.TrimSpace(c.s.Text())
	//fmt.Println(res, "'"+t+"'")
	res := readRepoRx.FindStringSubmatchIndex(t)
	if len(res) == 0 {
		return nil, ParseError{msg: fmt.Sprintf("Incorrect repo declaration at line %v ('%v')", c.l, t)}
	}
	rpmembers := strings.Split(strings.TrimSpace(t[res[2]:res[3]]), " ")
	seen := map[string]bool{}
	for _, val := range rpmembers {
		if _, ok := seen[val]; !ok {
			seen[val] = true
		} else {
			return nil, ParseError{msg: fmt.Sprintf("Duplicate repo element name '%v' at line %v ('%v')", val, c.l, t)}
		}
	}
	config := &Config{repos: []*Repo{}}
	c.gtl.configs = append(c.gtl.configs, config)
	for _, rpname := range rpmembers {
		if !strings.HasPrefix(rpname, "@") {
			addRepoFromName(c.gtl, rpname, c.gtl)
			addRepoFromName(config, rpname, c.gtl)
			if grps, ok := c.gtl.namesToGroups[rpname]; ok {
				for _, grp := range grps {
					if err := grp.markAsRepoGroup(); err != nil {
						return nil, ParseError{msg: fmt.Sprintf("repo name '%v' already used user group at line %v ('%v')\n%v", rpname, c.l, t, err.Error())}
					}
				}
			}
		} else {
			var group *Group
			for _, g := range c.gtl.groups {
				if g.name == rpname {
					group = g
					break
				}
			}
			if group == nil {
				return nil, ParseError{msg: fmt.Sprintf("repo group name '%v' undefined at line %v ('%v')", rpname, c.l, t)}
			}
			fmt.Printf("\n%v\n", group)
			group.markAsRepoGroup()
			for _, rpname := range group.members {
				addRepoFromName(c.gtl, rpname, c.gtl)
				addRepoFromName(config, rpname, c.gtl)
			}
		}
	}

	if !c.s.Scan() {
		return nil, nil
	}
	c.l = c.l + 1
	return readRepoRules, nil
}

func (r *Repo) String() string {
	return fmt.Sprintf("repo '%v'", r.name)
}
func (u *User) String() string {
	return fmt.Sprintf("user '%v'", u.name)
}

type repoContainer interface {
	getRepos() []*Repo
	addRepo(repo *Repo)
}

func (gtl *Gitolite) getRepos() []*Repo {
	return gtl.repos
}
func (gtl *Gitolite) addRepo(repo *Repo) {
	gtl.repos = append(gtl.repos, repo)
}

func (cfg *Config) getRepos() []*Repo {
	return cfg.repos
}
func (cfg *Config) addRepo(repo *Repo) {
	cfg.repos = append(cfg.repos, repo)
}

func addRepoFromName(rc repoContainer, rpname string, allReposCtn repoContainer) {
	var repo *Repo
	for _, r := range allReposCtn.getRepos() {
		if r.name == rpname {
			repo = r
		}
	}
	if repo == nil {
		repo = &Repo{name: rpname}
		if rc != allReposCtn {
			allReposCtn.addRepo(repo)
		}
	}
	seen := false
	for _, arepo := range rc.getRepos() {
		if arepo.name == repo.name {
			seen = true
			break
		}
	}
	if !seen {
		rc.addRepo(repo)
	}

}

func (gtl *Gitolite) addReposGroup(grp *Group) {
	gtl.repoGroups = append(gtl.repoGroups, grp)
	for _, reponame := range grp.members {
		addRepoFromName(gtl, reponame, gtl)
	}
}

func (grp *Group) markAsRepoGroup() error {
	if grp.kind == users {
		return fmt.Errorf("group '%v' is a users group, not a repo one", grp.name)
	}
	if grp.kind == undefined {
		grp.kind = repos
		grp.container.addReposGroup(grp)
	}
	return nil
}

// NbGroupRepos returns the number of groups identified as repos
func (gtl *Gitolite) NbGroupRepos() int {
	return len(gtl.repoGroups)
}

// User (or group of users)
type User struct {
	name string
}

// Config for repos with access rules
type Config struct {
	repos []*Repo
	rules []*Rule
	desc  string
}

func (cfg *Config) String() string {
	res := fmt.Sprintf("config %+v => %+v", cfg.repos, cfg.rules)
	return res
}

// Rule (of access to repo)
type Rule struct {
	access string
	param  string
	users  []*User
}

func (r *Rule) String() string {
	users := ""
	for _, user := range r.users {
		users = " " + user.name
	}
	users = strings.TrimSpace(users)
	return fmt.Sprintf("%v %v %v", r.access, r.param, users)
}

var readRepoRuleRx = regexp.MustCompile(`(?m)^\s*?([^@=]+)\s*?=\s*?((?:[a-zA-Z0-9_-]+\s*?)+)$`)
var repoRulePreRx = regexp.MustCompile(`(?m)^([RW+-]+?)\s*?(?:\s([a-zA-Z0-9_.-/]+))?$`)

func readRepoRules(c *content) (stateFn, error) {
	t := strings.TrimSpace(c.s.Text())
	//fmt.Printf("readRepoRules '%v'\n", t)
	//rules := []*Rule{}
	config := c.gtl.configs[len(c.gtl.configs)-1]
	for keepReading := true; keepReading; {
		res := readRepoRuleRx.FindStringSubmatchIndex(t)
		//fmt.Println(res, ">'"+t+"'")
		if res == nil {
			break
		}

		rule := &Rule{}
		pre := strings.TrimSpace(t[res[2]:res[3]])
		post := strings.TrimSpace(t[res[4]:res[5]])

		respre := repoRulePreRx.FindStringSubmatchIndex(pre)
		//fmt.Printf("\nrespre='%v' for '%v'\n", respre, pre)
		if respre == nil {
			return nil, ParseError{msg: fmt.Sprintf("Incorrect access rule '%v' at line %v ('%v')", pre, c.l, t)}
		}
		rule.access = pre[respre[2]:respre[3]]
		if respre[4] > -1 {
			rule.param = pre[respre[4]:respre[5]]
		}

		users := strings.Split(post, " ")
		for _, username := range users {
			user := &User{name: username}
			c.gtl.users = append(c.gtl.users, user)
			rule.users = append(rule.users, user)
			if grps, ok := c.gtl.namesToGroups[username]; ok {
				for _, grp := range grps {
					if err := grp.markAsUserGroup(); err != nil {
						return nil, ParseError{msg: fmt.Sprintf("user name '%v' already used repo group at line %v ('%v')\n%v", username, c.l, t, err.Error())}
					}
				}
			}
		}

		config.rules = append(config.rules, rule)
		for _, repo := range config.repos {
			if _, ok := c.gtl.reposToConfigs[repo.name]; !ok {
				c.gtl.reposToConfigs[repo.name] = []*Config{}
			}
			seen := false
			for _, aconfig := range c.gtl.reposToConfigs[repo.name] {
				if aconfig == config {
					seen = true
					break
				}
			}
			if !seen {
				c.gtl.reposToConfigs[repo.name] = append(c.gtl.reposToConfigs[repo.name], config)
			}
		}

		if !c.s.Scan() {
			keepReading = false
			return nil, nil
		}
		c.l = c.l + 1
		t = strings.TrimSpace(c.s.Text())
	}
	return readEmptyOrCommentLines, nil
}

func (grp *Group) markAsUserGroup() error {
	//fmt.Printf("\nmarkAsUserGroup '%v'", grp)
	if grp.kind == repos {
		return fmt.Errorf("group '%v' is a repos group, not a user one", grp.name)
	}
	if grp.kind == undefined {
		grp.kind = users
		grp.container.addUsersGroup(grp)
	}
	return nil
}

// NbUsers returns the number of users (single or groups)
func (gtl *Gitolite) NbUsers() int {
	return len(gtl.users)
}

// NbGroupUsers returns the number of groups identified as users
func (gtl *Gitolite) NbGroupUsers() int {
	return len(gtl.userGroups)
}

func (gtl *Gitolite) addUsersGroup(grp *Group) {
	gtl.userGroups = append(gtl.userGroups, grp)
}

// Rules get all  rules for a given repo
func (gtl *Gitolite) Rules(reponame string) ([]*Rule, error) {
	var res []*Rule
	res = append(res, gtl.rulesRepo(reponame)...)
	if groups, ok := gtl.namesToGroups[reponame]; ok {
		for _, group := range groups {
			if group.kind != repos {
				return nil, fmt.Errorf("repo name '%v' is part of a user group '%v', not a repo one", reponame, group)
			}
			for _, reponame := range group.members {
				res = append(res, gtl.rulesRepo(reponame)...)
			}
		}
	}
	return res, nil
}

func (gtl *Gitolite) rulesRepo(name string) []*Rule {
	var res []*Rule
	if configs, ok := gtl.reposToConfigs[name]; ok {
		for _, config := range configs {
			res = append(res, config.rules...)
		}
	}
	return res
}
