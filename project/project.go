package project

import (
	"fmt"
	"strings"

	"github.com/VonC/gogitolite/gitolite"
)

// Project has a name and users
type Project struct {
	name  string
	users []*gitolite.User
}

// Manager manages project for a gitolite instance
type Manager struct {
	gtl      *gitolite.Gitolite
	projects []*Project
}

// NbProjects returns the number of detected projects
func (pm *Manager) NbProjects() int {
	pm.updateProjects()
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
				currentProject = &Project{users: rule.GetUsers()}
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
		if currentProject != nil && !currentProject.hasSameUsers(rule.GetUsers()) {
			fmt.Printf("\nIgnore project name '%v': users differ on 'RW' (%v vs. %v)\n", projectname,
				currentProject.users, rule.GetUsers())
			currentProject = nil
		}
	}
	return isrw, currentProject
}

func (pm *Manager) currentProjectVREFName(currentProject *Project, rule *gitolite.Rule, gtl *gitolite.Gitolite) *Project {
	if currentProject != nil && !currentProject.hasSameUsers(rule.GetUsers()) {
		fmt.Printf("\nIgnore project name '%v': users differ on '-' (%v vs. %v)\n", currentProject.name,
			currentProject.users, rule.GetUsers())
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

func (p *Project) hasSameUsers(users []*gitolite.User) bool {
	//fmt.Printf("\nusers '%v'\n", users)
	if len(p.users) != len(users) {
		return false
	}
	for _, pusers := range p.users {
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
