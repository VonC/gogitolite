package gogitolite

import "io"

// Gitolite config decoded
type Gitolite struct{}

// Read a gitolite config file
func Read(r io.Reader) *Gitolite {
	res := &Gitolite{}
	if r == nil {
		return res
	}
	return res
}

func (gtl *Gitolite) IsEmpty() bool {
	return true
}
