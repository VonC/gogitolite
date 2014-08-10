package gogitolite

type Project struct{}

func (gtl *Gitolite) NbProjects() int {
	if gtl.projects != nil {
		return len(gtl.projects)
	}
	gtl.updateProjects()
	return len(gtl.projects)
}

func (gtl *Gitolite) updateProjects() {

}
