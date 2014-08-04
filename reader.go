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
	groups       []*Group
	repos        []*Repo
	users        []*User
	namesToGroup map[string]*Group
	repoGroups   []*Group
	userGroups   []*Group
	configs      []*Config
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
	res := &Gitolite{namesToGroup: make(map[string]*Group)}
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
	return res, err
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

var readGroupRx = regexp.MustCompile(`(?m)^\s*?@([a-zA-Z0-9_-]+)\s*?=\s*?((?:[a-zA-Z0-9_-]+\s*?)+)$`)

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
		c.gtl.namesToGroup[val] = grp
	}
	c.gtl.groups = append(c.gtl.groups, grp)
	c.gtl.namesToGroup[grpname] = grp
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
		repo := &Repo{name: rpname}
		c.gtl.repos = append(c.gtl.repos, repo)
		config.repos = append(config.repos, repo)
		if grp, ok := c.gtl.namesToGroup[rpname]; ok {
			if err := grp.markAsRepoGroup(); err != nil {
				return nil, ParseError{msg: fmt.Sprintf("repo name '%v' already used user group at line %v ('%v')\n%v", rpname, c.l, t, err.Error())}
			}
		}
	}
	if !c.s.Scan() {
		return nil, nil
	}
	c.l = c.l + 1
	return readRepoRules, nil
}

func (gtl *Gitolite) addReposGroup(grp *Group) {
	gtl.repoGroups = append(gtl.repoGroups, grp)
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

// Rule (of access to repo)
type Rule struct {
	access string
	param  string
	users  []*User
}

var readRepoRuleRx = regexp.MustCompile(`(?m)^\s*?([^@=]+)\s*?=\s*?((?:[a-zA-Z0-9_-]+\s*?)+)$`)
var repoRulePreRx = regexp.MustCompile(`(?m)^([RW+-]+?)\s*?([a-zA-Z0-9_.-/]+)?$`)

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
			return nil, ParseError{msg: fmt.Sprintf("Incorrect rule '%v' at line %v ('%v')", pre, c.l, t)}
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
			if grp, ok := c.gtl.namesToGroup[username]; ok {
				if err := grp.markAsUserGroup(); err != nil {
					return nil, ParseError{msg: fmt.Sprintf("user name '%v' already used repo group at line %v ('%v')\n%v", username, c.l, t, err.Error())}
				}
			}
		}

		config.rules = append(config.rules, rule)

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
