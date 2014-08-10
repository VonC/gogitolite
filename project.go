package gogitolite

import (
	"fmt"
	"strings"
)

type Project struct {
	name  string
	users []*User
}

func (gtl *Gitolite) NbProjects() int {
	gtl.updateProjects()
	return len(gtl.projects)
}

var prefix = "VREF/NAME/conf/subs/"

func (gtl *Gitolite) updateProjects() {
	configs := gtl.reposToConfigs["gitolite-admin"]
	for _, config := range configs {
		var currentProject *Project = nil
		rules := config.rules
		for _, rule := range rules {
			//fmt.Printf("\nRule looked at: '%v' => '%v' '%v'\n", rule, rule.access, rule.param)
			if rule.access == "RW" && rule.param == "" {
				currentProject = &Project{users: rule.users}
			} else if rule.access == "RW" && strings.HasPrefix(rule.param, prefix) {
				projectname := rule.param[len(prefix):]
				if currentProject == nil {
					fmt.Printf("\nIgnore project name '%v': no RW rule before\n", projectname)
				} else {
					currentProject.name = projectname
				}
			} else if rule.access == "-" && rule.param == "VREF/NAME/" {
				if currentProject != nil && currentProject.name != "" {
					gtl.projects = append(gtl.projects, currentProject)
				}
				if currentProject != nil && currentProject.name == "" {
					fmt.Printf("\nIgnore project with no name\n")
				}
				currentProject = nil
			} else {
				currentProject = nil
			}
		}
	}
}
