package gogitolite

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

// ProjectManager manages project for a gitolite instance
type ProjectManager struct {
	gtl      *gitolite.Gitolite
	projects []*Project
}

// NbProjects returns the number of detected projects
func (pm *ProjectManager) NbProjects() int {
	pm.updateProjects()
	return len(pm.projects)
}

var prefix = "VREF/NAME/conf/subs/"

func (pm *ProjectManager) updateProjects() {
	gtl := pm.gtl
	configs := gtl.GetConfigsForRepo("gitolite-admin")
	for _, config := range configs {
		var currentProject *Project
		rules := config.Rules()
		for _, rule := range rules {
			//fmt.Printf("\nRule looked at: '%v' => '%v' '%v'\n", rule, rule.Access(), rule.Param())
			if rule.Access() == "RW" && rule.Param() == "" {
				currentProject = &Project{users: rule.GetUsers()}
			} else if rule.Access() == "RW" && strings.HasPrefix(rule.Param(), prefix) {
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
			} else if rule.Access() == "-" && rule.Param() == "VREF/NAME/" {
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
			} else {
				currentProject = nil
			}
		}
	}
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
