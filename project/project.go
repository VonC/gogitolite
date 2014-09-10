package project

import (
	"fmt"
	"strings"

	"github.com/VonC/gogitolite/gitolite"
)

// Project has a name and users
type Project struct {
	name    string
	admins  []gitolite.UserOrGroup
	members []gitolite.UserOrGroup
}

func (p *Project) String() string {
	res := "project " + p.name
	return res
}

// Manager manages project for a gitolite instance
type Manager struct {
	gtl      *gitolite.Gitolite
	subconfs map[string]*gitolite.Gitolite
	projects []*Project
}

// NewManager creates a new project manager
func NewManager(gtl *gitolite.Gitolite, subconfs map[string]*gitolite.Gitolite) *Manager {
	pm := &Manager{gtl: gtl, subconfs: subconfs}
	pm.updateProjects()
	return pm
}

// Projects return the list of detected projects
func (pm *Manager) Projects() []*Project {
	return pm.projects
}

// NbProjects returns the number of detected projects
func (pm *Manager) NbProjects() int {
	return len(pm.projects)
}

var prefix = "VREF/NAME/conf/subs/"

func (pm *Manager) updateProjects() {
	gtl := pm.gtl
	configs := gtl.GetConfigsForRepo("gitolite-admin")
	for _, config := range configs {
		var currentProject *Project
		rules := config.Rules()
		var isrw bool
		for _, rule := range rules {
			//fmt.Printf("\nRule looked at: '%v' => '%v' '%v'\n", rule, rule.Access(), rule.Param())
			if rule.IsNakedRW() {
				currentProject = &Project{admins: rule.GetUsersFirstOrGroups()}
			} else if isrw, currentProject = pm.currentProjectRW(rule, currentProject); isrw {
				isrw = true
			} else if rule.Access() == "-" && rule.Param() == "VREF/NAME/" {
				if currentProject != nil && currentProject.name == "" {
					fmt.Printf("\nIgnore project with no name\n")
					currentProject = nil
				}
				currentProject = pm.currentProjectVREFName(currentProject, rule, gtl)
			} else {
				currentProject = nil
			}
		}
	}
}

func (pm *Manager) currentProjectRW(rule *gitolite.Rule, currentProject *Project) (bool, *Project) {
	var isrw = false
	if isrw = rule.Access() == "RW" && strings.HasPrefix(rule.Param(), prefix); isrw {
		projectname := rule.Param()[len(prefix):]
		if currentProject == nil {
			fmt.Printf("\nIgnore project name '%v': no RW rule before\n", projectname)
		} else {
			currentProject.name = projectname
		}
		if currentProject != nil && !currentProject.hasSameUsers(rule.GetUsersFirstOrGroups()) {
			fmt.Printf("\nIgnore project name '%v': Admins differ on 'RW' (%v vs. %v)\n", projectname,
				currentProject.admins, rule.GetUsersFirstOrGroups())
			currentProject = nil
		}
	}
	return isrw, currentProject
}

func (pm *Manager) currentProjectVREFName(currentProject *Project, rule *gitolite.Rule, gtl *gitolite.Gitolite) *Project {
	if currentProject != nil && !currentProject.hasSameUsers(rule.GetUsersFirstOrGroups()) {
		fmt.Printf("\nIgnore project name '%v': admins differ on '-' (%v vs. %v)\n", currentProject.name,
			currentProject.admins, rule.GetUsersFirstOrGroups())
		currentProject = nil
	}
	if currentProject != nil {
		group := gtl.GetGroup("@" + currentProject.name)
		if group == nil {
			fmt.Printf("\nIgnore project name '%v': no repo group found\n", currentProject.name)
			currentProject = nil
		} else if group.IsUsers() {
			fmt.Printf("\nIgnore project name '%v': user group found (instead of repo group)\n", currentProject.name)
			currentProject = nil
		} else if group.IsUndefined() {
			group.MarkAsRepoGroup()
		}
		if currentProject != nil {
			pm.projects = append(pm.projects, currentProject)
		}
	}
	currentProject = nil
	return currentProject
}

func (p *Project) hasSameUsers(users []gitolite.UserOrGroup) bool {
	//fmt.Printf("\nusers '%v'\n", users)
	if len(p.admins) != len(users) {
		return false
	}
	for _, pusers := range p.admins {
		seen := false
		for _, user := range users {
			if pusers.GetName() == user.GetName() {
				seen = true
				break
			}
		}
		if !seen {
			return false
		}
	}
	return true
}
