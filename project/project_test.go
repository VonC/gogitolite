package project

import (
	"fmt"
	"strings"
	"testing"

	"github.com/VonC/gogitolite/gitolite"
	"github.com/VonC/gogitolite/reader"
	. "github.com/smartystreets/goconvey/convey"
)

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
			So(err, ShouldBeNil)
			So(gtl.IsEmpty(), ShouldBeFalse)
			So(gtl.NbRepos(), ShouldEqual, 3)
			So(pm.NbProjects(), ShouldEqual, 0)
		})

	})
}
