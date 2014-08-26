package gitolite

import (
	"fmt"
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
}

// NewGitolite creates an empty gitolite config
func NewGitolite() *Gitolite {
	res := &Gitolite{
		namesToGroups:  make(map[string][]*Group),
		reposToConfigs: make(map[string][]*Config),
	}
	return res
}

// Group (of repo or resources, ie people)
type Group struct {
	name      string
	members   []string
	kind      kind
	container Container
	cmt       *Comment
	users     []*User
}

type kind int

const (
	undefined = iota
	users
	repos
)

// IsUsers checks if the current group has been marked for users group
func (grp *Group) IsUsers() bool {
	return grp.kind == users
}

// IsUndefined checks if the current group hasn't been marked yet
func (grp *Group) IsUndefined() bool {
	return grp.kind == undefined
}

// Container contains group (of repos or users) and users
type Container interface {
	addReposGroup(grp *Group)
	addUsersGroup(grp *Group)
	addUser(user *User)
	GetUsers() []*User
}

// Repo (single or group name)
type Repo struct {
	name string
}

// Comment groups empty or lines with #
type Comment struct {
	comments []string
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

// Rule (of access to repo)
type Rule struct {
	access        string
	param         string
	usersOrGroups []userOrGroup
	cmt           *Comment
}

// UserOrGroup represents a User or a Group. Used by Rule.
type userOrGroup interface {
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

// GetUsers returns the users of a rule (including the ones in a user group set for that rule)
func (rule *Rule) GetUsers() []*User {
	res := []*User{}
	for _, uog := range rule.usersOrGroups {
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

func (rule *Rule) addUser(user *User) {
	rule.usersOrGroups = append(rule.usersOrGroups, user)
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
	res := fmt.Sprintf("group '%v'%v: %+v", grp.name, grp.kind.String(), grp.GetMembers())
	return res
}

// Groups get the groups of a gitolite config
func (gtl *Gitolite) Groups() []*Group {
	return gtl.groups
}
func (gtl *Gitolite) GetGroup(groupname string) *Group {
	for _, group := range gtl.groups {
		if group.name == groupname {
			return group
		}
	}
	return nil
}

// GetConfigs return config for a given repo
func (gtl *Gitolite) GetConfigsForRepo(reponame string) []*Config {
	return gtl.GetConfigsForRepos([]string{reponame})
}

// GetConfigs return config for a given list of repos
func (gtl *Gitolite) GetConfigsForRepos(reponames []string) []*Config {
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

func (cmt *Comment) String() string {
	res := ""
	for _, comment := range cmt.comments {
		res = res + comment + "\n"
	}
	return res
}

// NbGroupRepos returns the number of groups identified as repos
func (gtl *Gitolite) NbGroupRepos() int {
	return len(gtl.repoGroups)
}

func (cfg *Config) String() string {
	res := fmt.Sprintf("config %+v => %+v", cfg.repos, cfg.rules)
	return res
}

func (rule *Rule) String() string {
	users := ""
	if len(rule.usersOrGroups) > 0 {
		users = "="
	}
	first := true
	for _, userOrGroup := range rule.usersOrGroups {
		if !first {
			users = users + ","
		}
		users = users + " " + userOrGroup.GetName()
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
		first = false
	}
	users = strings.TrimSpace(users)
	return strings.TrimSpace(fmt.Sprintf("%v %v %v", rule.access, rule.param, users))
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
func (r *Repo) String() string {
	return fmt.Sprintf("repo '%v'", r.name)
}
func (usr *User) String() string {
	return fmt.Sprintf("user '%v'", usr.name)
}

// IsEmpty checks if config includes any repo or groups
func (gtl *Gitolite) IsEmpty() bool {
	return (gtl.groups == nil || len(gtl.groups) == 0) && (gtl.repos == nil || len(gtl.repos) == 0)
}

// NbGroup returns the number of groups (people or repos)
func (gtl *Gitolite) NbGroup() int {
	return len(gtl.groups)
}

// NbRepos returns the number of repos (single or groups)
func (gtl *Gitolite) NbRepos() int {
	return len(gtl.repos)
}

func (rule *Rule) addGroup(group *Group) {
	notFound := true
	for _, uog := range rule.usersOrGroups {
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

func (gtl *Gitolite) addReposGroup(grp *Group) {
	gtl.repoGroups = append(gtl.repoGroups, grp)
	for _, reponame := range grp.GetMembers() {
		addRepoFromName(gtl, reponame, gtl)
	}
}

// MarkAsRepoGroup makes sure a group is a repo group
func (grp *Group) MarkAsRepoGroup() error {
	if grp.kind == users {
		return fmt.Errorf("group '%v' is a users group, not a repo one", grp.name)
	}
	if grp.kind == undefined {
		grp.kind = repos
		grp.container.addReposGroup(grp)
	}
	return nil
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

func AddUserFromName(uc userContainer, username string, allUsersCtn userContainer) {
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

func (grp *Group) MarkAsUserGroup() error {
	//fmt.Printf("\nmarkAsUserGroup '%v'", grp)
	if grp.kind == repos {
		return fmt.Errorf("group '%v' is a repos group, not a user one", grp.name)
	}
	if grp.kind == undefined {
		grp.kind = users
		grp.container.addUsersGroup(grp)
	}
	for _, member := range grp.GetMembers() {
		AddUserFromName(grp, member, grp.container)
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

// AddComment adds a new line of comment to the current set.
func (cmt *Comment) AddComment(comment string) {
	comment = strings.TrimSpace(comment)
	if comment != "" {
		cmt.comments = append(cmt.comments, comment)
	}
}

// Rules are the rules listed in a config
func (cfg *Config) Rules() []*Rule {
	return cfg.rules
}

// Access part of a Rule: R, RW, RW+,...
func (rule *Rule) Access() string {
	return rule.access
}

// Param is the parameter of a rule, like a refspec, or VREF
func (rule *Rule) Param() string {
	return rule.param
}

// HasAnyUserOrGroup checks if a rule reference any user or group.
func (rule *Rule) HasAnyUserOrGroup() bool {
	return len(rule.usersOrGroups) > 0
}

// AddUserGroup adds a user group to a gitolite config
func (gtl *Gitolite) AddUserGroup(grpname string, grpmembers []string, currentComment *Comment) error {
	grp := &Group{name: grpname, members: grpmembers, container: gtl, cmt: currentComment}
	for _, g := range gtl.Groups() {
		if g.GetName() == grpname {
			return fmt.Errorf("Duplicate group name '%v'", grpname)
		}
	}
	seen := map[string]bool{}
	for _, val := range grpmembers {
		if _, ok := seen[val]; !ok {
			seen[val] = true
		} else {
			return fmt.Errorf("Duplicate group element name '%v'", val)
		}
		gtl.namesToGroups[val] = append(gtl.namesToGroups[val], grp)
	}
	gtl.groups = append(gtl.groups, grp)
	gtl.namesToGroups[grpname] = append(gtl.namesToGroups[grpname], grp)
	return nil
}

// AddConfig adds a new config and returns it,
// unless a repo name is used as user group, or is an undefined group name
func (gtl *Gitolite) AddConfig(rpmembers []string, comment *Comment) (*Config, error) {
	config := &Config{repos: []*Repo{}, cmt: comment}
	for _, rpname := range rpmembers {
		if !strings.HasPrefix(rpname, "@") {
			addRepoFromName(config, rpname, gtl)
			addRepoFromName(gtl, rpname, gtl)
			if grps, ok := gtl.namesToGroups[rpname]; ok {
				for _, grp := range grps {
					if err := grp.MarkAsRepoGroup(); err != nil {
						return nil, fmt.Errorf("repo name '%v' already used in a user group\n%v", rpname, err.Error())
					}
				}
			}
		} else {
			var group *Group
			for _, g := range gtl.groups {
				if g.name == rpname {
					group = g
					break
				}
			}
			if group == nil {
				return nil, fmt.Errorf("repo group name '%v' undefined", rpname)
			}
			//fmt.Printf("\n%v\n", group)
			group.MarkAsRepoGroup()
			for _, rpname := range group.GetMembers() {
				addRepoFromName(gtl, rpname, gtl)
				addRepoFromName(config, rpname, gtl)
			}
		}
	}
	gtl.configs = append(gtl.configs, config)
	return config, nil
}

// SetDesc set description for a config, unless there is already one.
func (cfg *Config) SetDesc(desc string, comment *Comment) error {
	if cfg.desc != "" {
		return fmt.Errorf("No more than one desc per config")
	}
	cfg.descCmt = comment
	cfg.desc = desc
	return nil
}

func (gtl *Gitolite) AddUserToRule(rule *Rule, username string) error {
	AddUserFromName(rule, username, gtl)
	AddUserFromName(gtl, username, gtl)
	if grps, ok := gtl.namesToGroups[username]; ok {
		for _, grp := range grps {
			if err := grp.MarkAsUserGroup(); err != nil {
				return fmt.Errorf("user name '%v' already used in a repo group\n%v", username, err.Error())
			}
		}
	}
	return nil
}

// Desc get description for a config, empty string if there is none.
func (cfg *Config) Desc() string {
	return cfg.desc
}

func (gtl *Gitolite) AddUserGroupToRule(rule *Rule, username string) error {
	var group *Group
	for _, g := range gtl.Groups() {
		if g.GetName() == username {
			group = g
			break
		}
	}
	if group == nil {
		group = &Group{name: username, container: gtl}
		group.MarkAsUserGroup()
	}
	if group.kind == repos {
		return fmt.Errorf("user group '%v' named after a repo group", username)
	}
	if group.kind == undefined {
		group.MarkAsUserGroup()
	}
	for _, username := range group.GetMembers() {
		AddUserFromName(gtl, username, gtl)
		rule.addGroup(group)
	}
	return nil
}

func NewRule(access, param string, comment *Comment) *Rule {
	res := &Rule{access: access, param: param, cmt: comment}
	return res
}

func (gtl *Gitolite) AddRuleToConfig(rule *Rule, config *Config) {
	config.rules = append(config.rules, rule)
	for _, repo := range config.repos {
		if _, ok := gtl.reposToConfigs[repo.name]; !ok {
			gtl.reposToConfigs[repo.name] = []*Config{}
		}
		seen := false
		for _, aconfig := range gtl.reposToConfigs[repo.name] {
			if aconfig == config {
				seen = true
				break
			}
		}
		if !seen {
			gtl.reposToConfigs[repo.name] = append(gtl.reposToConfigs[repo.name], config)
		}
	}
}

// Print prints a Gitolite with reformat.
func (gtl *Gitolite) Print() string {
	res := ""
	for _, group := range gtl.groups {
		res = res + group.Print()
	}
	for _, config := range gtl.GetConfigsForRepo("gitolite-admin") {
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
	for _, member := range grp.GetMembers() {
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
	for _, rule := range cfg.Rules() {
		res = res + rule.Print()
	}
	return res
}

// Print prints the comments and access/params and user or groups of a rule
func (rule *Rule) Print() string {
	res := rule.cmt.Print()
	res = res + rule.Access()
	if rule.Param() != "" {
		res = res + " " + rule.Param()
	}
	res = res + " ="
	for _, userOrGroup := range rule.usersOrGroups {
		res = res + " " + userOrGroup.GetName()
	}
	res = res + "\n"
	return res
}

// NBConfigs returns the number of configs
func (gtl *Gitolite) NbConfigs() int {
	return len(gtl.configs)
}

// Comment returns comment associated with Group
func (grp *Group) Comment() *Comment {
	return grp.cmt
}

// Comment returns comment associated with Config
func (cfg *Config) Comment() *Comment {
	return cfg.cmt
}

// Comment returns comment associated with Rule
func (rule *Rule) Comment() *Comment {
	return rule.cmt
}
