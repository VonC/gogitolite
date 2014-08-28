package gitolite

import (
	"fmt"
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
			So(grp.GetName(), ShouldEqual, "grp1")
			So(grp.User(), ShouldBeNil)
			So(grp.Group(), ShouldEqual, grp)

			grp = &Group{name: "grp2"}
			grp.container = gtl
			err = grp.MarkAsRepoGroup()
			So(err, ShouldBeNil)
			So(grp.IsUndefined(), ShouldBeFalse)
			So(grp.IsUsers(), ShouldBeFalse)
			err = grp.markAsUserGroup()
			So(err.Error(), ShouldEqual, "group 'grp2' is a repos group, not a user one")
			So(grp.GetName(), ShouldEqual, "grp2")
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
			So(fmt.Sprintf("%v", grp.GetMembers()), ShouldEqual, "[user1]")
			usr := grp.GetUsers()[0]
			So(usr, ShouldNotBeNil)
			So(usr.GetName(), ShouldEqual, "user1")
			So(fmt.Sprintf("%v", usr.GetMembers()), ShouldEqual, "[]")
			So(usr.User(), ShouldEqual, usr)
			So(usr.Group(), ShouldBeNil)
		})

		Convey("Repos can be added", func() {
			gtl := NewGitolite()
			grp := &Group{}
			grp.members = append(grp.members, "repo1")
			grp.container = gtl
			var err error
			err = grp.MarkAsRepoGroup()
			So(err, ShouldBeNil)
			So(gtl.NbRepos(), ShouldEqual, 1)
			So(gtl.NbUsers(), ShouldEqual, 0)
			So(len(grp.GetUsers()), ShouldEqual, 0)
			So(fmt.Sprintf("%v", grp.GetMembers()), ShouldEqual, "[repo1]")
		})

		Convey("Comments can be added", func() {
			cmt := &Comment{}
			So(len(cmt.comments), ShouldEqual, 0)
			cmt.AddComment("test")
			So(len(cmt.comments), ShouldEqual, 1)
			So(cmt.comments[0], ShouldEqual, "test")
			cmt.AddComment("  test1")
			cmt.AddComment("   # test2  ")
			So(len(cmt.comments), ShouldEqual, 3)
			So(cmt.comments[1], ShouldEqual, "test1")
			So(cmt.comments[2], ShouldEqual, "# test2")
		})

	})
}
