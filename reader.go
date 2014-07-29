package gogitolite

import (
	"io"
	"io/ioutil"
	"regexp"
	"strings"
)

// Gitolite config decoded
type Gitolite struct {
	groups []*Group
}

type content struct {
	s   string
	gtl *Gitolite
}

type stateFn func(*content) stateFn

// Group (of repo or resources, ie people)
type Group struct {
	name    string
	members []string
}

// Read a gitolite config file
func Read(r io.Reader) *Gitolite {
	res := &Gitolite{}
	if r == nil {
		return res
	}
	s, _ := ioutil.ReadAll(r)
	c := &content{s: string(s), gtl: res}
	for state := readUpToRepoOrGroup(c); state != nil; {
		state = state(c)
	}
	return res
}

// IsEmpty checks if config includes any repo or groups
func (gtl *Gitolite) IsEmpty() bool {
	return gtl.groups == nil || len(gtl.groups) == 0
}

var readUpToRepoOrGroupRx = regexp.MustCompile(`(^\s*?$|^\s*?#.*?$)*?^\s*?(repo |@)`)

func readUpToRepoOrGroup(c *content) stateFn {
	res := readUpToRepoOrGroupRx.FindStringSubmatchIndex(c.s)
	if res == nil {
		return nil
	}
	// c.i = res[4]
	prefix := c.s[res[4]:res[5]]
	c.s = c.s[res[4]:]
	if prefix == "@" {
		return readGroup
	}
	return nil
}

var readGroupRx = regexp.MustCompile(`^@([a-zA-Z0-9_-]+)\s*?=\s*?((?:[a-zA-Z0-9_-]+\s*?)+)$`)

func readGroup(c *content) stateFn {
	res := readGroupRx.FindStringSubmatchIndex(c.s)
	//fmt.Println(res, "'"+c.s+"'", "'"+c.s[res[2]:res[3]]+"'", "'"+c.s[res[4]:res[5]]+"'")
	grpname := c.s[res[2]:res[3]]
	grpmembers := strings.Split(strings.TrimSpace(c.s[res[4]:res[5]]), " ")
	grp := &Group{name: grpname, members: grpmembers}
	c.gtl.groups = append(c.gtl.groups, grp)
	c.s = c.s[res[5]:]
	return readUpToRepoOrGroup
}

// NbGroup returns the number of groups (people or repos)
func (gtl *Gitolite) NbGroup() int {
	return len(gtl.groups)
}
