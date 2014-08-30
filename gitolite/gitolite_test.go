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
			So(grp.kind.String(), ShouldEqual, "[undefined]")
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
			So(grp.kind.String(), ShouldEqual, "<users>")
			err = grp.MarkAsRepoGroup()
			So(err.Error(), ShouldEqual, "group 'grp1' is a users group, not a repo one")
			So(grp.GetName(), ShouldEqual, "grp1")
			So(grp.User(), ShouldBeNil)
			So(grp.Group(), ShouldEqual, grp)
			So(gtl.NbGroup(), ShouldEqual, 1)
			So(gtl.NbGroupRepos(), ShouldEqual, 0)
			So(gtl.NbGroupUsers(), ShouldEqual, 1)

			grp = &Group{name: "grp2"}
			grp.container = gtl
			err = grp.MarkAsRepoGroup()
			So(err, ShouldBeNil)
			So(grp.IsUndefined(), ShouldBeFalse)
			So(grp.IsUsers(), ShouldBeFalse)
			So(grp.kind.String(), ShouldEqual, "<repos>")
			err = grp.markAsUserGroup()
			So(err.Error(), ShouldEqual, "group 'grp2' is a repos group, not a user one")
			So(grp.GetName(), ShouldEqual, "grp2")
			So(gtl.NbGroup(), ShouldEqual, 2)
			So(gtl.NbGroupRepos(), ShouldEqual, 1)
			So(gtl.NbGroupUsers(), ShouldEqual, 1)

			So(gtl.GetGroup("grp1"), ShouldNotBeNil)
			So(gtl.GetGroup("grp2"), ShouldEqual, grp)
			So(gtl.GetGroup("grp3"), ShouldBeNil)
		})
		Convey("Users can be added", func() {
			gtl := NewGitolite()
			grp := &Group{name: "grp1"}
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
			So(gtl.NbGroup(), ShouldEqual, 1)
			So(gtl.NbGroupRepos(), ShouldEqual, 0)
			So(gtl.NbGroupUsers(), ShouldEqual, 1)

			addUserFromName(grp, "user1", gtl)
			So(len(grp.GetUsers()), ShouldEqual, 1)

			err = gtl.AddUserGroup("grp1", []string{"u1", "u2"}, &Comment{[]string{"duplicate group"}})
			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldEqual, "Duplicate group name 'grp1'")
			So(gtl.NbGroupUsers(), ShouldEqual, 1)

			err = gtl.AddUserGroup("grp2", []string{"u1", "u2", "u1"}, &Comment{[]string{"duplicate user"}})
			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldEqual, "Duplicate group element name 'u1'")
			So(gtl.NbGroupUsers(), ShouldEqual, 1)

			err = gtl.AddUserGroup("grp2", []string{"u1", "u2"}, &Comment{[]string{"legit user group"}})
			So(err, ShouldBeNil)
			So(gtl.NbGroupUsers(), ShouldEqual, 2)
		})

		Convey("Repos can be added", func() {
			gtl := NewGitolite()
			grp := &Group{name: "grp1"}
			grp.members = append(grp.members, "repo1")
			grp.container = gtl
			var err error
			err = grp.MarkAsRepoGroup()
			So(err, ShouldBeNil)
			So(gtl.NbRepos(), ShouldEqual, 1)
			So(gtl.NbUsers(), ShouldEqual, 0)
			So(len(grp.GetUsers()), ShouldEqual, 0)
			So(fmt.Sprintf("%v", grp.GetMembers()), ShouldEqual, "[repo1]")
			So(gtl.NbGroup(), ShouldEqual, 1)
			So(gtl.NbGroupRepos(), ShouldEqual, 1)
			So(gtl.NbGroupUsers(), ShouldEqual, 0)

			addRepoFromName(grp, "repo1", gtl)
			So(gtl.NbGroupRepos(), ShouldEqual, 1)
			So(len(grp.getRepos()), ShouldEqual, 1)
			addRepoFromName(grp, "repo2", gtl)
			addRepoFromName(grp, "repo2", gtl)
			So(len(grp.getRepos()), ShouldEqual, 2)

			repo := gtl.repos[0]
			So(repo, ShouldNotBeNil)
			So(repo.String(), ShouldEqual, `repo 'repo1'`)
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

			So(cmt.String(), ShouldEqual, `test
test1
# test2
`)
		})

		Convey("Rules can be added", func() {
			cmt := &Comment{[]string{"rule comment"}}
			rule := NewRule("RW", "test", cmt)
			So(rule.Access(), ShouldEqual, "RW")
			So(rule.Param(), ShouldEqual, "test")
			So(rule.Comment().comments[0], ShouldEqual, "rule comment")
			So(rule.HasAnyUserOrGroup(), ShouldBeFalse)

			usr := &User{"u1"}
			rule.addUser(usr)
			So(len(rule.usersOrGroups), ShouldEqual, 1)
			So(len(rule.GetUsers()), ShouldEqual, 1)
			So(rule.HasAnyUserOrGroup(), ShouldBeTrue)

			grp := &Group{name: "grp1"}
			usr = &User{"u21"}
			grp.addUser(usr)

			rule.addGroup(grp)
			// grp is still undefined: its user doesn't count yet
			So(len(rule.GetUsers()), ShouldEqual, 1)
			rule.addGroup(grp)
			So(len(rule.GetUsers()), ShouldEqual, 1)
			So(rule.HasAnyUserOrGroup(), ShouldBeTrue)

			gtl := NewGitolite()
			grp.container = gtl
			grp.markAsUserGroup()
			// grp is defined as user group: its user does count
			So(len(rule.GetUsers()), ShouldEqual, 2)
			So(gtl.NbUsers(), ShouldEqual, 1)

			So(rule.String(), ShouldEqual, `RW test = u1, grp1 (u21)`)
			usr = &User{"u22"}
			grp.addUser(usr)
			gtl.addUser(usr)
			So(rule.String(), ShouldEqual, `RW test = u1, grp1 (u21, u22)`)

			gtl.AddUserToRule(rule, "u3")
			So(rule.String(), ShouldEqual, `RW test = u1, grp1 (u21, u22), u3`)
			// u1 was never added to glt, only to rule
			So(gtl.NbUsers(), ShouldEqual, 3)
			// Rule was never properly added to gitolite or any conf
			So(fmt.Sprintf("%v", gtl.namesToGroups), ShouldEqual, "[]")
		})
	})
}
