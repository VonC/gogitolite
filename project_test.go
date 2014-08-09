package gogitolite

import (
	"strings"
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

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

/*
   subconf "subs/*.conf"

   more subs/project.conf

   repo @project
     RW = projectowner @almadmins
     RW = otheruser1 otheruser2...
*/
func TestProject(t *testing.T) {

	Convey("Detects projects", t, func() {
		r := strings.NewReader(gitoliteconf)
		gtl, err := Read(r)
		So(err, ShouldBeNil)
		So(gtl.IsEmpty(), ShouldBeFalse)
		So(gtl.NbRepos(), ShouldEqual, 3)
		So(gtl.NbProjects(), ShouldEqual, 1)
	})
}
