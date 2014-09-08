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

			err = gtl.AddUserOrRepoGroup("grp1", []string{"u1", "u2"}, &Comment{[]string{"duplicate group"}})
			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldEqual, "Duplicate group name 'grp1'")
			So(gtl.NbGroupUsers(), ShouldEqual, 1)

			err = gtl.AddUserOrRepoGroup("grp2", []string{"u1", "u2", "u1"}, &Comment{[]string{"duplicate user"}})
			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldEqual, "Duplicate group element name 'u1'")
			So(gtl.NbGroupUsers(), ShouldEqual, 1)

			err = gtl.AddUserOrRepoGroup("grp2", []string{"u1", "u2"}, &Comment{[]string{"legit user group"}})
			So(err, ShouldBeNil)
			grp = gtl.GetGroup("grp2")
			err = grp.markAsUserGroup()
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
			So(len(grp.GetRepos()), ShouldEqual, 1)
			addRepoFromName(grp, "repo2", gtl)
			addRepoFromName(grp, "repo2", gtl)
			So(len(grp.GetRepos()), ShouldEqual, 2)
			So(grp.GetRepos()[0].GetName(), ShouldEqual, "repo1")

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

			grp := &Group{name: "grp1", cmt: &Comment{[]string{"grp1 comment"}}}
			usr = &User{"u21"}
			grp.addUser(usr)
			So(grp.Comment().String(), ShouldEqual, `grp1 comment
`)

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
			So(fmt.Sprintf("%v", gtl.namesToGroups), ShouldEqual, "map[]")

			reposgrp := &Group{name: "repogrp", container: gtl, members: []string{"repo1", "u4", "repo2"}}
			// gtl.addReposGroup(reposgrp)
			reposgrp.MarkAsRepoGroup()
			So(gtl.NbRepos(), ShouldEqual, 3)
			err := gtl.AddUserToRule(rule, "u4")
			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldStartWith, "user name 'u4' already used in a repo group")

			err = gtl.AddUserOrRepoGroup("grp4", []string{"u41", "u42"}, &Comment{[]string{"legit user group4"}})
			So(err, ShouldBeNil)
			grp = gtl.GetGroup("grp4")
			err = grp.markAsUserGroup()
			So(err, ShouldBeNil)
			So(gtl.NbGroupUsers(), ShouldEqual, 2)
			So(gtl.NbUsers(), ShouldEqual, 6)
			// So(fmt.Sprintf("%v", gtl.users), ShouldEqual, "e")
			err = gtl.AddUserGroupToRule(rule, "grp4")
			So(err, ShouldBeNil)
			So(gtl.NbGroupUsers(), ShouldEqual, 2)
			So(gtl.NbUsers(), ShouldEqual, 6)

			err = gtl.AddUserGroupToRule(rule, "grp5")
			So(err, ShouldBeNil)

			grp = &Group{name: "@all", container: gtl}
			grp.MarkAsRepoGroup()
			err = gtl.AddUserGroupToRule(rule, "@all")
			So(err, ShouldBeNil)

			err = gtl.AddUserGroupToRule(rule, "repogrp")
			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldEqual, "user group 'repogrp' named after a repo group")

			// define a user group *after* being used in a rule
			err = gtl.AddUserGroupToRule(rule, "@grpusers")
			So(err, ShouldBeNil)
			So(gtl.NbGroupUsers(), ShouldEqual, 5)
			So(gtl.NbUsers(), ShouldEqual, 6)

			err = gtl.AddUserOrRepoGroup("@grpusers", []string{"u41", "u42"}, &Comment{[]string{"legit user @grpusers"}})
			So(err, ShouldBeNil)
			grp = gtl.GetGroup("grp4")
			So(len(grp.GetUsers()), ShouldEqual, 2)
		})

		Convey("Configs can be added", func() {
			gtl := NewGitolite()
			So(gtl.NbConfigs(), ShouldEqual, 0)

			cfg, err := gtl.AddConfig([]string{"@grprepo"}, &Comment{[]string{"@grprepo comment"}})
			So(cfg, ShouldBeNil)
			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldEqual, "repo group name '@grprepo' undefined")

			cfg, err = gtl.AddConfig([]string{"@all"}, &Comment{[]string{"@grprepo comment"}})
			So(cfg, ShouldNotBeNil)
			So(err, ShouldBeNil)

			gtl.AddConfig([]string{"repo1", "repo2"}, &Comment{[]string{"cfg1 comment"}})
			So(gtl.NbConfigs(), ShouldEqual, 2)
			So(fmt.Sprintf("%v", gtl.GetConfigsForRepo("repo1")), ShouldEqual, "[config [repo 'repo1' repo 'repo2'] => []]")
			So(len(gtl.GetConfigsForRepos([]string{})), ShouldEqual, 0)
			So(gtl.NbRepos(), ShouldEqual, 2)

			cfg = gtl.GetConfigsForRepo("repo1")[0]
			So(len(cfg.GetRepos()), ShouldEqual, 2)
			So(len(cfg.Rules()), ShouldEqual, 0)

			reposgrp := &Group{name: "@repogrp1", container: gtl, members: []string{"repo11", "repo12"}}
			gtl.addReposGroup(reposgrp)
			So(gtl.NbRepos(), ShouldEqual, 4)

			//reposusr := &Group{name: "@usrgrp1", container: gtl, members: []string{"user11", "user12"}}
			gtl.AddUserOrRepoGroup("@usrgrp1", []string{"user11", "user12"}, &Comment{[]string{"usrgrp1 comment"}})
			gtl.AddUserOrRepoGroup("@usrgrp2", []string{}, &Comment{[]string{"usrgrp2 comment"}})
			grp := gtl.GetGroup("@usrgrp1")
			err = grp.markAsUserGroup()
			So(err, ShouldBeNil)
			grp = gtl.GetGroup("@usrgrp2")
			err = grp.markAsUserGroup()
			So(err, ShouldBeNil)
			So(gtl.NbUsers(), ShouldEqual, 2)
			So(gtl.NbRepos(), ShouldEqual, 4)

			cfg2, err := gtl.AddConfig([]string{"@usrgrp1"}, &Comment{[]string{"cfg2 comment"}})
			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldEqual, "group '@usrgrp1' is a users group, not a repo one")
			So(cfg2, ShouldBeNil)
			So(gtl.NbRepos(), ShouldEqual, 4)
			So(len(gtl.Configs()), ShouldEqual, 2)

			cfg2, err = gtl.AddConfig([]string{"@repounknown"}, &Comment{[]string{"cfg2 comment"}})
			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldEqual, "repo group name '@repounknown' undefined")
			So(cfg2, ShouldBeNil)

			cfg2, err = gtl.AddConfig([]string{"user11"}, &Comment{[]string{"cfg2 comment"}})
			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldEqual, `repo name 'user11' already used in a user group
group '@usrgrp1' is a users group, not a repo one`)
			So(cfg2, ShouldBeNil)
			So(gtl.NbRepos(), ShouldEqual, 4)

			cfg2, err = gtl.AddConfig([]string{"@repogrp1"}, &Comment{[]string{"cfg2 comment"}})
			So(err, ShouldBeNil)
			So(cfg2.Comment().String(), ShouldEqual, `cfg2 comment
`)

			cfga, err := gtl.AddConfig([]string{"gitolite-admin"}, &Comment{[]string{"ga comment"}})
			So(err, ShouldBeNil)
			So(cfga, ShouldNotBeNil)

			err = cfg2.SetDesc("cfg2 desc", &Comment{[]string{"cfg2 desc comment"}})
			So(err, ShouldBeNil)
			So(cfg2.Desc(), ShouldEqual, "cfg2 desc")
			err = cfg2.SetDesc("cfg2 desc", &Comment{[]string{"cfg2 desc comment"}})
			So(err.Error(), ShouldEqual, "No more than one desc per config")

			So(len(gtl.reposToConfigs["repo11"]), ShouldEqual, 0)
			cmt := &Comment{[]string{"rule comment"}}
			rule := NewRule("RW", "test", cmt)
			grp = gtl.GetGroup("@usrgrp1")
			rule.addGroup(grp)
			gtl.AddRuleToConfig(rule, cfg2)
			So(len(gtl.reposToConfigs["repo11"]), ShouldEqual, 1)
			// So(fmt.Sprintf("%v", gtl.reposToConfigs), ShouldEqual, "z")
			So(len(cfg2.Rules()), ShouldEqual, 1)
			gtl.AddRuleToConfig(rule, cfg2)
			So(len(gtl.reposToConfigs["repo11"]), ShouldEqual, 1)
			So(len(cfg2.Rules()), ShouldEqual, 1)
			// So(fmt.Sprintf("%v", gtl.reposToConfigs), ShouldEqual, "z")

			So(gtl.Print(), ShouldEqual, `@repogrp1 = repo11 repo12
# usrgrp1 comment
@usrgrp1 = user11 user12
# usrgrp2 comment
@usrgrp2 =
# ga comment
repo gitolite-admin
# @grprepo comment
repo
# cfg1 comment
repo repo1 repo2
# cfg2 comment
repo repo11 repo12
# cfg2 desc comment
desc = cfg2 desc
# rule comment
RW test = @usrgrp1
`)

			So(gtl.String(), ShouldEqual, `NbGroups: 4 [@all, @repogrp1, @usrgrp1, @usrgrp2]
NbRepoGroups: 3 [@all, @repogrp1, @repogrp1]
NbRepos: 5 [repo 'repo1' repo 'repo2' repo 'repo11' repo 'repo12' repo 'gitolite-admin']
NbUsers: 2 [user 'user11' user 'user12']
NbUserGroups: 2 [@usrgrp1, @usrgrp2]
NbConfigs: 4 [config [] => [], config [repo 'repo1' repo 'repo2'] => [], config [repo 'repo11' repo 'repo12'] => [RW test = @usrgrp1 (user11, user12)], config [repo 'gitolite-admin'] => []]
namesToGroups: 6 [@usrgrp1 => [group '@usrgrp1'<users>: [user11 user12]], @usrgrp2 => [group '@usrgrp2'<users>: []], repo11 => [group '@repogrp1'<repos>: [repo11 repo12]], repo12 => [group '@repogrp1'<repos>: [repo11 repo12]], user11 => [group '@usrgrp1'<users>: [user11 user12]], user12 => [group '@usrgrp1'<users>: [user11 user12]]]
reposToConfigs: 2 [repo11 => [config [repo 'repo11' repo 'repo12'] => [RW test = @usrgrp1 (user11, user12)]], repo12 => [config [repo 'repo11' repo 'repo12'] => [RW test = @usrgrp1 (user11, user12)]]]
`)

			r, err := gtl.Rules("repo12")
			So(err, ShouldBeNil)
			So(len(r), ShouldEqual, 1)

		})
	})
}
