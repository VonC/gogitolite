package gogitolite

import (
	"io"
	"io/ioutil"
)

// Gitolite config decoded
type Gitolite struct {
	groups []*Group
}

// Read a gitolite config file
func Read(r io.Reader) *Gitolite {
	res := &Gitolite{}
	if r == nil {
		return res
	}
	s, _ := ioutil.ReadAll(r)
	if len(s) > 0 {
		grp := &Group{}
		res.groups = append(res.groups, grp)
	}
	return res
}

// Group (of repo or resources, ie people)
type Group struct{}

// IsEmpty checks if config includes any repo or groups
func (gtl *Gitolite) IsEmpty() bool {
	return gtl.groups == nil || len(gtl.groups) == 0
}
