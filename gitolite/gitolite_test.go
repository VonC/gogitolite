package gitolite

import (
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

	Convey("Test Gitolite", t, func() {

		Convey("Empty gitolite", func() {
			gtl := NewGitolite()
			So(gtl.IsEmpty(), ShouldBeTrue)
		})
	})
	Convey("Test Gitolite", t, func() {

		Convey("Undefined group", func() {
			grp := &Group{}
			So(grp.IsUndefined(), ShouldBeTrue)
		})

		Convey("Users or Repos Group", func() {
			gtl := NewGitolite()
			grp := &Group{}
			grp.container = gtl
			grp.markAsUserGroup()
			So(grp.IsUndefined(), ShouldBeFalse)
			So(grp.IsUsers(), ShouldBeTrue)
			grp = &Group{}
			grp.container = gtl
			grp.MarkAsRepoGroup()
			So(grp.IsUndefined(), ShouldBeFalse)
			So(grp.IsUsers(), ShouldBeFalse)
		})

	})
}
