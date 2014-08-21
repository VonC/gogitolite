package gogitolite

import (
	"fmt"
	"strings"
)

// Project has a name and users
type Project struct {
	name  string
	users []*User
}

// NbProjects returns the number of detected projects
func (gtl *Gitolite) NbProjects() int {
	gtl.updateProjects()
	return len(gtl.projects)
}

var prefix = "VREF/NAME/conf/subs/"

func (gtl *Gitolite) updateProjects() {
	configs := gtl.reposToConfigs["gitolite-admin"]
	for _, config := range configs {
		var currentProject *Project
		rules := config.rules
		for _, rule := range rules {
			//fmt.Printf("\nRule looked at: '%v' => '%v' '%v'\n", rule, rule.access, rule.param)
			if rule.access == "RW" && rule.param == "" {
				currentProject = &Project{users: rule.GetUsers()}
			} else if rule.access == "RW" && strings.HasPrefix(rule.param, prefix) {
				projectname := rule.param[len(prefix):]
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
			} else if rule.access == "-" && rule.param == "VREF/NAME/" {
				if currentProject != nil && currentProject.name == "" {
					fmt.Printf("\nIgnore project with no name\n")
					currentProject = nil
				}
				if currentProject != nil && !currentProject.hasSameUsers(rule.GetUsers()) {
					fmt.Printf("\nIgnore project name '%v': users differ on '-' (%v vs. %v)\n", currentProject.name,
						currentProject.users, rule.GetUsers())
					currentProject = nil
				}
				if currentProject != nil {
					group := gtl.getGroup("@" + currentProject.name)
					if group == nil {
						fmt.Printf("\nIgnore project name '%v': no repo group found\n", currentProject.name)
						currentProject = nil
					} else if group.kind == users {
						fmt.Printf("\nIgnore project name '%v': user group found (instead of repo group)\n", currentProject.name)
						currentProject = nil
					} else if group.kind == undefined {
						group.markAsRepoGroup()
					}
					if currentProject != nil {
						gtl.projects = append(gtl.projects, currentProject)
					}
				}
				currentProject = nil
			} else {
				currentProject = nil
			}
		}
	}
}

func (p *Project) hasSameUsers(users []*User) bool {
	//fmt.Printf("\nusers '%v'\n", users)
	if len(p.users) != len(users) {
		return false
	}
	for _, pusers := range p.users {
		seen := false
		for _, user := range users {
			if pusers.name == user.name {
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
