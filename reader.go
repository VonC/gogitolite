package gogitolite

import (
	"bufio"
	"fmt"
	"io"
	"regexp"
	"strings"
)

// Gitolite config decoded
type Gitolite struct {
	groups []*Group
}

type content struct {
	s   *bufio.Scanner
	l   int
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
	s := bufio.NewScanner(r)
	s.Scan()
	c := &content{s: s, gtl: res}
	var state stateFn
	var err error
	for state, err = readEmptyOrCommentLines(c); state != nil && err == nil; {
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

var readEmptyOrCommentLinesRx = regexp.MustCompile(`(?m)^\s*?$|^\s*?#(.*?)$`)

func readEmptyOrCommentLines(c *content) (stateFn, error) {
	for keepReading := true; keepReading; {
		t := c.s.Text()
		res := readEmptyOrCommentLinesRx.FindStringSubmatchIndex(t)
		if res == nil {
			return readRepoOrGroup, nil
		}
		if !c.s.Scan() {
			return nil, ParseError{msg: fmt.Sprintf("group or repo expected after line %v ('%v')", c.l, t)}
		}
		c.l = c.l + 1
	}
	return nil, nil
}

var readRepoOrGroupRx = regexp.MustCompile(`^\s*?(repo |@)`)

func readRepoOrGroup(c *content) (stateFn, error) {
	t := c.s.Text()
	res := readRepoOrGroupRx.FindStringSubmatchIndex(t)
	if res == nil {
		return nil, ParseError{msg: fmt.Sprintf("group or repo expect after line %v ('%v')", c.l, t)}
	}
	prefix := t[res[2]:res[3]]
	if prefix == "@" {
		return readGroup, nil
	}
	return nil, nil
}

var readGroupRx = regexp.MustCompile(`(?m)^@([a-zA-Z0-9_-]+)\s*?=\s*?((?:[a-zA-Z0-9_-]+\s*?)+)$`)

func readGroup(c *content) (stateFn, error) {
	t := c.s.Text()
	res := readGroupRx.FindStringSubmatchIndex(t)
	// fmt.Println(res, "'"+c.s+"'")
	if len(res) == 0 {
		return nil, ParseError{msg: fmt.Sprintf("Incorrect repo declaration at line %v ('%v')", c.l, t)}
	}
	//fmt.Println(res, "'"+c.s+"'", "'"+c.s[res[2]:res[3]]+"'", "'"+c.s[res[4]:res[5]]+"'")
	grpname := t[res[2]:res[3]]
	grpmembers := strings.Split(strings.TrimSpace(t[res[4]:res[5]]), " ")
	grp := &Group{name: grpname, members: grpmembers}
	c.gtl.groups = append(c.gtl.groups, grp)
	// fmt.Println("'" + c.s + "'")
	return readEmptyOrCommentLines, nil
}

// NbGroup returns the number of groups (people or repos)
func (gtl *Gitolite) NbGroup() int {
	return len(gtl.groups)
}
