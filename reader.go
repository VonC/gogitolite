package gogitolite

import (
	"fmt"
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

type stateFn func(*content) (stateFn, error)

// Group (of repo or resources, ie people)
type Group struct {
	name    string
	members []string
}

// Read a gitolite config file
func Read(r io.Reader) (*Gitolite, error) {
	res := &Gitolite{}
	if r == nil {
		return res, nil
	}
	s, _ := ioutil.ReadAll(r)
	c := &content{s: string(s), gtl: res}
	var state stateFn
	var err error
	for state, err = readUpToRepoOrGroup(c); state != nil && err == nil; {
		state, err = state(c)
	}
	return res, err
}

// IsEmpty checks if config includes any repo or groups
func (gtl *Gitolite) IsEmpty() bool {
	return gtl.groups == nil || len(gtl.groups) == 0
}

// ParseError indicates gitolite.conf parsing error
type ParseError struct {
	msg string
}

func (pe ParseError) Error() string {
	return fmt.Sprintf("Parse Error: %s", pe.msg)
}

var readUpToRepoOrGroupRx = regexp.MustCompile(`(^\s*?$|^\s*?#.*?$)*?^\s*?(repo |@)`)

func readUpToRepoOrGroup(c *content) (stateFn, error) {
	res := readUpToRepoOrGroupRx.FindStringSubmatchIndex(c.s)
	if res == nil {
		return nil, nil
	}
	// c.i = res[4]
	prefix := c.s[res[4]:res[5]]
	c.s = c.s[res[4]:]
	if prefix == "@" {
		return readGroup, nil
	}
	return nil, nil
}

var readGroupRx = regexp.MustCompile(`(?m)^@([a-zA-Z0-9_-]+)\s*?=\s*?((?:[a-zA-Z0-9_-]+\s*?)+)$`)

func readGroup(c *content) (stateFn, error) {
	res := readGroupRx.FindStringSubmatchIndex(c.s)
	// fmt.Println(res, "'"+c.s+"'")
	if len(res) == 0 {
		return nil, ParseError{msg: "bad repo line"}
	}
	//fmt.Println(res, "'"+c.s+"'", "'"+c.s[res[2]:res[3]]+"'", "'"+c.s[res[4]:res[5]]+"'")
	grpname := c.s[res[2]:res[3]]
	grpmembers := strings.Split(strings.TrimSpace(c.s[res[4]:res[5]]), " ")
	grp := &Group{name: grpname, members: grpmembers}
	c.gtl.groups = append(c.gtl.groups, grp)
	c.s = c.s[res[5]:]
	// fmt.Println("'" + c.s + "'")
	return readUpToRepoOrGroup, nil
}

// NbGroup returns the number of groups (people or repos)
func (gtl *Gitolite) NbGroup() int {
	return len(gtl.groups)
}
