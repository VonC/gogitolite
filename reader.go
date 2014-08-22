package gogitolite

import (
	"bufio"
	"fmt"
	"io"
	"regexp"
	"sort"
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
	projects       []*Project
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
	cmt       *Comment
	users     []*User
}

func (k kind) String() string {
	if k == repos {
		return "<repos>"
	}
	if k == users {
		return "<users>"
	}
	return "[undefined]"
}

func (grp *Group) String() string {
	res := fmt.Sprintf("group '%v'%v: %+v", grp.name, grp.kind.String(), grp.members)
	return res
}

func (gtl *Gitolite) getGroup(groupname string) *Group {
	for _, group := range gtl.groups {
		if group.name == groupname {
			return group
		}
	}
	return nil
}

// GetConfigs return config for a given list of repos
func (gtl *Gitolite) GetConfigs(reponames []string) []*Config {
	res := []*Config{}
	if len(reponames) == 0 {
		return res
	}
	for _, config := range gtl.configs {
		rpn := []string{}
		for _, repo := range config.repos {
			for _, reponame := range reponames {
				if repo.name == reponame {
					rpn = append(rpn, reponame)
				}
			}
		}
		if len(rpn) == len(reponames) {
			res = append(res, config)
		}
	}
	return res
}

type container interface {
	addReposGroup(grp *Group)
	addUsersGroup(grp *Group)
	addUser(user *User)
	GetUsers() []*User
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

var test = ""

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
	if err == nil && test != "ignorega" {
		configs := res.GetConfigs([]string{"gitolite-admin"})
		if len(configs) != 1 {
			err = fmt.Errorf("There must be one and only gitolite-admin repo config")
			return res, err
		}
		config := configs[0]
		if len(config.rules) == 0 {
			err = fmt.Errorf("There must be at least one rule for gitolite-admin repo config")
			return res, err
		}
		rule := config.rules[0]
		if rule.access != "RW+" || rule.param != "" {
			err = fmt.Errorf("First rule for gitolite-admin repo config must be 'RW+', empty param, instead of '%v'-'%v'", rule.access, rule.param)
			return res, err
		}
		if len(rule.usersOrGroups) == 0 {
			err = fmt.Errorf("First rule for gitolite-admin repo must have at least one user or group of users")
			return res, err
		}
	}
	//fmt.Printf("\nGitolite res='%v'\n", res)
	return res, err
}

func (gtl *Gitolite) String() string {
	res := fmt.Sprintf("NbGroups: %v [", len(gtl.groups))
	for i, group := range gtl.groups {
		if i > 0 {
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
		if i > 0 {
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
	names := make([]string, 0, len(gtl.namesToGroups))
	for i := range gtl.namesToGroups {
		names = append(names, i)
	}
	sort.Strings(names)
	first := true
	for _, name := range names {
		groups := gtl.namesToGroups[name]
		if !first {
			res = res + ", "
		}
		first = false
		res = res + fmt.Sprintf("%v => %+v", name, groups)
	}
	res = res + "]\n"
	res = res + fmt.Sprintf("reposToConfigs: %v [", len(gtl.reposToConfigs))
	reponames := make([]string, 0, len(gtl.reposToConfigs))
	for i := range gtl.reposToConfigs {
		reponames = append(reponames, i)
	}
	sort.Strings(reponames)
	first = true
	for _, reponame := range reponames {
		config := gtl.reposToConfigs[reponame]
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
	t := c.s.Text()
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
			currentComment.addComment(t)
			t = c.s.Text()
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
	grp := &Group{name: grpname, members: grpmembers, container: c.gtl, cmt: currentComment}
	currentComment = &Comment{}
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
	config := &Config{repos: []*Repo{}, cmt: currentComment}
	currentComment = &Comment{}
	c.gtl.configs = append(c.gtl.configs, config)
	for _, rpname := range rpmembers {
		if !strings.HasPrefix(rpname, "@") {
			addRepoFromName(config, rpname, c.gtl)
			addRepoFromName(c.gtl, rpname, c.gtl)
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
			//fmt.Printf("\n%v\n", group)
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
func (usr *User) String() string {
	return fmt.Sprintf("user '%v'", usr.name)
}

type repoContainer interface {
	getRepos() []*Repo
	addRepo(repo *Repo)
}
type userContainer interface {
	GetUsers() []*User
	addUser(user *User)
}

func (gtl *Gitolite) getRepos() []*Repo {
	return gtl.repos
}
func (gtl *Gitolite) addRepo(repo *Repo) {
	gtl.repos = append(gtl.repos, repo)
}

// GetUsers returns all users found in a gitolite config
func (gtl *Gitolite) GetUsers() []*User {
	return gtl.users
}
func (gtl *Gitolite) addUser(user *User) {
	gtl.users = append(gtl.users, user)
}

func (cfg *Config) getRepos() []*Repo {
	return cfg.repos
}
func (cfg *Config) addRepo(repo *Repo) {
	cfg.repos = append(cfg.repos, repo)
}

func (rule *Rule) getUsersOrGroups() []UserOrGroup {
	return rule.usersOrGroups
}
func (rule *Rule) addUser(user *User) {
	rule.usersOrGroups = append(rule.usersOrGroups, user)
}
func (rule *Rule) addGroup(group *Group) {
	notFound := true
	for _, uog := range rule.getUsersOrGroups() {
		rulegrp := uog.Group()
		if rulegrp != nil && rulegrp.GetName() == group.GetName() {
			notFound = false
			break
		}
	}
	if notFound {
		rule.usersOrGroups = append(rule.usersOrGroups, group)
	}
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

func addUserFromName(uc userContainer, username string, allUsersCtn userContainer) {
	var user *User
	for _, u := range allUsersCtn.GetUsers() {
		if u.name == username {
			user = u
		}
	}
	if user == nil {
		user = &User{name: username}
		if uc != allUsersCtn {
			allUsersCtn.addUser(user)
		}
	}
	seen := false
	for _, auser := range uc.GetUsers() {
		if auser.name == user.name {
			seen = true
			break
		}
	}
	if !seen {
		uc.addUser(user)
	}

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
	repos   []*Repo
	rules   []*Rule
	descCmt *Comment
	desc    string
	cmt     *Comment
}

func (cfg *Config) String() string {
	res := fmt.Sprintf("config %+v => %+v", cfg.repos, cfg.rules)
	return res
}

// UserOrGroup represents a User or a Group. Used by Rule.
type UserOrGroup interface {
	GetName() string
	GetMembers() []string
	User() *User
	Group() *Group
	String() string
}

// User help a UserOrGroup to know it is a User
func (usr *User) User() *User {
	return usr
}

// Group help a UserOrGroup to know it is a Group
func (grp *Group) Group() *Group {
	return grp
}

// Group help a UserOrGroup to know it is *not* a Group
func (usr *User) Group() *Group {
	return nil
}

// User help a UserOrGroup to know it is *not* a User
func (grp *Group) User() *User {
	return nil
}

// GetName helps a UserOrGroup to access its name ('name' for User, or '@name' for group)
func (usr *User) GetName() string {
	return usr.name
}

// GetMembers helps a UserOrGroup to get all its users (itself for Users, its members for a group)
func (usr *User) GetMembers() []string {
	return []string{}
}

// GetName helps a UserOrGroup to access its name ('name' for User, or '@name' for group)
func (grp *Group) GetName() string {
	return grp.name
}

// GetMembers helps a UserOrGroup to get all its users (itself for Users, its members for a group)
func (grp *Group) GetMembers() []string {
	return grp.members
}

// GetUsers returns the users of a user group, or an empty list for a repo group
func (grp *Group) GetUsers() []*User {
	if grp.kind == users {
		return grp.users
	}
	return []*User{}
}

// Rule (of access to repo)
type Rule struct {
	access        string
	param         string
	usersOrGroups []UserOrGroup
	cmt           *Comment
}

// GetUsers returns the users of a rule (including the ones in a user group set for that rule)
func (rule *Rule) GetUsers() []*User {
	res := []*User{}
	for _, uog := range rule.getUsersOrGroups() {
		if uog.User() != nil {
			res = append(res, uog.User())
		}
		if uog.Group() != nil {
			grp := uog.Group()
			//fmt.Println(grp)
			for _, usr := range grp.GetUsers() {
				//fmt.Println(usr)
				res = append(res, usr)
			}
		}
	}
	return res
}

func (rule *Rule) String() string {
	users := ""
	for _, userOrGroup := range rule.usersOrGroups {
		users = " " + userOrGroup.GetName()
		members := userOrGroup.GetMembers()
		if len(members) > 0 {
			users = users + " ("
			first := true
			for _, member := range members {
				if !first {
					users = users + ", "
				}
				users = users + member
				first = false
			}
			users = users + ")"
		}
	}
	users = strings.TrimSpace(users)
	return strings.TrimSpace(fmt.Sprintf("%v %v %v", rule.access, rule.param, users))
}

var readRepoRuleRx = regexp.MustCompile(`(?m)^\s*?([^@=]+)\s*?=\s*?((?:@?[a-zA-Z0-9_-]+\s*?)+)$`)
var repoRulePreRx = regexp.MustCompile(`(?m)^([RW+-]+?)\s*?(?:\s([a-zA-Z0-9_.-/]+))?$`)
var repoRuleDescRx = regexp.MustCompile(`(?m)^desc\s*?=\s*?(\S.*?)$`)

func readRepoRules(c *content) (stateFn, error) {
	t := strings.TrimSpace(c.s.Text())
	//fmt.Printf("readRepoRules '%v'\n", t)
	//rules := []*Rule{}
	config := c.gtl.configs[len(c.gtl.configs)-1]
	for keepReading := true; keepReading; {
		readComment := false
		readDesc := false
		res := repoRuleDescRx.FindStringSubmatchIndex(t)
		//fmt.Println(res, ">0'"+t+"'")
		if res != nil && len(res) > 0 {
			if config.desc != "" {
				return nil, ParseError{msg: fmt.Sprintf("No more than one desc per config, line %v ('%v')", c.l, t)}
			}
			config.descCmt = currentComment
			currentComment = &Comment{}
			config.desc = strings.TrimSpace(t[res[2]:res[3]])
			readDesc = true
		} else {
			res = readRepoRuleRx.FindStringSubmatchIndex(t)
			//fmt.Println(res, ">1'"+t+"'")
			if res == nil {
				res = readEmptyOrCommentLinesRx.FindStringSubmatchIndex(t)
				if res != nil && len(res) > 0 {
					readComment = true
					currentComment.addComment(t)
				} else {
					if len(config.rules) == 0 {
						return nil, ParseError{msg: fmt.Sprintf("At least one access rule expected at line %v ('%v')", c.l, t)}
					}
					break
				}
			}
		}
		if !readComment && !readDesc {
			rule := &Rule{cmt: currentComment}
			currentComment = &Comment{}
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
				if !strings.HasPrefix(username, "@") {
					addUserFromName(rule, username, c.gtl)
					addUserFromName(c.gtl, username, c.gtl)
					if grps, ok := c.gtl.namesToGroups[username]; ok {
						for _, grp := range grps {
							if err := grp.markAsUserGroup(); err != nil {
								return nil, ParseError{msg: fmt.Sprintf("user name '%v' already used repo group at line %v ('%v')\n%v", username, c.l, t, err.Error())}
							}
						}
					}
				} else {
					var group *Group
					for _, g := range c.gtl.groups {
						if g.name == username {
							group = g
							break
						}
					}
					if group == nil {
						group = &Group{name: username, container: c.gtl}
						group.markAsUserGroup()
					}
					if group.kind == repos {
						return nil, ParseError{msg: fmt.Sprintf("user group '%v' named after a repo group at line %v ('%v')", username, c.l, t)}
					}
					if group.kind == undefined {
						group.markAsUserGroup()
					}
					for _, username := range group.members {
						addUserFromName(c.gtl, username, c.gtl)
						rule.addGroup(group)
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
	for _, member := range grp.members {
		addUserFromName(grp, member, grp.container)
	}
	return nil
}

func (grp *Group) addUser(user *User) {
	grp.users = append(grp.users, user)
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
	return res, nil
}

func (gtl *Gitolite) rulesRepo(reponame string) []*Rule {
	var res []*Rule
	if configs, ok := gtl.reposToConfigs[reponame]; ok {
		for _, config := range configs {
			res = append(res, config.rules...)
		}
	}
	//fmt.Printf("\nrulesRepo for rpname '%v': %v\n", reponame, res)
	return res
}

// Comment groups empty or lines with #
type Comment struct {
	comments []string
}

var currentComment *Comment

func (cmt *Comment) addComment(comment string) {
	comment = strings.TrimSpace(comment)
	if comment != "" {
		cmt.comments = append(cmt.comments, comment)
	}
}

func (cmt *Comment) String() string {
	res := ""
	for _, comment := range cmt.comments {
		res = res + comment + "\n"
	}
	return res
}

// Print prints a Gitolite with reformat.
func (gtl *Gitolite) Print() string {
	res := ""
	for _, group := range gtl.groups {
		res = res + group.Print()
	}
	for _, config := range gtl.GetConfigs([]string{"gitolite-admin"}) {
		res = res + config.Print()
	}
	for _, config := range gtl.configs {
		skip := false
		for _, repo := range config.getRepos() {
			if repo.name == "gitolite-admin" {
				skip = true
			}
		}
		if !skip {
			res = res + config.Print()
		}
	}
	return res
}

// Print prints the comments (empty string if no comments)
func (cmt *Comment) Print() string {
	res := ""
	for _, comment := range cmt.comments {
		res = res + comment + "\n"
	}
	return res
}

// Print prints a Group of repos/users with reformat.
func (grp *Group) Print() string {
	res := grp.cmt.Print()
	res = res + grp.name + " ="
	for _, member := range grp.members {
		m := strings.TrimSpace(member)
		if m != "" {
			res = res + " " + m
		}
	}
	res = res + "\n"
	return res
}

// Print prints a Config with reformat.
func (cfg *Config) Print() string {
	res := cfg.cmt.Print()
	res = res + "repo"
	for _, repo := range cfg.repos {
		res = res + " " + repo.name
	}
	res = res + "\n"
	if cfg.desc != "" {
		if cfg.descCmt != nil {
			res = res + cfg.descCmt.Print()
		}
		res = res + "desc = " + cfg.desc + "\n"
	}
	for _, rule := range cfg.rules {
		res = res + rule.Print()
	}
	return res
}

// Print prints the comments and access/params and user or groups of a rule
func (rule *Rule) Print() string {
	res := rule.cmt.Print()
	res = res + rule.access
	if rule.param != "" {
		res = res + " " + rule.param
	}
	res = res + " ="
	for _, userOrGroup := range rule.usersOrGroups {
		res = res + " " + userOrGroup.GetName()
	}
	res = res + "\n"
	return res
}
