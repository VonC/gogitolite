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
			gtl := NewGitolite(nil)
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
			gtl := NewGitolite(nil)
			grp := &Group{name: "@grp1"}
			grp.container = gtl
			var err error
			err = grp.markAsUserGroup()
			So(err, ShouldBeNil)
			So(grp.IsUndefined(), ShouldBeFalse)
			So(grp.IsUsers(), ShouldBeTrue)
			So(grp.kind.String(), ShouldEqual, "<users>")
			err = grp.MarkAsRepoGroup()
			So(err.Error(), ShouldEqual, "group '@grp1' is a users group, not a repo one")
			So(grp.GetName(), ShouldEqual, "@grp1")
			So(grp.User(), ShouldBeNil)
			So(grp.Group(), ShouldEqual, grp)
			So(gtl.NbGroup(), ShouldEqual, 1)
			So(gtl.NbRepoGroups(), ShouldEqual, 0)
			So(gtl.NbUserGroups(), ShouldEqual, 1)

			grp = &Group{name: "@grp2"}
			grp.container = gtl
			err = grp.MarkAsRepoGroup()
			So(err, ShouldBeNil)
			So(grp.IsUndefined(), ShouldBeFalse)
			So(grp.IsUsers(), ShouldBeFalse)
			So(grp.kind.String(), ShouldEqual, "<repos>")
			err = grp.markAsUserGroup()
			So(err.Error(), ShouldEqual, "group '@grp2' is a repos group, not a user one")
			So(grp.GetName(), ShouldEqual, "@grp2")
			So(gtl.NbGroup(), ShouldEqual, 2)
			So(gtl.NbRepoGroups(), ShouldEqual, 1)
			So(gtl.NbUserGroups(), ShouldEqual, 1)

			So(gtl.GetGroup("@grp1"), ShouldNotBeNil)
			So(gtl.GetGroup("@grp2"), ShouldEqual, grp)
			So(gtl.GetGroup("@grp3"), ShouldBeNil)
		})
		Convey("Users can be added", func() {
			gtl := NewGitolite(nil)
			grp := &Group{name: "@grp1"}
			grp.members = append(grp.members, "user1")
			grp.members = append(grp.members, "u2")
			grp.container = gtl
			var err error
			err = grp.markAsUserGroup()
			So(err, ShouldBeNil)
			So(gtl.NbUsers(), ShouldEqual, 2)
			So(fmt.Sprintf("%v", grp.GetMembers()), ShouldEqual, "[user1 u2]")
			usr := grp.GetUsersOrGroups()[0]
			So(usr, ShouldNotBeNil)
			So(usr.GetName(), ShouldEqual, "user1")
			So(fmt.Sprintf("%v", usr.GetMembers()), ShouldEqual, "[]")
			So(usr.User(), ShouldEqual, usr)
			So(usr.Group(), ShouldBeNil)
			So(gtl.NbGroup(), ShouldEqual, 1)
			So(gtl.NbRepoGroups(), ShouldEqual, 0)
			So(gtl.NbUserGroups(), ShouldEqual, 1)
			So(gtl.NbUsersOrGroups(), ShouldEqual, 3)

			addUserOrGroupFromName(grp, "user1", gtl)
			So(len(grp.GetUsersOrGroups()), ShouldEqual, 2)
			So(gtl.NbUsersOrGroups(), ShouldEqual, 3)
			So(gtl.userOrGroupFromName("user1"), ShouldNotBeNil)
			So(gtl.userOrGroupFromName("user1b"), ShouldBeNil)

			err = gtl.AddUserOrRepoGroup("@grp1", []string{"u1", "u2"}, &Comment{[]string{"duplicate group"}, ""})
			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldEqual, "Duplicate group name '@grp1'")
			So(gtl.NbUserGroups(), ShouldEqual, 1)
			So(gtl.NbUsersOrGroups(), ShouldEqual, 3)

			err = gtl.AddUserOrRepoGroup("@grp2", []string{"u1", "u2", "u1"}, &Comment{[]string{"duplicate user"}, ""})
			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldEqual, "Duplicate group element name 'u1'")
			So(gtl.NbUserGroups(), ShouldEqual, 1)

			err = gtl.AddUserOrRepoGroup("@grp2", []string{"u1", "u2", "@grp1"}, &Comment{[]string{"legit user group"}, ""})
			So(err, ShouldBeNil)
			grp = gtl.GetGroup("@grp2")
			err = grp.markAsUserGroup()
			So(err, ShouldBeNil)
			So(gtl.NbUserGroups(), ShouldEqual, 2)
			So(len(grp.GetAllUsers()), ShouldEqual, 3)
			So(gtl.NbUsersOrGroups(), ShouldEqual, 5)
			grp1 := gtl.getGroup("@grp1")

			So(len(gtl.groupsFromUserOrGroup(nil)), ShouldEqual, 0)
			So(len(gtl.groupsFromUserOrGroup(grp)), ShouldEqual, 0)
			So(len(gtl.groupsFromUserOrGroup(grp1)), ShouldEqual, 1)
		})

		Convey("Repos can be added", func() {
			gtl := NewGitolite(nil)
			grp := &Group{name: "@grp1"}
			grp.members = append(grp.members, "repo1")
			grp.container = gtl
			var err error
			err = grp.MarkAsRepoGroup()
			So(err, ShouldBeNil)
			So(gtl.NbRepos(), ShouldEqual, 1)
			So(gtl.NbUsers(), ShouldEqual, 0)
			So(len(grp.GetAllRepos()), ShouldEqual, 1)
			So(len(grp.GetUsersOrGroups()), ShouldEqual, 0)
			So(fmt.Sprintf("%v", grp.GetMembers()), ShouldEqual, "[repo1]")
			So(gtl.NbGroup(), ShouldEqual, 1)
			So(gtl.NbRepoGroups(), ShouldEqual, 1)
			So(gtl.NbUserGroups(), ShouldEqual, 0)

			addRepoOrGroupFromName(grp, "repo1", gtl)
			So(gtl.NbRepoGroups(), ShouldEqual, 1)
			So(len(grp.GetAllRepos()), ShouldEqual, 1)
			So(len(grp.GetReposOrGroups()), ShouldEqual, 1)
			addRepoOrGroupFromName(grp, "repo2", gtl)
			addRepoOrGroupFromName(grp, "repo2", gtl)
			So(len(grp.GetReposOrGroups()), ShouldEqual, 2)
			So(len(grp.GetAllRepos()), ShouldEqual, 2)
			So(grp.GetReposOrGroups()[0].GetName(), ShouldEqual, "repo1")
			So(gtl.repoOrGroupFromName("repo1"), ShouldNotBeNil)
			So(gtl.repoOrGroupFromName("repo1b"), ShouldBeNil)

			repo := gtl.reposOrGroups[1]
			So(repo, ShouldNotBeNil)
			So(repo.String(), ShouldEqual, `repo 'repo1'`)
			So(len(repo.GetMembers()), ShouldEqual, 0)
			So(len(repo.Repo().GetReposOrGroups()), ShouldEqual, 1)

			addRepoOrGroupFromName(gtl, "@grp2", gtl)
			So(gtl.NbReposOrGroups(), ShouldEqual, 4)
			grp2 := gtl.GetRepoGroup("@grp2").Group()
			So(grp2, ShouldNotBeNil)
			grp2n := gtl.GetRepoGroup("@grp22").Group()
			So(grp2n, ShouldBeNil)
			addRepoOrGroupFromName(grp2, "repo22", gtl)
			So(gtl.NbReposOrGroups(), ShouldEqual, 5)
			addRepoOrGroupFromName(grp2, "@grp1", gtl)
			So(gtl.NbReposOrGroups(), ShouldEqual, 5)
			So(len(grp2.GetAllRepos()), ShouldEqual, 3)
			So(grp2.hasRepoOrGroup("repo1"), ShouldBeTrue)
			So(grp2.hasRepoOrGroup("repo22"), ShouldBeTrue)
			So(grp.hasRepoOrGroup("repo22"), ShouldBeFalse)

			addRepoOrGroupFromName(grp2, "repo1", gtl)

			So(gtl.NbReposOrGroups(), ShouldEqual, 5)
			So(len(grp2.GetAllRepos()), ShouldEqual, 3)
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
			cmt.SetSameLine("   # same line test  ")
			So(cmt.SameLine(), ShouldEqual, "# same line test")

			So(cmt.String(), ShouldEqual, `test
test1
# test2
`)
		})

		Convey("Rules can be added", func() {
			cmt := &Comment{[]string{"rule comment"}, ""}
			rule := NewRule("RW", "test", cmt)
			So(rule.Access(), ShouldEqual, "RW")
			So(rule.Param(), ShouldEqual, "test")
			So(rule.Comment().comments[0], ShouldEqual, "rule comment")
			So(rule.HasAnyUserOrGroup(), ShouldBeFalse)
			So(len(rule.GetAllUsers()), ShouldEqual, 0)

			usr := &User{"u1"}
			rule.addUserOrGroup(usr)
			So(len(rule.usersOrGroups), ShouldEqual, 1)
			So(len(rule.GetUsersOrGroups()), ShouldEqual, 1)
			So(rule.HasAnyUserOrGroup(), ShouldBeTrue)
			So(len(rule.GetAllUsers()), ShouldEqual, 1)

			grp := &Group{name: "@grp1", cmt: &Comment{[]string{"@grp1 comment"}, ""}}
			usr = &User{"u21"}
			grp.addUserOrGroup(usr)
			So(grp.Comment().String(), ShouldEqual, `@grp1 comment
`)

			rule.addGroup(grp)
			// grp is still undefined: its user doesn't count yet
			So(len(rule.GetUsersOrGroups()), ShouldEqual, 2)
			So(len(rule.GetUsersFirstOrGroups()), ShouldEqual, 2)
			rule.addGroup(grp)
			So(len(rule.GetUsersOrGroups()), ShouldEqual, 2)
			So(len(rule.GetUsersFirstOrGroups()), ShouldEqual, 2)
			So(rule.HasAnyUserOrGroup(), ShouldBeTrue)
			So(len(rule.GetAllUsers()), ShouldEqual, 1)

			gtl := NewGitolite(nil)
			grp.container = gtl
			//fmt.Println(grp.String())
			grp.markAsUserGroup()
			// grp is defined as user group: its user does count
			So(len(rule.GetUsersOrGroups()), ShouldEqual, 2)
			So(gtl.NbUsers(), ShouldEqual, 1)
			//fmt.Println(gtl.String())
			//os.Exit(1)
			So(len(rule.GetUsersFirstOrGroups()), ShouldEqual, 2)
			So(len(rule.GetAllUsers()), ShouldEqual, 2)

			grpp := &Group{name: "@grp1", cmt: &Comment{[]string{"@grp1 comment"}, ""}}
			//usr = &User{"u22"}
			//grpp.addUserOrGroup(usr)
			gtl.addGroup(grpp)

			So(rule.String(), ShouldEqual, `RW test = u1, @grp1 (u21)`)
			So(rule.IsNakedRW(), ShouldBeFalse)
			usr = &User{"u22"}
			grp.addUserOrGroup(usr)
			gtl.addUserOrGroup(usr)
			gtl.addUserOrGroup(usr)
			So(rule.String(), ShouldEqual, `RW test = u1, @grp1 (u21, u22)`)

			gtl.AddUserOrGroupToRule(rule, "u3")
			So(rule.String(), ShouldEqual, `RW test = u1, @grp1 (u21, u22), u3`)
			// u1 was never added to glt, only to rule
			So(gtl.NbUsers(), ShouldEqual, 3)
			// Rule was never properly added to gitolite or any conf
			// So(fmt.Sprintf("%v", gtl.namesToGroups), ShouldEqual, "map[]")

			reposgrp := &Group{name: "@repogrp", container: gtl, members: []string{"repo1", "u4", "repo2"}}
			// gtl.addReposGroup(reposgrp)
			reposgrp.MarkAsRepoGroup()
			So(gtl.NbRepos(), ShouldEqual, 3)

			err := gtl.AddUserOrGroupToRule(rule, "u4")
			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldStartWith, "user or user group name 'u4' already used in a repo group")

			err = gtl.AddUserOrRepoGroup("@grp4", []string{"u41", "u42"}, &Comment{[]string{"legit user group4"}, ""})
			So(err, ShouldBeNil)

			grp = gtl.GetGroup("@grp4")
			err = grp.markAsUserGroup()
			So(err, ShouldBeNil)

			So(gtl.NbUserGroups(), ShouldEqual, 2)
			So(gtl.NbUsers(), ShouldEqual, 5)
			// So(fmt.Sprintf("%v", gtl.users), ShouldEqual, "e")
			err = gtl.AddUserOrGroupToRule(rule, "@grp4")
			So(err, ShouldBeNil)
			So(gtl.NbUserGroups(), ShouldEqual, 2)
			So(gtl.NbUsers(), ShouldEqual, 5)

			err = gtl.AddUserOrGroupToRule(rule, "@grp5")
			So(err, ShouldBeNil)

			grp = &Group{name: "@all", container: gtl}
			grp.MarkAsRepoGroup()
			err = gtl.AddUserOrGroupToRule(rule, "@all")
			So(err, ShouldBeNil)

			err = gtl.AddUserOrGroupToRule(rule, "@repogrp")
			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldEqual, "user or user group name '@repogrp' already used in a repo group")

			// define a user group *after* being used in a rule
			err = gtl.AddUserOrGroupToRule(rule, "@grpusers")
			So(err, ShouldBeNil)
			So(gtl.NbUserGroups(), ShouldEqual, 5)
			So(gtl.NbUsers(), ShouldEqual, 5)

			err = gtl.AddUserOrRepoGroup("@grpusers", []string{"u41", "u42"}, &Comment{[]string{"legit user @grpusers"}, ""})
			So(err, ShouldBeNil)
			grp = gtl.GetGroup("@grp4")
			So(len(grp.GetUsersOrGroups()), ShouldEqual, 2)
		})

		Convey("Configs can be added", func() {
			gtl := NewGitolite(nil)
			gtl2 := NewGitolite(gtl)
			So(gtl.NbConfigs(), ShouldEqual, 0)
			//fmt.Println("\nCFG: ", gtl)

			cfg, err := gtl.AddConfig([]string{"@grprepo"}, &Comment{[]string{"@grprepo comment"}, ""})
			So(cfg, ShouldBeNil)
			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldEqual, "repo group name '@grprepo' undefined")
			//fmt.Println("\nCFG: ", gtl)

			cfg, err = gtl.AddConfig([]string{"@all"}, &Comment{[]string{"@all comment"}, ""})
			So(cfg, ShouldNotBeNil)
			So(err, ShouldBeNil)
			//fmt.Println("\nCFG: ", gtl)

			gtl.AddConfig([]string{"repo1", "repo2"}, &Comment{[]string{"cfg1 comment"}, ""})
			So(gtl.NbConfigs(), ShouldEqual, 2)
			So(fmt.Sprintf("%v", gtl.GetConfigsForRepo("repo1")), ShouldEqual, "[config [repo 'repo1' repo 'repo2'] => rules []]")
			So(len(gtl.GetConfigsForRepos([]string{})), ShouldEqual, 0)
			So(gtl.NbRepos(), ShouldEqual, 2)

			cfg = gtl.GetConfigsForRepo("repo1")[0]
			So(len(cfg.GetReposOrGroups()), ShouldEqual, 2)
			So(len(cfg.Rules()), ShouldEqual, 0)

			reposgrp := &Group{name: "@repogrp1", container: gtl, members: []string{"repo11", "repo12"}}
			gtl.addRepoOrGroup(reposgrp)
			So(gtl.NbRepos(), ShouldEqual, 4)
			So(len(gtl.configsFromRepoOrGroup(nil)), ShouldEqual, 0)
			So(len(gtl.configsFromRepoOrGroup(reposgrp)), ShouldEqual, 0)

			//reposusr := &Group{name: "@usrgrp1", container: gtl, members: []string{"user11", "user12"}}
			gtl.AddUserOrRepoGroup("@usrgrp1", []string{"user11", "user12"}, &Comment{[]string{"usrgrp1 comment"}, ""})
			gtl.AddUserOrRepoGroup("@usrgrp2", []string{}, &Comment{[]string{"usrgrp2 comment"}, ""})
			grp := gtl.GetGroup("@usrgrp1")
			err = grp.markAsUserGroup()
			//user11 := grp.GetAllUsers()[0]
			So(err, ShouldBeNil)
			grp = gtl.GetGroup("@usrgrp2")
			err = grp.markAsUserGroup()
			So(err, ShouldBeNil)
			So(gtl.NbUsers(), ShouldEqual, 2)
			So(gtl.NbRepos(), ShouldEqual, 4)

			cfg2, err := gtl.AddConfig([]string{"@usrgrp1"}, &Comment{[]string{"cfg2 comment"}, ""})
			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldEqual, "group '@usrgrp1' is a users group, not a repo one")
			So(cfg2, ShouldBeNil)
			So(gtl.NbRepos(), ShouldEqual, 4)
			So(len(gtl.Configs()), ShouldEqual, 2)

			cfg2, err = gtl.AddConfig([]string{"@repounknown"}, &Comment{[]string{"cfg2 comment"}, ""})
			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldEqual, "repo group name '@repounknown' undefined")
			So(cfg2, ShouldBeNil)

			cfg2, err = gtl.AddConfig([]string{"user11"}, &Comment{[]string{"cfg2 comment"}, ""})
			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldEqual, `repo name 'user11' already used in a user group
group '@usrgrp1' is a users group, not a repo one`)
			So(cfg2, ShouldBeNil)
			So(gtl.NbRepos(), ShouldEqual, 4)

			cfg2, err = gtl.AddConfig([]string{"@repogrp1"}, &Comment{[]string{"cfg2 comment"}, ""})
			So(err, ShouldBeNil)
			So(cfg2.Comment().String(), ShouldEqual, `cfg2 comment
`)
			cfg2b, err := gtl2.AddConfig([]string{"@repogrp1"}, &Comment{[]string{"cfg2b comment"}, ""})
			So(err, ShouldBeNil)
			So(cfg2b.Comment().String(), ShouldEqual, `cfg2b comment
`)
			So(len(gtl.configsFromRepoOrGroup(reposgrp)), ShouldEqual, 1)

			cfga, err := gtl.AddConfig([]string{"gitolite-admin"}, &Comment{[]string{"ga comment"}, ""})
			So(err, ShouldBeNil)
			So(cfga, ShouldNotBeNil)

			err = cfg2.SetDesc("cfg2 desc", &Comment{[]string{"cfg2 desc comment"}, ""})
			So(err, ShouldBeNil)
			So(cfg2.Desc(), ShouldEqual, "cfg2 desc")
			err = cfg2.SetDesc("cfg2 desc", &Comment{[]string{"cfg2 desc comment"}, ""})
			So(err.Error(), ShouldEqual, "No more than one desc per config")

			//fmt.Println("\nGTL ", gtl)
			//os.Exit(0)
			So(len(gtl.GetConfigsForRepo("repo11")), ShouldEqual, 1)
			cmt := &Comment{[]string{"rule comment"}, ""}
			cmt.sameLine = "test"
			rule := NewRule("RW", "test", cmt)
			grp = gtl.GetGroup("@usrgrp2")
			rule.addGroup(grp)
			gtl.AddRuleToConfig(rule, cfg2)
			gtl.AddUserOrGroupToRule(rule, "user11")
			gtl.GetConfigsForRepo("repo11")
			So(len(gtl.GetConfigsForRepo("repo11")), ShouldEqual, 1)
			// So(fmt.Sprintf("%v", gtl.reposToConfigs), ShouldEqual, "z")
			So(len(cfg2.Rules()), ShouldEqual, 1)
			gtl.AddRuleToConfig(rule, cfg2)
			So(len(gtl.GetConfigsForRepo("repo11")), ShouldEqual, 1)
			So(len(cfg2.Rules()), ShouldEqual, 1)
			// So(fmt.Sprintf("%v", gtl.reposToConfigs), ShouldEqual, "z")

			So(gtl.Print(), ShouldEqual, `# @all comment
repo @all

# cfg1 comment
repo repo1 repo2

@repogrp1 = repo11 repo12

# usrgrp1 comment
@usrgrp1 = user11 user12

# usrgrp2 comment
@usrgrp2 =

# cfg2 comment
repo @repogrp1
# cfg2 desc comment
    desc  = cfg2 desc
    # rule comment
    RW   test = @usrgrp2 user11 # test

# ga comment
repo gitolite-admin

`)

			So(gtl.String(), ShouldEqual, `NbGroups: 4 [@all, @repogrp1, @usrgrp1, @usrgrp2]
NbRepoOrGroups: 7 [@all, repo1, repo2, @repogrp1, repo11, repo12, gitolite-admin]
NbUserOrGroups: 4 [@usrgrp1, user11, user12, @usrgrp2]
NbConfigs: 4 [config [group '@all'<repos>: []] => rules [], config [repo 'repo1' repo 'repo2'] => rules [], config [group '@repogrp1'<repos>: [repo11 repo12]] => rules [RW test = @usrgrp2, user11], config [repo 'gitolite-admin'] => rules []]
`)

			r, err := gtl.Rules("repo12")
			So(err, ShouldBeNil)
			So(len(r), ShouldEqual, 1)

		})

		Convey("Subconfs can be added", func() {
			gtl := NewGitolite(nil)
			err := gtl.AddSubconf("(invalid conf")
			So(err, ShouldNotBeNil)

			err = gtl.AddSubconf("subs/*.conf")
			So(err, ShouldBeNil)
			So(len(gtl.Subconfs()), ShouldEqual, 1)

			err = gtl.AddSubconf("subs/*.conf")
			So(err, ShouldBeNil)
			So(len(gtl.Subconfs()), ShouldEqual, 1)

			err = gtl.AddSubconf("subs2/*.conf")
			So(err, ShouldBeNil)
			So(len(gtl.Subconfs()), ShouldEqual, 2)

		})
	})

}
