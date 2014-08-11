package gogitolite

import (
	"strings"
	"testing"

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
			gtl, err := Read(r)
			So(err, ShouldBeNil)
			So(gtl.IsEmpty(), ShouldBeFalse)
			So(gtl.NbRepos(), ShouldEqual, 3)
			So(gtl.NbProjects(), ShouldEqual, 1)
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
			gtl, err := Read(r)
			So(err, ShouldBeNil)
			So(gtl.IsEmpty(), ShouldBeFalse)
			So(gtl.NbRepos(), ShouldEqual, 3)
			So(gtl.NbProjects(), ShouldEqual, 0)
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
			gtl, err := Read(r)
			So(err, ShouldBeNil)
			So(gtl.IsEmpty(), ShouldBeFalse)
			So(gtl.NbRepos(), ShouldEqual, 3)
			So(gtl.NbProjects(), ShouldEqual, 0)
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
			gtl, err := Read(r)
			So(err, ShouldBeNil)
			So(gtl.IsEmpty(), ShouldBeFalse)
			So(gtl.NbRepos(), ShouldEqual, 3)
			So(gtl.NbUsers(), ShouldEqual, 3)
			So(gtl.NbProjects(), ShouldEqual, 0)
		})

		Convey("No project if users changes in VREF/NAME/", func() {
			var gitoliteconf = `
		@project = module1 module2

		repo gitolite-admin
	      RW+     =   gitoliteadm @almadmins
	      RW                                = projectowner
	      RW VREF/NAME/conf/subs/project    = projectowner
	      -  VREF/NAME/                     = projectowner otheruser

	    repo module1
	      RW+ = projectowner @almadmins
`
			r := strings.NewReader(gitoliteconf)
			gtl, err := Read(r)
			So(err, ShouldBeNil)
			So(gtl.IsEmpty(), ShouldBeFalse)
			So(gtl.NbRepos(), ShouldEqual, 3)
			So(gtl.NbUsers(), ShouldEqual, 3)
			So(gtl.NbProjects(), ShouldEqual, 0)
		})

	})
}
