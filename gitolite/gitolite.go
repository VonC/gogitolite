package gitolite

import (
	"fmt"
	"regexp"
	"strings"
)

// Gitolite config decoded
type Gitolite struct {
	groups        []*Group
	reposOrGroups []RepoOrGroup
	usersOrGroups []UserOrGroup
	configs       []*Config
	subconfs      []*regexp.Regexp
	parent        *Gitolite
	elts          []Printable
}

// Printable is an element which can be printed
type Printable interface {
	Print() string
}

// NewGitolite creates an empty gitolite config
func NewGitolite(parent *Gitolite) *Gitolite {
	res := &Gitolite{
		parent: parent,
	}
	return res
}

// Configs gets gitolite configs
func (gtl *Gitolite) Configs() []*Config {
	return gtl.configs
}

// Group (of repo or resources, ie people)
type Group struct {
	name          string
	members       []string
	kind          kind
	container     Container
	cmt           *Comment
	usersOrGroups []UserOrGroup
	reposOrGroups []RepoOrGroup
}

// AddSubconf adds a new subconf regexp to the gitolite configuration
// Duplicate is ignored
// If regexp doesn't compile (by replacing * with '.*'), return the error
func (gtl *Gitolite) AddSubconf(subconf string) error {
	subconf = strings.Replace(subconf, "*", ".*", -1)
	r, err := regexp.Compile(subconf)
	if err != nil {
		return err
	}
	seen := false
	for _, sc := range gtl.subconfs {
		if sc.String() == r.String() {
			seen = true
			break
		}
	}
	if !seen {
		gtl.subconfs = append(gtl.subconfs, r)
	}
	return nil
}

// Subconfs returns the subconf regexps read in the gitolite.conf
func (gtl *Gitolite) Subconfs() []*regexp.Regexp {
	return gtl.subconfs
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
	addRepoOrGroup(rog RepoOrGroup)
	addUserOrGroup(uog UserOrGroup)
	GetReposOrGroups() []RepoOrGroup
	GetUsersOrGroups() []UserOrGroup
}

// Repo (single or group name)
type Repo struct {
	name string
}

// Comment groups empty or lines with #
type Comment struct {
	comments []string
	sameLine string
}

// User (or group of users)
type User struct {
	name string
}

// Config for repos with access rules
type Config struct {
	reposOrGroups []RepoOrGroup
	rules         []*Rule
	descCmt       *Comment
	desc          string
	cmt           *Comment
}

// Rule (of access to repo)
type Rule struct {
	access        string
	param         string
	usersOrGroups []UserOrGroup
	cmt           *Comment
	space         int
	pspace        int
}

func (rule *Rule) maxSpace() (int, int) {
	s := len(rule.Access())
	if s < 5 {
		s = 5
	}
	ps := len(rule.Param())
	return s, ps
}

// GetUsersOrGroups returns the users or groups of users associated to the rule
func (rule *Rule) GetUsersOrGroups() []UserOrGroup {
	return rule.usersOrGroups
}

// UserOrGroup represents a User or a Group. Used by Rule and Group.
type UserOrGroup interface {
	GetName() string
	GetMembers() []string
	User() *User
	Group() *Group
	String() string
}

// RepoOrGroup represents a Repo or a Group. Used by Group.
type RepoOrGroup interface {
	GetName() string
	GetMembers() []string
	Repo() *Repo
	Group() *Group
	String() string
}

// Repo help a RepoOrGroup to know it is a Repo
func (repo *Repo) Repo() *Repo {
	return repo
}

// User help a UserOrGroup to know it is a User
func (usr *User) User() *User {
	return usr
}

// Repo help a UserOrGroup to know it is *not* a Repo
func (grp *Group) Repo() *Repo {
	return nil
}

// Group help a UserOrGroup to know it is a Group
func (grp *Group) Group() *Group {
	return grp
}

// Group help a RepoOrGroup to know it is *not* a Group
func (repo *Repo) Group() *Group {
	return nil
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

// GetName returns the name of a repo
func (repo *Repo) GetName() string {
	return repo.name
}

// GetMembers helps a RepoOrGroup to get all its repos (itself for Repos, its members for a group)
func (repo *Repo) GetMembers() []string {
	return []string{}
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

// GetUsersOrGroups returns the users of a user group, or an empty list for a repo group
func (grp *Group) GetUsersOrGroups() []UserOrGroup {
	if grp.kind == users {
		return grp.usersOrGroups
	}
	return []UserOrGroup{}
}

func addUser(users []*User, user *User) []*User {
	/*if user == nil {
		return users
	}*/
	seen := false
	for _, u := range users {
		if u.GetName() == user.GetName() {
			seen = true
		}
	}
	if !seen {
		users = append(users, user)
	}
	return users
}

func addRepo(repos []*Repo, repo *Repo) []*Repo {
	seen := false
	for _, r := range repos {
		if r.GetName() == repo.GetName() {
			seen = true
		}
	}
	if !seen {
		repos = append(repos, repo)
	}
	return repos
}

// GetAllRepos returns the repos  of a Group (including the ones in a repo group including the Group)
func (grp *Group) GetAllRepos() []*Repo {
	res := []*Repo{}
	for _, rog := range grp.GetReposOrGroups() {
		if rog.Repo() != nil {
			repo := rog.Repo()
			res = addRepo(res, repo)
		}
		if rog.Group() != nil {
			group := rog.Group()
			repos := group.GetAllRepos()
			for _, repo := range repos {
				res = addRepo(res, repo)
			}
		}
	}
	return res
}

// GetAllUsers returns the users of a Group (including the ones in a user group including the Group)
func (grp *Group) GetAllUsers() []*User {
	res := []*User{}
	for _, uog := range grp.GetUsersOrGroups() {
		if uog.User() != nil {
			user := uog.User()
			res = addUser(res, user)
		}
		if uog.Group() != nil {
			group := uog.Group()
			users := group.GetAllUsers()
			for _, user := range users {
				res = addUser(res, user)
			}
		}
	}
	return res
}

// GetAllUsers returns the users of a rule (including the ones in a user group set for that rule)
func (rule *Rule) GetAllUsers() []*User {
	res := []*User{}
	for _, uog := range rule.usersOrGroups {
		if uog.User() != nil {
			res = addUser(res, uog.User())
			//fmt.Println(uog.User())
		}
		if uog.Group() != nil {
			grp := uog.Group()
			//fmt.Println(grp)
			for _, usr := range grp.GetAllUsers() {
				//fmt.Println(usr)
				res = append(res, usr)
			}
		}
	}
	return res
}

// GetUsersFirstOrGroups gets all users or, if the user group is not defined,
// the (empty) user group.
func (rule *Rule) GetUsersFirstOrGroups() []UserOrGroup {
	res := []UserOrGroup{}
	for _, uog := range rule.usersOrGroups {
		if uog.User() != nil {
			res = append(res, uog)
		}
		if uog.Group() != nil {
			grp := uog.Group()
			users := grp.GetUsersOrGroups()
			for _, usr := range grp.GetAllUsers() {
				res = append(res, usr)
			}
			if len(users) == 0 {
				res = append(res, grp)
			}
		}
	}
	return res
}

type repoContainer interface {
	GetReposOrGroups() []RepoOrGroup
	addRepoOrGroup(rog RepoOrGroup)
}
type userContainer interface {
	GetUsersOrGroups() []UserOrGroup
	addUserOrGroup(uog UserOrGroup)
}

// GetReposOrGroups returns the repos or groups of repos found in a gitolite conf
func (gtl *Gitolite) GetReposOrGroups() []RepoOrGroup {
	return gtl.reposOrGroups
}
func (grp *Group) addRepoOrGroup(rog RepoOrGroup) {
	grp.reposOrGroups = append(grp.reposOrGroups, rog)
	grp.members = addStringNoDup(grp.members, rog.GetName())
}

// GetReposOrGroups returns the repos or groups of repos listed in a repos group
func (grp *Group) GetReposOrGroups() []RepoOrGroup {
	if grp.kind == repos {
		return grp.reposOrGroups
	}
	return []RepoOrGroup{}
}

// GetReposOrGroups returns itself since it is a repo
func (repo *Repo) GetReposOrGroups() []RepoOrGroup {
	return []RepoOrGroup{repo}
}

// GetUsersOrGroups returns all users found in a gitolite config
func (gtl *Gitolite) GetUsersOrGroups() []UserOrGroup {
	return gtl.usersOrGroups
}
func (gtl *Gitolite) addUserOrGroup(uog UserOrGroup) {
	seen := isUserOrGroupSeen(uog.GetName(), gtl.usersOrGroups)
	if !seen {
		gtl.usersOrGroups = append(gtl.usersOrGroups, uog)
	}
	grp := uog.Group()
	if grp != nil {
		grp.container = gtl
		for _, userOrGroupName := range uog.GetMembers() {
			addUserOrGroupFromName(grp, userOrGroupName, gtl)
			addUserOrGroupFromName(gtl, userOrGroupName, gtl)
		}
		grp := uog.Group()
		gtl.addGroup(grp)
	}
}

// GetReposOrGroups returns the repos oer groups of repos listed in config
func (cfg *Config) GetReposOrGroups() []RepoOrGroup {
	return cfg.reposOrGroups
}
func (cfg *Config) addRepoOrGroup(rog RepoOrGroup) {
	cfg.reposOrGroups = append(cfg.reposOrGroups, rog)
}

func (rule *Rule) addUserOrGroup(uog UserOrGroup) {
	rule.usersOrGroups = append(rule.usersOrGroups, uog)
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

// String displays group internals data (name, type and members)
func (grp *Group) String() string {
	res := fmt.Sprintf("group '%v'%v: %+v", grp.name, grp.kind.String(), grp.GetMembers())
	return res
}

// GetGroup get group for a given group name
func (gtl *Gitolite) GetGroup(groupname string) *Group {
	for _, group := range gtl.groups {
		if group.name == groupname {
			return group
		}
	}
	return nil
}

// GetConfigsForRepo return config for a given repo name
func (gtl *Gitolite) GetConfigsForRepo(reponame string) []*Config {
	return gtl.GetConfigsForRepos([]string{reponame})
}

func (grp *Group) hasRepoOrGroup(rogname string) bool {
	for _, rog := range grp.reposOrGroups {
		if rog.GetName() == rogname {
			return true
		} else if rog.Group() != nil {
			subgrp := rog.Group()
			if subgrp.hasRepoOrGroup(rogname) {
				return true
			}
		}
	}
	return false
}

// GetConfigsForRepos return config for a given list of repos
func (gtl *Gitolite) GetConfigsForRepos(reponames []string) []*Config {
	res := []*Config{}
	if len(reponames) == 0 {
		return res
	}
	for _, config := range gtl.configs {
		rpn := []string{}
		for _, rog := range config.reposOrGroups {
			for _, reponame := range reponames {
				if rog.GetName() == reponame {
					rpn = append(rpn, reponame)
				} else if rog.Group() != nil {
					grp := rog.Group()
					if grp.hasRepoOrGroup(reponame) {
						rpn = append(rpn, reponame)
					}
				}
			}
		}
		if len(rpn) == len(reponames) {
			res = append(res, config)
		}
	}
	return res
}

// String exposes comment lines
func (cmt *Comment) String() string {
	res := ""
	for _, comment := range cmt.comments {
		res = res + comment + "\n"
	}
	return res
}

// NbRepoGroups returns the number of groups identified as repos
func (gtl *Gitolite) NbRepoGroups() int {
	res := 0
	for _, grp := range gtl.groups {
		if grp.kind == repos {
			res = res + 1
		}
	}
	return res
}

// NbUserGroups returns the number of groups identified as users
func (gtl *Gitolite) NbUserGroups() int {
	res := 0
	for _, grp := range gtl.groups {
		if grp.kind == users {
			res = res + 1
		}
	}
	return res
}

// NbUsers returns the number of users
func (gtl *Gitolite) NbUsers() int {
	res := 0
	for _, uog := range gtl.usersOrGroups {
		if uog.User() != nil {
			res = res + 1
		}
	}
	return res
}

// NbRepos returns the number of repos
func (gtl *Gitolite) NbRepos() int {
	res := 0
	for _, rog := range gtl.reposOrGroups {
		if rog.Repo() != nil {
			res = res + 1
		}
	}
	return res
}

// String exposes Config internals (each repos and rules)
func (cfg *Config) String() string {
	res := fmt.Sprintf("config %+v => rules %+v", cfg.reposOrGroups, cfg.rules)
	return res
}

// String exposes Rule internals (each users)
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

// String exposes Gitolite internals (including maps like namesToGroups and reposToConfigs)
func (gtl *Gitolite) String() string {
	res := fmt.Sprintf("NbGroups: %v [", len(gtl.groups))
	for i, group := range gtl.groups {
		if i > 0 {
			res = res + ", "
		}
		res = res + group.name
	}
	res = res + "]\n"

	res = res + fmt.Sprintf("NbRepoOrGroups: %v [", len(gtl.reposOrGroups))
	for i, rog := range gtl.reposOrGroups {
		if i > 0 {
			res = res + ", "
		}
		res = res + rog.GetName()
	}
	res = res + "]\n"

	res = res + fmt.Sprintf("NbUserOrGroups: %v [", len(gtl.usersOrGroups))
	for i, uog := range gtl.usersOrGroups {
		if i > 0 {
			res = res + ", "
		}
		res = res + uog.GetName()
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
	return res
}

// IsNakedRW check if rule as RW without any param
func (rule *Rule) IsNakedRW() bool {
	return rule.Access() == "RW" && rule.Param() == ""
}

// String exposes Repo internals (its name)
func (repo *Repo) String() string {
	return fmt.Sprintf("repo '%v'", repo.name)
}

// String exposes User internals (its name)
func (usr *User) String() string {
	return fmt.Sprintf("user '%v'", usr.name)
}

// IsEmpty checks if config includes any repo or groups
func (gtl *Gitolite) IsEmpty() bool {
	return !(len(gtl.groups) > 0 || len(gtl.usersOrGroups) > 0 || len(gtl.reposOrGroups) > 0)
}

// NbGroup returns the number of groups (people or repos)
func (gtl *Gitolite) NbGroup() int {
	return len(gtl.groups)
}

// NbReposOrGroups returns the number of repos (single or groups)
func (gtl *Gitolite) NbReposOrGroups() int {
	return len(gtl.reposOrGroups)
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

func (gtl *Gitolite) addGroup(grp *Group) {
	seen := false
	var gg *Group
	for _, group := range gtl.groups {
		if grp.GetName() == group.GetName() {
			seen = true
			gg = group
			break
		}
	}
	if !seen {
		gtl.groups = append(gtl.groups, grp)
	} else {
		//fmt.Println("\nGRP: ", grp, ", GG: ", gg)
		for _, member := range gg.GetMembers() {
			if !isNameSeen(member, grp.GetMembers()) {
				grp.members = append(grp.members, member)
			}
		}
		if grp.kind != undefined {
			gg.kind = grp.kind
		}
	}
}

func (gtl *Gitolite) addRepoOrGroup(rog RepoOrGroup) {
	seen := isRepoOrGroupSeen(rog.GetName(), gtl.reposOrGroups)
	if !seen {
		gtl.reposOrGroups = append(gtl.reposOrGroups, rog)
	}
	grp := rog.Group()
	if grp != nil {
		for _, repoOrGroupName := range rog.GetMembers() {
			addRepoOrGroupFromName(grp, repoOrGroupName, gtl)
			addRepoOrGroupFromName(gtl, repoOrGroupName, gtl)
		}
		grp := rog.Group()
		gtl.addGroup(grp)
	}
}

// MarkAsRepoGroup makes sure a group is a repo group
func (grp *Group) MarkAsRepoGroup() error {
	if grp.kind == users {
		return fmt.Errorf("group '%v' is a users group, not a repo one", grp.name)
	}
	if grp.kind == undefined {
		grp.kind = repos
	}
	grp.container.addRepoOrGroup(grp)
	for _, member := range grp.GetMembers() {
		addRepoOrGroupFromName(grp, member, grp.container)
	}
	return nil
}

// GetRepoGroup returns the repo group if found
func (gtl *Gitolite) GetRepoGroup(name string) *Group {
	for _, grp := range gtl.groups {
		//fmt.Printf("group '%v' '%v'\n", grp.GetName(), grp.GetRepos() != nil)
		if !grp.IsUsers() && grp.GetName() == name {
			return grp
		}
	}
	return nil
}

func addRepoOrGroupFromName(rc repoContainer, rogname string, allReposCtn repoContainer) {
	var rog RepoOrGroup
	for _, arog := range allReposCtn.GetReposOrGroups() {
		if arog.GetName() == rogname {
			rog = arog
		}
	}
	//fmt.Println("addRepoOrGroupFromName ", rc, " => rog ", rog, " (", rogname, ")")
	if rog == nil {
		if !strings.HasPrefix(rogname, "@") {
			rog = &Repo{name: rogname}
		} else {
			rog = &Group{name: rogname, kind: repos}
		}
		if rc != allReposCtn {
			allReposCtn.addRepoOrGroup(rog)
		}
	}
	seen := isRepoOrGroupSeen(rog.GetName(), rc.GetReposOrGroups())
	if !seen {
		rc.addRepoOrGroup(rog)
	}
}

func isUserOrGroupSeen(uogName string, uogs []UserOrGroup) bool {
	for _, uog := range uogs {
		if uogName == uog.GetName() {
			return true
		}
	}
	return false
}

func isRepoOrGroupSeen(rogName string, rogs []RepoOrGroup) bool {
	for _, rog := range rogs {
		if rogName == rog.GetName() {
			return true
		}
	}
	return false
}

func isNameSeen(name string, names []string) bool {
	for _, aname := range names {
		if name == aname {
			return true
		}
	}
	return false
}

// AddUserFromName add a user to a user container.
// If the user name doesn't match a user, creates the user
func addUserOrGroupFromName(uc userContainer, uogname string, allUsersCtn userContainer) {
	var uog UserOrGroup
	//fmt.Println("addUserOrGroupFromName ", uogname, " ")
	for _, auog := range allUsersCtn.GetUsersOrGroups() {
		if auog.GetName() == uogname {
			uog = auog
		}
	}
	if uog == nil {
		if !strings.HasPrefix(uogname, "@") {
			uog = &User{name: uogname}
		} else {
			uog = &Group{name: uogname, kind: users}
		}
		if uc != allUsersCtn {
			allUsersCtn.addUserOrGroup(uog)
		}
	}
	seen := isUserOrGroupSeen(uog.GetName(), uc.GetUsersOrGroups())
	if !seen {
		uc.addUserOrGroup(uog)
	}

}

func (grp *Group) markAsUserGroup() error {
	//fmt.Printf("\nmarkAsUserGroup '%v'", grp)
	if grp.kind == repos {
		return fmt.Errorf("group '%v' is a repos group, not a user one", grp.name)
	}
	if grp.kind == undefined {
		grp.kind = users
	}
	grp.container.addUserOrGroup(grp)
	for _, member := range grp.GetMembers() {
		addUserOrGroupFromName(grp, member, grp.container)
	}
	return nil
}

func addStringNoDup(list []string, s string) []string {
	notSeen := !isNameSeen(s, list)
	res := list
	if notSeen {
		res = append(list, s)
	}
	return res
}

func (grp *Group) addUserOrGroup(uog UserOrGroup) {
	grp.usersOrGroups = append(grp.usersOrGroups, uog)
	grp.members = addStringNoDup(grp.members, uog.GetName())
}

// NbUsersOrGroups returns the number of users (single or groups)
func (gtl *Gitolite) NbUsersOrGroups() int {
	return len(gtl.usersOrGroups)
}

// Rules get all rules for a given repo name or group of repos name
func (gtl *Gitolite) Rules(rogname string) ([]*Rule, error) {
	var res []*Rule
	res = append(res, gtl.rulesRepo(rogname)...)
	return res, nil
}

func (gtl *Gitolite) repoOrGroupFromName(rogname string) RepoOrGroup {
	for _, rog := range gtl.reposOrGroups {
		if rog.GetName() == rogname {
			return rog
		}
	}
	return nil
}

func (gtl *Gitolite) configsFromRepoOrGroup(rog RepoOrGroup) []*Config {
	res := []*Config{}
	if rog == nil {
		return res
	}
	for _, config := range gtl.configs {
		for _, arog := range config.reposOrGroups {
			if arog.GetName() == rog.GetName() {
				res = append(res, config)
				break
			} else if arog.Group() != nil {
				grp := arog.Group()
				if grp.hasRepoOrGroup(rog.GetName()) {
					res = append(res, config)
					break
				}
			}
		}
	}
	return res
}

func (gtl *Gitolite) rulesRepo(rogname string) []*Rule {
	var res []*Rule
	rog := gtl.repoOrGroupFromName(rogname)
	configs := gtl.configsFromRepoOrGroup(rog)
	//fmt.Printf("\nrulesRepo for rpname '%v': %v\n", rog, configs)
	for _, config := range configs {
		res = append(res, config.rules...)
	}
	//fmt.Printf("\nrulesRepo for rpname '%v': %v\n", rogname, res)
	return res
}

// AddComment adds a new line of comment to the current set.
func (cmt *Comment) AddComment(comment string) {
	comment = strings.TrimSpace(comment)
	cmt.comments = append(cmt.comments, comment)
}

// SameLineComment set the same line comment
func (cmt *Comment) SetSameLine(comment string) {
	comment = strings.TrimSpace(comment)
	cmt.sameLine = comment
}

func (cmt *Comment) SameLine() string {
	return cmt.sameLine
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

// AddUserOrRepoGroup adds a user or repo group to a gitolite config
func (gtl *Gitolite) AddUserOrRepoGroup(grpname string, grpmembers []string, currentComment *Comment) error {
	grp := &Group{name: grpname, members: grpmembers, container: gtl, cmt: currentComment}
	for _, g := range gtl.groups {
		if g.GetName() == grpname {
			if len(g.members) > 0 {
				return fmt.Errorf("Duplicate group name '%v'", grpname)
			}
			g.cmt = grp.cmt
			grp = g
		}
	}
	seen := map[string]bool{}
	for _, val := range grpmembers {
		if _, ok := seen[val]; !ok {
			seen[val] = true
		} else {
			return fmt.Errorf("Duplicate group element name '%v'", val)
		}
		//addRepoOrGroupFromName(grp, val)
	}
	gtl.addGroup(grp)
	gtl.elts = append(gtl.elts, grp)
	return nil
}

func (gtl *Gitolite) getGroupsForMember(memberName string) []*Group {
	res := []*Group{}
	for _, grp := range gtl.groups {
		for _, mname := range grp.members {
			if mname == memberName {
				res = append(res, grp)
				break
			}
		}
	}
	return res
}

// AddConfig adds a new config and returns it,
// unless a repo name is used as user group, or is an undefined group name
func (gtl *Gitolite) AddConfig(rpmembers []string, comment *Comment) (*Config, error) {
	config := &Config{reposOrGroups: []RepoOrGroup{}, cmt: comment}
	for _, rpname := range rpmembers {
		if !strings.HasPrefix(rpname, "@") {
			grps := gtl.getGroupsForMember(rpname)
			for _, grp := range grps {
				if err := grp.MarkAsRepoGroup(); err != nil {
					return nil, fmt.Errorf("repo name '%v' already used in a user group\n%v", rpname, err.Error())
				}
			}
			addRepoOrGroupFromName(config, rpname, gtl)
			addRepoOrGroupFromName(gtl, rpname, gtl)
		} else {
			err := gtl.addRepoGroupToConfig(config, rpname)
			if err != nil {
				return nil, err
			}
		}
	}
	gtl.configs = append(gtl.configs, config)
	gtl.elts = append(gtl.elts, config)
	return config, nil
}

func (gtl *Gitolite) getGroup(rpname string) *Group {
	var group *Group
	for _, g := range gtl.groups {
		if g.name == rpname {
			group = g
			break
		}
	}
	if group == nil && gtl.parent != nil {
		// fmt.Println("=> ", gtl.parent.groups)
		for _, g := range gtl.parent.groups {
			if g.name == rpname {
				group = g
				break
			}
		}
	}
	return group
}

func (gtl *Gitolite) addRepoGroupToConfig(config *Config, repogrpname string) error {
	group := gtl.getGroup(repogrpname)
	if group == nil {
		if repogrpname == "@all" {
			group = &Group{name: "@all", container: gtl}
		} else {
			return fmt.Errorf("repo group name '%v' undefined", repogrpname)
		}
	}
	if !isRepoOrGroupSeen(repogrpname, config.reposOrGroups) {
		config.addRepoOrGroup(group)
	}
	//fmt.Printf("\n%v\n", group)
	if err := group.MarkAsRepoGroup(); err != nil {
		return err
	}
	for _, rporgrpname := range group.GetMembers() {
		addRepoOrGroupFromName(gtl, rporgrpname, gtl)
		// no repo @grp doesn't mean repo @grep
		//addRepoOrGroupFromName(config, rporgrpname, gtl)
	}
	return nil
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

func (gtl *Gitolite) userOrGroupFromName(uogname string) UserOrGroup {
	for _, uog := range gtl.usersOrGroups {
		if uog.GetName() == uogname {
			return uog
		}
	}
	return nil
}

func (gtl *Gitolite) groupsFromUserOrGroup(uog UserOrGroup) []*Group {
	res := []*Group{}
	if uog == nil {
		return res
	}
	for _, grp := range gtl.groups {
		for _, member := range grp.members {
			if member == uog.GetName() {
				res = append(res, grp)
			}
		}
	}
	return res
}

// AddUserOrGroupToRule adds user to rule unless user name already used in a repo group
func (gtl *Gitolite) AddUserOrGroupToRule(rule *Rule, uogname string) error {
	if uogname != "@all" && isRepoOrGroupSeen(uogname, gtl.reposOrGroups) {
		return fmt.Errorf("user or user group name '%v' already used in a repo group", uogname)
	}
	addUserOrGroupFromName(rule, uogname, gtl)
	addUserOrGroupFromName(gtl, uogname, gtl)
	uog := gtl.userOrGroupFromName(uogname)
	grps := gtl.groupsFromUserOrGroup(uog)
	for _, grp := range grps {
		grp.markAsUserGroup()
	}
	if uog.Group() != nil {
		grp := uog.Group()
		//fmt.Println("\nAddUserOrGroupToRule ", grp)
		grp.markAsUserGroup()
	}
	return nil
}

// Desc get description for a config, empty string if there is none.
func (cfg *Config) Desc() string {
	return cfg.desc
}

// NewRule creates a new Rule with access, param and comment
func NewRule(access, param string, comment *Comment) *Rule {
	res := &Rule{access: access, param: param, cmt: comment}
	return res
}

// AddRuleToConfig adds rule to config and update repo to config map
func (gtl *Gitolite) AddRuleToConfig(rule *Rule, config *Config) {
	seen := false
	for _, arule := range config.rules {
		if arule == rule {
			seen = true
		}
	}
	if !seen {
		config.rules = append(config.rules, rule)
	}
}

// Print prints a Gitolite with reformat.
func (gtl *Gitolite) Print() string {
	res := ""
	for _, p := range gtl.elts {
		res = res + p.Print()
	}
	return res
}

// Print prints the comments (empty string if no comments)
func (cmt *Comment) Print() string {
	res := ""
	if cmt != nil {
		for _, comment := range cmt.comments {
			if !strings.HasPrefix(comment, "#") && comment != "" {
				res = res + "# "
			}
			res = res + comment + "\n"
		}
	}
	return res
}

// Print prints a Group of repos/users with reformat.
func (grp *Group) Print() string {
	res := grp.cmt.Print()
	if len(grp.members) > 0 || res != "" {
		res = res + grp.name + " ="
		for _, member := range grp.GetMembers() {
			m := strings.TrimSpace(member)
			if m != "" {
				res = res + " " + m
			}
		}
		res = res + "\n"
	}
	return res
}

// Print prints a Config with reformat.
func (cfg *Config) Print() string {
	res := cfg.cmt.Print()
	res = res + "repo"
	for _, rog := range cfg.reposOrGroups {
		res = res + " " + rog.GetName()
	}
	res = res + "\n"
	if cfg.desc != "" {
		if cfg.descCmt != nil {
			res = res + cfg.descCmt.Print()
		}
		res = res + "    desc  = " + cfg.desc + "\n"
	}
	maxspace := 0
	maxpspace := 0
	for _, rule := range cfg.Rules() {
		space, pspace := rule.maxSpace()
		if space > maxspace {
			maxspace = space
		}
		if pspace > maxpspace {
			maxpspace = pspace
		}
	}
	for _, rule := range cfg.Rules() {
		rule.space = maxspace
		rule.pspace = maxpspace
		res = res + rule.Print()
	}
	return res
}

// Print prints the comments and access/params and user or groups of a rule
func (rule *Rule) Print() string {
	res := rule.cmt.Print()
	f := "    %-" + fmt.Sprintf("%d", rule.space) + "s"
	res = res + fmt.Sprintf(f, rule.Access())
	f = "%-" + fmt.Sprintf("%d", rule.pspace) + "s"
	res = res + fmt.Sprintf(f, rule.Param())
	res = res + " ="
	for _, userOrGroup := range rule.usersOrGroups {
		res = res + " " + userOrGroup.GetName()
	}
	if rule.cmt != nil && rule.cmt.sameLine != "" {
		res = res + " # " + rule.cmt.sameLine
	}
	res = res + "\n"
	return res
}

// NbConfigs returns the number of configs
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
