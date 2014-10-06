package project

import (
	"bufio"
	"bytes"
	"fmt"
	"strings"
	"testing"

	"github.com/VonC/gogitolite/gitolite"
	"github.com/VonC/gogitolite/reader"
	. "github.com/smartystreets/goconvey/convey"
)

var (
	bout *bytes.Buffer
	wout *bufio.Writer
	berr *bytes.Buffer
	werr *bufio.Writer
)

func init() {

	out()
	oerr()
	bout = bytes.NewBuffer(nil)
	wout = bufio.NewWriter(bout)
	sout = wout
	berr = bytes.NewBuffer(nil)
	werr = bufio.NewWriter(berr)
	serr = werr
	out()
}

func resetStds() {
	bout = bytes.NewBuffer(nil)
	sout.Reset(bout)
	berr = bytes.NewBuffer(nil)
	serr.Reset(berr)
}

func flushStds() {
	sout.Flush()
	serr.Flush()
}

/*
   subconf "subs/*.conf"

   more subs/project.conf

   repo @project
     RW = projectowner @almadmins
     RW = otheruser1 otheruser2...
*/
func TestProject(t *testing.T) {

	Convey("Detects projects", t, func() {

		Convey("Detects one project", func() {
			var gitoliteconf = `
		@project = module1 module2

		repo gitolite-admin
	      RW+     =   gitoliteadm @almadmins
	      RW                                = projectowner
	      RW VREF/NAME/conf/subs/project    = projectowner
	      -  VREF/NAME/                     = projectowner

	    repo module1
	      RW+ = projectowner @almadmins
`
			r := strings.NewReader(gitoliteconf)
			gtl, err := reader.Read(r)
			subconfs := make(map[string]*gitolite.Gitolite)
			subconfs["path/project.conf"] = gtl
			pm := NewManager(gtl, subconfs)
			flushStds()
			resetStds()
			So(err, ShouldBeNil)
			So(gtl.IsEmpty(), ShouldBeFalse)
			So(gtl.NbRepos(), ShouldEqual, 3)
			So(pm.NbProjects(), ShouldEqual, 1)
			So(fmt.Sprintf("groups '%v'", gtl.GetGroup("@project")), ShouldEqual, "groups 'group '@project'<repos>: [module1 module2]'")
		})

		Convey("Detects one project even if group undefined", func() {
			var gitoliteconf = `
		@project = module1 module2

		repo gitolite-admin
	      RW+     =   gitoliteadm @almadmins
	      RW                                = projectowner
	      RW VREF/NAME/conf/subs/project    = projectowner
	      -  VREF/NAME/                     = projectowner
`
			r := strings.NewReader(gitoliteconf)
			gtl, err := reader.Read(r)
			subconfs := make(map[string]*gitolite.Gitolite)
			subconfs["path/project.conf"] = gtl
			pm := NewManager(gtl, subconfs)
			So(err, ShouldBeNil)
			So(gtl, ShouldNotBeNil)
			So(gtl.IsEmpty(), ShouldBeFalse)
			So(gtl.NbRepos(), ShouldEqual, 3)
			So(pm.NbProjects(), ShouldEqual, 1)
			//fmt.Println("\nPRJ: ", pm)
			So(gtl.NbRepos(), ShouldEqual, 3)
			So(fmt.Sprintf("groups '%v'", gtl.GetGroup("@project")), ShouldEqual, "groups 'group '@project'<repos>: [module1 module2]'")
			So(pm.Projects()[0].String(), ShouldEqual, "project project, admins: projectowner, members: ")
			flushStds()
			resetStds()
		})

		Convey("Detects one project with several admins and users", func() {
			var gitoliteconf = `
		@project = module1 module2

		repo gitolite-admin
	      RW+     =   gitoliteadm @almadmins
	      RW                                = projectowner1 projectowner2
	      RW VREF/NAME/conf/subs/project    = projectowner1 projectowner2
	      -  VREF/NAME/                     = projectowner1 projectowner2

	    repo module1
	      RW = user1 user11 user2
	    repo module2
	      RW = user2 user21
`
			r := strings.NewReader(gitoliteconf)
			gtl, err := reader.Read(r)
			subconfs := make(map[string]*gitolite.Gitolite)
			subconfs["path/project.conf"] = gtl
			pm := NewManager(gtl, subconfs)
			So(err, ShouldBeNil)
			So(gtl, ShouldNotBeNil)
			So(gtl.IsEmpty(), ShouldBeFalse)
			So(gtl.NbRepos(), ShouldEqual, 3)
			So(pm.NbProjects(), ShouldEqual, 1)
			//fmt.Println("\nPRJ: ", pm)
			So(gtl.NbRepos(), ShouldEqual, 3)
			So(fmt.Sprintf("groups '%v'", gtl.GetGroup("@project")), ShouldEqual, "groups 'group '@project'<repos>: [module1 module2]'")
			So(pm.Projects()[0].String(), ShouldEqual, "project project, admins: projectowner1, projectowner2, members: user1, user11, user2, user21")
			flushStds()
			resetStds()
		})

		Convey("No project if no RW rule before", func() {
			var gitoliteconf = `
		@project = module1 module2

		repo gitolite-admin
	      RW+     =   gitoliteadm @almadmins
	      RW VREF/NAME/conf/subs/project    = projectowner
	      RW                                = projectowner
	      -  VREF/NAME/                     = projectowner

	    repo module1
	      RW+ = projectowner @almadmins
`
			r := strings.NewReader(gitoliteconf)
			gtl, err := reader.Read(r)
			pm := &Manager{gtl: gtl}
			So(err, ShouldBeNil)
			So(gtl.IsEmpty(), ShouldBeFalse)
			So(gtl.NbRepos(), ShouldEqual, 3)
			So(pm.NbProjects(), ShouldEqual, 0)
			flushStds()
			resetStds()
		})

		Convey("No project if none detected", func() {
			var gitoliteconf = `
		@project = module1 module2

		repo gitolite-admin
	      RW+     =   gitoliteadm @almadmins

	    repo module1
	      RW+ = projectowner @almadmins
`
			r := strings.NewReader(gitoliteconf)
			gtl, err := reader.Read(r)
			pm := &Manager{gtl: gtl}
			So(err, ShouldBeNil)
			So(gtl.IsEmpty(), ShouldBeFalse)
			So(gtl.NbRepos(), ShouldEqual, 3)
			So(pm.NbProjects(), ShouldEqual, 0)
			flushStds()
			resetStds()
		})

		Convey("No project if users changes in VREF/NAME/conf", func() {
			var gitoliteconf = `
		@project = module1 module2

		repo gitolite-admin
	      RW+     =   gitoliteadm @almadmins
	      RW                                = projectowner
	      RW VREF/NAME/conf/subs/project    = projectowner otheruser
	      -  VREF/NAME/                     = projectowner

	    repo module1
	      RW+ = projectowner @almadmins
`
			r := strings.NewReader(gitoliteconf)
			gtl, err := reader.Read(r)
			pm := &Manager{gtl: gtl}
			So(err, ShouldBeNil)
			So(gtl.IsEmpty(), ShouldBeFalse)
			So(gtl.NbRepos(), ShouldEqual, 3)
			So(gtl.NbUsers(), ShouldEqual, 3)
			So(pm.NbProjects(), ShouldEqual, 0)
			flushStds()
			resetStds()
		})

		Convey("No project if users changes in VREF/NAME/", func() {
			var gitoliteconf = `
		@project = module1 module2

		repo gitolite-admin
	      RW+     =   gitoliteadm @almadmins
	      RW                                = projectowner
	      RW VREF/NAME/conf/subs/project    = projectowner
	      -  VREF/NAME/                     = otheruser1

	    repo module1
	      RW+ = projectowner @almadmins
`
			r := strings.NewReader(gitoliteconf)
			gtl, err := reader.Read(r)
			pm := &Manager{gtl: gtl}
			So(err, ShouldBeNil)
			So(gtl.IsEmpty(), ShouldBeFalse)
			So(gtl.NbRepos(), ShouldEqual, 3)
			So(gtl.NbUsers(), ShouldEqual, 3)
			So(pm.NbProjects(), ShouldEqual, 0)
			flushStds()
			resetStds()
		})

		Convey("No project if no repo group", func() {
			var gitoliteconf = `
		repo gitolite-admin
	      RW+     =   gitoliteadm @almadmins
	      RW                                = projectowner
	      RW VREF/NAME/conf/subs/project    = projectowner
	      -  VREF/NAME/                     = projectowner

	    repo module1
	      RW+ = projectowner @almadmins
`
			r := strings.NewReader(gitoliteconf)
			gtl, err := reader.Read(r)
			pm := &Manager{gtl: gtl}
			So(err, ShouldBeNil)
			So(gtl.IsEmpty(), ShouldBeFalse)
			So(gtl.NbRepos(), ShouldEqual, 2)
			So(pm.NbProjects(), ShouldEqual, 0)
			flushStds()
			resetStds()
		})

		Convey("No project if user group", func() {
			var gitoliteconf = `
	    @project = user1 user2

	    repo arepo
	      RW+ = @project

		repo gitolite-admin
	      RW+     =   gitoliteadm @almadmins
	      RW                                = projectowner
	      RW VREF/NAME/conf/subs/project    = projectowner
	      -  VREF/NAME/                     = projectowner

	    repo module1
	      RW+ = projectowner @almadmins
`
			r := strings.NewReader(gitoliteconf)
			gtl, err := reader.Read(r)
			pm := &Manager{gtl: gtl}
			So(err, ShouldNotBeNil)
			So(gtl, ShouldNotBeNil)
			So(strings.Contains(err.Error(), "group '@project' is a users group, not a repo one"), ShouldBeTrue)
			So(gtl.IsEmpty(), ShouldBeFalse)
			So(gtl.NbRepos(), ShouldEqual, 2)
			So(pm.NbProjects(), ShouldEqual, 0)
			flushStds()
			resetStds()
		})

		Convey("No project if empty name", func() {
			var gitoliteconf = `
		@project = module1 module2

		repo gitolite-admin
	      RW+     =   gitoliteadm @almadmins
	      RW                                = projectowner
	      RW VREF/NAME/conf/subs/           = projectowner
	      -  VREF/NAME/                     = projectowner

	    repo module1
	      RW+ = projectowner @almadmins
`
			r := strings.NewReader(gitoliteconf)
			gtl, err := reader.Read(r)
			subconfs := make(map[string]*gitolite.Gitolite)
			subconfs["path/project.conf"] = gtl
			pm := NewManager(gtl, subconfs)
			So(err, ShouldBeNil)
			So(gtl, ShouldNotBeNil)
			//So(strings.Contains(err.Error(), "group '@project' is a users group, not a repo one"), ShouldBeTrue)
			So(gtl.IsEmpty(), ShouldBeFalse)
			So(gtl.NbRepos(), ShouldEqual, 3)
			So(pm.NbProjects(), ShouldEqual, 0)
			flushStds()
			So(berr.String(), ShouldEqual, `Ignore project with no name
`)
			resetStds()
		})

		Convey("No project if no subconfs", func() {
			var gitoliteconf = `
		@project = module1 module2

		repo gitolite-admin
	      RW+     =   gitoliteadm @almadmins
	      RW                                = projectowner
	      RW VREF/NAME/conf/subs/project    = projectowner
	      -  VREF/NAME/                     = projectowner

	    repo module1
	      RW+ = projectowner @almadmins
`
			r := strings.NewReader(gitoliteconf)
			gtl, err := reader.Read(r)
			pm := NewManager(gtl, nil)
			So(err, ShouldBeNil)
			So(gtl, ShouldNotBeNil)
			//So(strings.Contains(err.Error(), "group '@project' is a users group, not a repo one"), ShouldBeTrue)
			So(gtl.IsEmpty(), ShouldBeFalse)
			So(gtl.NbRepos(), ShouldEqual, 3)
			So(pm.NbProjects(), ShouldEqual, 0)
			flushStds()
			So(berr.String(), ShouldEqual, `Ignore project name 'project': no subconf found
`)
			resetStds()
		})

		Convey("No project if no RW rule first", func() {
			var gitoliteconf = `
		@project = module1 module2

		repo gitolite-admin
	      RW+     =   gitoliteadm @almadmins
	      RW VREF/NAME/conf/subs/project    = projectowner
	      -  VREF/NAME/                     = projectowner

	    repo module1
	      RW+ = projectowner @almadmins
`
			r := strings.NewReader(gitoliteconf)
			gtl, err := reader.Read(r)
			subconfs := make(map[string]*gitolite.Gitolite)
			subconfs["path/project.conf"] = gtl
			pm := NewManager(gtl, subconfs)
			So(err, ShouldBeNil)
			So(gtl, ShouldNotBeNil)
			//So(strings.Contains(err.Error(), "group '@project' is a users group, not a repo one"), ShouldBeTrue)
			So(gtl.IsEmpty(), ShouldBeFalse)
			So(gtl.NbRepos(), ShouldEqual, 3)
			So(pm.NbProjects(), ShouldEqual, 0)
			flushStds()
			So(berr.String(), ShouldEqual, `Ignore project name 'project': no RW rule before.
`)
			resetStds()
		})

		Convey("No project if admins change", func() {
			var gitoliteconf = `
		@project = module1 module2

		repo gitolite-admin
	      RW+     =   gitoliteadm @almadmins
	      RW                                = projectowner1
	      RW VREF/NAME/conf/subs/project    = projectowner1 po2
	      -  VREF/NAME/                     = projectowner

	    repo module1
	      RW+ = projectowner @almadmins
`
			r := strings.NewReader(gitoliteconf)
			gtl, err := reader.Read(r)
			subconfs := make(map[string]*gitolite.Gitolite)
			subconfs["path/project.conf"] = gtl
			pm := NewManager(gtl, subconfs)
			So(err, ShouldBeNil)
			So(gtl, ShouldNotBeNil)
			//So(strings.Contains(err.Error(), "group '@project' is a users group, not a repo one"), ShouldBeTrue)
			So(gtl.IsEmpty(), ShouldBeFalse)
			So(gtl.NbRepos(), ShouldEqual, 3)
			So(pm.NbProjects(), ShouldEqual, 0)
			flushStds()
			So(berr.String(), ShouldEqual, `Ignore project name 'project': Admins differ on 'RW' ([user 'projectowner1'] vs. [user 'projectowner1' user 'po2'])
`)
			resetStds()
		})

		Convey("No project if admins change on last rule", func() {
			var gitoliteconf = `
		@project = module1 module2

		repo gitolite-admin
	      RW+     =   gitoliteadm @almadmins
	      RW                                = projectowner
	      RW VREF/NAME/conf/subs/project    = projectowner
	      -  VREF/NAME/                     = projectowner1

	    repo module1
	      RW+ = projectowner @almadmins
`
			r := strings.NewReader(gitoliteconf)
			gtl, err := reader.Read(r)
			subconfs := make(map[string]*gitolite.Gitolite)
			subconfs["path/project.conf"] = gtl
			pm := NewManager(gtl, subconfs)
			So(err, ShouldBeNil)
			So(gtl, ShouldNotBeNil)
			//So(strings.Contains(err.Error(), "group '@project' is a users group, not a repo one"), ShouldBeTrue)
			So(gtl.IsEmpty(), ShouldBeFalse)
			So(gtl.NbRepos(), ShouldEqual, 3)
			So(pm.NbProjects(), ShouldEqual, 0)
			flushStds()
			So(berr.String(), ShouldEqual, `Ignore project name 'project': admins differ on '-' ([user 'projectowner'] vs. [user 'projectowner1'])
`)
			resetStds()
		})

	})

	Convey("Update projects", t, func() {

		Convey("Update a non-existing project means adding/creating one", func() {
			var gitoliteconf = `
		@project2 = module1 module2

		repo gitolite-admin
	      RW+     =   gitoliteadm @almadmins
	      RW                                = projectowner
	      RW VREF/NAME/conf/subs/project2   = projectowner
	      -  VREF/NAME/                     = projectowner

	    repo module1
	      RW+ = projectowner @almadmins
`
			r := strings.NewReader(gitoliteconf)
			gtl, err := reader.Read(r)
			pm := NewManager(gtl, nil)
			So(err, ShouldBeNil)
			So(gtl, ShouldNotBeNil)
			//So(strings.Contains(err.Error(), "group '@project' is a users group, not a repo one"), ShouldBeTrue)
			So(gtl.IsEmpty(), ShouldBeFalse)
			So(gtl.NbRepos(), ShouldEqual, 3)
			So(pm.NbProjects(), ShouldEqual, 0)
			flushStds()
			So(berr.String(), ShouldEqual, `Ignore project name 'project2': no subconf found
`)
			resetStds()
		})
	})

	Convey("Add a project", t, func() {
		var gitoliteconf = `
		@project = module1 module2

		repo gitolite-admin
	      RW+     =   gitoliteadm @almadmins
	      RW                                = projectowner
	      RW VREF/NAME/conf/subs/project    = projectowner
	      -  VREF/NAME/                     = projectowner

	    repo module1
	      RW+ = projectowner @almadmins
`
		r := strings.NewReader(gitoliteconf)
		gtl, err := reader.Read(r)
		subconfs := make(map[string]*gitolite.Gitolite)
		subconfs["path/project.conf"] = gtl
		pm := NewManager(gtl, subconfs)
		flushStds()
		resetStds()

		Convey("Adding an existing project errors", func() {

			So(err, ShouldBeNil)
			err = pm.AddProject("project")
			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldEqual, "project 'project' already exits")
		})
		Convey("Adding a new project works", func() {

			So(err, ShouldBeNil)
			err = pm.AddProject("project2")
			So(err, ShouldBeNil)
			So(pm.NbProjects(), ShouldEqual, 2)
		})
	})
}
