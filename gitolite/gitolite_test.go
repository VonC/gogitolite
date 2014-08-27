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
			grp := &Group{name: "grp1"}
			grp.container = gtl
			var err error
			err = grp.markAsUserGroup()
			So(err, ShouldBeNil)
			So(grp.IsUndefined(), ShouldBeFalse)
			So(grp.IsUsers(), ShouldBeTrue)
			err = grp.MarkAsRepoGroup()
			So(err.Error(), ShouldEqual, "group 'grp1' is a users group, not a repo one")

			grp = &Group{name: "grp2"}
			grp.container = gtl
			err = grp.MarkAsRepoGroup()
			So(err, ShouldBeNil)
			So(grp.IsUndefined(), ShouldBeFalse)
			So(grp.IsUsers(), ShouldBeFalse)
			err = grp.markAsUserGroup()
			So(err.Error(), ShouldEqual, "group 'grp2' is a repos group, not a user one")
		})
		Convey("Users can be added", func() {
			gtl := NewGitolite()
			grp := &Group{}
			grp.members = append(grp.members, "user1")
			grp.container = gtl
			var err error
			err = grp.markAsUserGroup()
			So(err, ShouldBeNil)
			So(gtl.NbUsers(), ShouldEqual, 1)
		})

	})
}
