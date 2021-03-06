package reader

import (
	"fmt"
	"strings"
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

/*
# sample conf/gitolite.conf file

@staff              =   dilbert alice           # groups
@projects           =   foo bar

repo @projects baz                              # repos
    RW+             =   @staff                  # rules
    -       master  =   ashok
    RW              =   ashok
    R               =   wally

    option deny-rules           =   1           # options
    config hooks.emailprefix    = '[%GL_REPO] ' # git-config
*/

func TestRead(t *testing.T) {

	Convey("An empty reader means no repos", t, func() {
		gtl, _ := Read(nil)
		So(gtl.IsEmpty(), ShouldBeTrue)
		r := strings.NewReader("")
		gtl, err := Read(r)
		So(gtl.IsEmpty(), ShouldBeTrue)
		So(strings.Contains(err.Error(), ": comment, group or repo expected"), ShouldBeTrue)
	})

	Convey("If the content is not empty, it should declare a group or repo", t, func() {

		Convey("no comment, dummy content", func() {
			r := strings.NewReader("  foobar")
			gtl, err := Read(r)
			So(gtl.IsEmpty(), ShouldBeTrue)
			So(strings.Contains(err.Error(), ": group or repo expected"), ShouldBeTrue)
		})

		Convey("comments only, no content", func() {
			r := strings.NewReader(" # foo\n# bar")
			gtl, err := Read(r)
			So(gtl.IsEmpty(), ShouldBeTrue)
			So(strings.Contains(err.Error(), ": comment, group or repo expected"), ShouldBeTrue)
		})

	})

	Convey("An reader can read groups", t, func() {
		test = "ignorega"
		Convey("single group, followed by content", func() {
			r := strings.NewReader("  @developers     =   dilbert alice wally  \n#comment ")
			gtl, err := Read(r)
			So(err, ShouldBeNil)
			So(gtl.IsEmpty(), ShouldBeFalse)
			So(gtl.NbGroup(), ShouldEqual, 1)
		})
		Convey("single group, followed by no content", func() {
			r := strings.NewReader("  @developers2  =   dilbert alice wally2")
			gtl, err := Read(r)
			So(err, ShouldBeNil)
			So(gtl.IsEmpty(), ShouldBeFalse)
			So(gtl.NbGroup(), ShouldEqual, 1)
			So(fmt.Sprintf("%v", gtl.GetGroup("@developers2")), ShouldEqual, "group '@developers2'[undefined]: [dilbert alice wally2]")
		})
		Convey("An group name must be [a-zA-Z0-9_-]", func() {
			r := strings.NewReader("  @develop;ers     =   dilbert alice wally")
			gtl, err := Read(r)
			So(gtl.IsEmpty(), ShouldBeTrue)
			So(strings.Contains(err.Error(), ": Incorrect group declaration"), ShouldBeTrue)
		})
		Convey("An group name must be unique", func() {
			r := strings.NewReader("  @grp1     =   el1 elt2\n @grp1     =   el4 elt5")
			gtl, err := Read(r)
			So(gtl.NbGroup(), ShouldEqual, 1)
			So(strings.Contains(err.Error(), ": Duplicate group name"), ShouldBeTrue)
		})

		Convey("An group element  must be unique", func() {
			r := strings.NewReader("@grp1     =   elt1 elt2 elt1")
			gtl, err := Read(r)
			So(gtl.NbGroup(), ShouldEqual, 0)
			So(strings.Contains(err.Error(), ": Duplicate group element name"), ShouldBeTrue)
		})
	})

	Convey("An reader can read repos", t, func() {
		Convey("single repo", func() {
			r := strings.NewReader("  repo arepo1")
			gtl, err := Read(r)
			So(err, ShouldBeNil)
			So(gtl.IsEmpty(), ShouldBeFalse)
			So(gtl.NbRepos(), ShouldEqual, 1)
		})
		Convey("Repo names must be [a-zA-Z0-9_-]", func() {
			r := strings.NewReader("repo rep,1 rep2")
			gtl, err := Read(r)
			So(gtl.IsEmpty(), ShouldBeTrue)
			So(strings.Contains(err.Error(), ": Incorrect repo declaration"), ShouldBeTrue)
		})
		Convey("Repo names must be unique on one line", func() {
			r := strings.NewReader("repo rep1 rep2 rep1")
			gtl, err := Read(r)
			So(gtl.IsEmpty(), ShouldBeTrue)
			So(strings.Contains(err.Error(), ": Duplicate repo element name"), ShouldBeTrue)
		})
		Convey("Repo names can be part of a group", func() {
			r := strings.NewReader(
				`@grp1 = rep1 rep2
							 repo  rep1 rep2 rep3`)
			gtl, err := Read(r)
			So(err, ShouldBeNil)
			So(gtl.NbRepoGroups(), ShouldEqual, 1)
			grp := gtl.GetGroup("@grp1")
			So(grp, ShouldNotBeNil)
			So(len(grp.GetUsersOrGroups()), ShouldEqual, 0)
		})
	})

	Convey("An reader can read repo rules", t, func() {
		Convey("single rule", func() {
			r := strings.NewReader(
				`repo arepo1
								RW+ = user1 `)
			gtl, err := Read(r)
			So(err, ShouldBeNil)
			So(gtl.IsEmpty(), ShouldBeFalse)
			So(gtl.NbUsers(), ShouldEqual, 1)
		})

		Convey("at least one rule is expected", func() {
			r := strings.NewReader(
				`repo arepo1
								,,,`)
			_, err := Read(r)
			So(strings.Contains(err.Error(), ": At least one access rule expected"), ShouldBeTrue)
		})

		Convey("Access rule must be well formed: RW+- only", func() {
			r := strings.NewReader(
				`repo arepo1
								RW+a = user1`)
			_, err := Read(r)
			So(strings.Contains(err.Error(), ": Incorrect access rule"), ShouldBeTrue)
		})

		Convey("Access rule must be well formed: data alphanum only", func() {
			r := strings.NewReader(
				`repo arepo1
								RW+ a,b = user1`)
			_, err := Read(r)
			So(strings.Contains(err.Error(), ": Incorrect access rule"), ShouldBeTrue)
		})

		Convey("Access rule can have a param", func() {
			r := strings.NewReader(
				`repo arepo1
								RW+ master = user1`)
			gtl, err := Read(r)
			So(err, ShouldBeNil)
			rules, err := gtl.Rules("arepo1")
			So(err, ShouldBeNil)
			So(len(rules), ShouldEqual, 1)
			So(rules[0].String(), ShouldEqual, "RW+ master = user1")
		})
		Convey("Access rule can reference a group of repos", func() {
			r := strings.NewReader(
				`@grp1 = rep1 rep2
						@grp2 = rep1 rep3
						@usr1 = user11
						@usr2 = user12
						repo @grp1
							RW+ master = user11
						repo @grp2
							RW+ dev = user12`)
			gtl, err := Read(r)
			So(err, ShouldBeNil)
			rules, err := gtl.Rules("rep1")
			So(err, ShouldBeNil)
			So(len(rules), ShouldEqual, 2)
			So(rules[0].String(), ShouldEqual, "RW+ master = user11")
			So(gtl.String(), ShouldEqual, `NbGroups: 4 [@grp1, @grp2, @usr1, @usr2]
NbRepoOrGroups: 5 [@grp1, rep1, rep2, @grp2, rep3]
NbUserOrGroups: 4 [user11, @usr1, user12, @usr2]
NbConfigs: 2 [config [group '@grp1'<repos>: [rep1 rep2]] => rules [RW+ master = user11], config [group '@grp2'<repos>: [rep1 rep3]] => rules [RW+ dev = user12]]
`)
		})

		Convey("Access rules can reference a group of repos", func() {
			r := strings.NewReader(
				`@grp1 = rep1 rep2
						repo @grp1
							RW+ dev = user21
							RW+ master = user11`)
			gtl, err := Read(r)
			So(err, ShouldBeNil)
			rules, err := gtl.Rules("rep1")
			So(err, ShouldBeNil)
			So(len(rules), ShouldEqual, 2)
			So(rules[0].String(), ShouldEqual, "RW+ dev = user21")
		})

		Convey("Access rules can reference a group of repos in param", func() {
			r := strings.NewReader(
				`@grp1 = rep1 rep2
						@grp3 = user3 user4
						repo @grp1
							RW+ dev = @grp2
							RW+ master = user11 @grp3`)
			gtl, err := Read(r)
			So(err, ShouldBeNil)
			rules, err := gtl.Rules("rep1")
			So(err, ShouldBeNil)
			So(len(rules), ShouldEqual, 2)
			So(rules[0].String(), ShouldEqual, "RW+ dev = @grp2")
			So(len(gtl.GetUsersOrGroups()), ShouldEqual, 5)
			So(rules[1].String(), ShouldEqual, "RW+ master = user11, @grp3 (user3, user4)")
		})

		Convey("undefined repo group", func() {
			r := strings.NewReader(
				`repo @grp1
								RW+ = user1`)
			gtl, err := Read(r)
			So(err, ShouldNotBeNil)
			So(strings.Contains(err.Error(), ": repo group name"), ShouldBeTrue)
			So(gtl.IsEmpty(), ShouldBeTrue)
			So(gtl.NbUsers(), ShouldEqual, 0)
		})
		Convey("undefined repo used in group", func() {
			r := strings.NewReader(
				`@grp1 = rep1
						repo @grp1 rep2
								RW+ = user1`)
			gtl, err := Read(r)
			So(err, ShouldBeNil)
			So(gtl.IsEmpty(), ShouldBeFalse)
			So(gtl.NbRepos(), ShouldEqual, 2)
		})

		Convey("Invalid user group named after repo group", func() {
			r := strings.NewReader(
				`@grp1 = rep1
						repo @grp1 rep2
								RW+ = user1 @grp1`)
			gtl, err := Read(r)
			So(err, ShouldNotBeNil)
			So(strings.Contains(err.Error(), "user or user group name '@grp1' already used in a repo group"), ShouldBeTrue)
			So(gtl.IsEmpty(), ShouldBeFalse)
			So(gtl.NbRepos(), ShouldEqual, 2)
		})

		Convey("Users are detected", func() {
			var gitoliteconf = `
				@project = module1 module2
				@almadmins = alm1 alm2

				repo gitolite-admin
				   RW+     =   gitoliteadm @almadmins
				   RW                                = projectowner
				   RW VREF/NAME/conf/subs/project    = projectowner
				   -  VREF/NAME/                     = projectowner
				   RW                                = projectowner2
				   RW VREF/NAME/conf/subs/project2   = projectowner2
				   -  VREF/NAME/                     = projectowner2

				 repo module1
				   RW+ = projectowner @almadmins
		`
			r := strings.NewReader(gitoliteconf)
			gtl, err := Read(r)
			So(err, ShouldBeNil)
			So(gtl.IsEmpty(), ShouldBeFalse)
			So(gtl.NbRepos(), ShouldEqual, 3)
			So(gtl.NbRepoGroups(), ShouldEqual, 2)
			So(gtl.NbUsers(), ShouldEqual, 5)
			So(fmt.Sprintf("%v", gtl.GetUsersOrGroups()), ShouldEqual, "[user 'gitoliteadm' group '@almadmins'<users>: [alm1 alm2] user 'alm1' user 'alm2' user 'projectowner' user 'projectowner2']")
			almadmin := gtl.GetGroup("@almadmins")
			So(almadmin, ShouldNotBeNil)
			So(fmt.Sprintf("%v", almadmin.String()), ShouldEqual, "group '@almadmins'<users>: [alm1 alm2]")
		})

		Convey("subconf with a repo group should not be a user group", func() {
			var gitoliteconf = `
				@project = module1 module2

				repo arepo
				   RW+     =   gitoliteadm @project
				   RW                                = projectowner
				   RW VREF/NAME/conf/subs/project    = projectowner
				   -  VREF/NAME/                     = projectowner
		`
			r := strings.NewReader(gitoliteconf)
			gtl, err := Read(r)
			So(err, ShouldNotBeNil)
			So(strings.Contains(err.Error(), "group '@project' is a users group, not a repo one"), ShouldBeTrue)
			So(gtl, ShouldNotBeNil)
		})

	})

	Convey("An reader can read repo and users", t, func() {
		Convey("A repo name shouldn't be part of a users group", func() {
			r := strings.NewReader(
				`@ausergrp = user1 user2 user3
							 repo arepo1
								RW+ = user1
							 repo user1`)
			gtl, err := Read(r)
			So(err, ShouldNotBeNil)
			So(strings.Contains(err.Error(), "already used in a user group"), ShouldBeTrue)
			So(gtl.IsEmpty(), ShouldBeFalse)
			So(gtl.NbUsers(), ShouldEqual, 3)
			So(gtl.NbUserGroups(), ShouldEqual, 1)
		})
		Convey("A user name shouldn't be part of a repo group", func() {
			r := strings.NewReader(
				`@arepogrp = repo1 repo2 repo3
							 repo repo1
								RW+ = repo2`)
			gtl, err := Read(r)
			So(err, ShouldNotBeNil)
			So(strings.Contains(err.Error(), "already used in a repo group"), ShouldBeTrue)
			So(gtl.IsEmpty(), ShouldBeFalse)
			So(gtl.NbUsers(), ShouldEqual, 0)
		})

	})

	Convey("An reader can get configs", t, func() {
		test = ""

		Convey("A gitolite.conf must have at least a gitolite-admin config", func() {
			r := strings.NewReader(
				`@ausergrp = user1 user2 user3`)
			gtl, err := Read(r)
			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldEqual, "There must be one and only gitolite-admin repo config")
			So(gtl.IsEmpty(), ShouldBeFalse)
			So(gtl.NbUsers(), ShouldEqual, 0)
			So(gtl.NbUserGroups(), ShouldEqual, 0)
		})

		Convey("A gitolite-admin config must have at least one rule", func() {
			r := strings.NewReader(
				`repo gitolite-admin`)
			gtl, err := Read(r)
			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldEqual, "There must be at least one rule for gitolite-admin repo config")
			So(gtl.IsEmpty(), ShouldBeFalse)
			So(gtl.NbConfigs(), ShouldEqual, 1)
		})

		Convey("A gitolite-admin must have one RW+ rule", func() {
			r := strings.NewReader(
				`repo gitolite-admin
							RW = user1`)
			gtl, err := Read(r)
			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldContainSubstring, "First rule for gitolite-admin repo config must be 'RW+', empty param")
			So(gtl.IsEmpty(), ShouldBeFalse)
			So(gtl.NbConfigs(), ShouldEqual, 1)
		})

		Convey("A gitolite-admin must have one RW+ rule with no param", func() {
			r := strings.NewReader(
				`repo gitolite-admin
							RW+ param= user2`)
			gtl, err := Read(r)
			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldContainSubstring, "First rule for gitolite-admin repo config must be 'RW+', empty param")
			So(gtl.IsEmpty(), ShouldBeFalse)
			So(gtl.NbConfigs(), ShouldEqual, 1)
		})

		Convey("A gitolite-admin must have one RW+ rule with at least one user", func() {
			r := strings.NewReader(
				`repo gitolite-admin
							RW+ = @users`)
			gtl, err := Read(r)
			// fmt.Println(gtl.String())
			// Now the group is mark as user group
			So(err, ShouldBeNil)
			//So(err.Error(), ShouldEqual, "First rule for gitolite-admin repo must have at least one user or group of users")
			So(gtl.IsEmpty(), ShouldBeFalse)
			So(gtl.NbConfigs(), ShouldEqual, 1)
		})

	})

	Convey("An reader can get configs with description", t, func() {
		test = ""

		Convey("A Config can have a description", func() {
			r := strings.NewReader(
				`repo gitolite-admin
							desc = test  d
							RW+ = user1`)
			gtl, err := Read(r)
			So(err, ShouldBeNil)
			So(gtl.IsEmpty(), ShouldBeFalse)
			So(gtl.GetConfigsForRepo("gitolite-admin")[0].Desc(), ShouldEqual, "test  d")
		})
		Convey("A Config can should have no more than one description", func() {
			r := strings.NewReader(
				`repo gitolite-admin
							desc = test  d
							RW+ = user1
							desc = 2`)
			gtl, err := Read(r)
			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldEqual, "Parse Error: No more than one desc per config, line 4 ('desc = 2')")
			So(gtl.IsEmpty(), ShouldBeFalse)
			So(gtl.GetConfigsForRepo("gitolite-admin")[0].Desc(), ShouldEqual, "test  d")
		})
	})

	Convey("An reader can get comments and empty lines", t, func() {
		test = "ignorega"

		Convey("A reader can get comments before a Group", func() {
			r := strings.NewReader(
				`
						  #  a   comment

						@grpusers = user1 user2`)
			gtl, err := Read(r)
			So(err, ShouldBeNil)
			So(gtl.IsEmpty(), ShouldBeFalse)
			So(gtl.GetGroup("@grpusers").Comment().String(), ShouldEqual, `
#  a   comment

`)
		})

		Convey("A reader can get comments before a Config", func() {
			r := strings.NewReader(
				`#  a group  comment
						@grpusers = user1 user2

						# config comment
						repo r1`)
			gtl, err := Read(r)
			So(err, ShouldBeNil)
			So(gtl.IsEmpty(), ShouldBeFalse)
			So(gtl.GetGroup("@grpusers").Comment().String(), ShouldEqual, `#  a group  comment
`)
			So(gtl.GetConfigsForRepo("r1")[0].Comment().String(), ShouldEqual, `
# config comment
`)
		})

		Convey("A reader can get comments before a Rule", func() {
			r := strings.NewReader(
				`#  a group  comment
						@grpusers = user1 user2

						# config comment
						repo r1
						 #   main admins
						RW+     =   gitoliteadm @almadmins`)
			gtl, err := Read(r)
			So(err, ShouldBeNil)
			So(gtl.IsEmpty(), ShouldBeFalse)
			So(gtl.GetGroup("@grpusers").Comment().String(), ShouldEqual, `#  a group  comment
`)
			So(gtl.GetConfigsForRepo("r1")[0].Rules()[0].Comment().String(), ShouldEqual, `#   main admins
`)
		})

		Convey("A reader can get comment in the same line as a repo rule", func() {
			r := strings.NewReader(
				`repo arepo1
								RW+ = user1  # com d ev`)
			gtl, err := Read(r)
			So(err, ShouldBeNil)
			So(gtl.IsEmpty(), ShouldBeFalse)
			cmt := gtl.GetConfigsForRepo("arepo1")[0].Rules()[0].Comment()
			So(cmt.SameLine(), ShouldEqual, "# com d ev")
		})

	})

	Convey("A Gitolite print itself", t, func() {
		test = "ignorega"

		Convey("A Gitolite can print a single group with content", func() {
			r := strings.NewReader(`
						# comment

						@developers3  =   dilbert alice  wally3`)
			gtl, err := Read(r)
			So(err, ShouldBeNil)
			So(gtl.Print(), ShouldEqual, `
# comment

@developers3 = dilbert alice wally3

`)
		})

		Convey("A Gitolite can print a single config with content", func() {
			r := strings.NewReader(
				`
						  # config comment
						repo r1
						 #   main admins
						RW+     =   gitoliteadm @almadmins`)
			gtl, err := Read(r)
			So(err, ShouldBeNil)
			So(gtl.IsEmpty(), ShouldBeFalse)
			So(gtl.Print(), ShouldEqual, `
# config comment
repo r1
    #   main admins
    RW+   = gitoliteadm @almadmins

`)
		})

		Convey("A Gitolite can print a configs including gitolite-admin config", func() {
			r := strings.NewReader(
				`  # ga comment
					repo gitolite-admin
					RW+ = admin
					repo otherRepo
					R  param  = user`)
			gtl, err := Read(r)
			So(err, ShouldBeNil)
			So(gtl.IsEmpty(), ShouldBeFalse)
			So(gtl.Print(), ShouldEqual, `# ga comment
repo gitolite-admin
    RW+   = admin

repo otherRepo
    R    param = user

`)
		})

		Convey("A Gitolite can print a configs with desc, incliding group", func() {
			r := strings.NewReader(
				`# group users
						@users = u1 u2
						# ga comment
					repo gitolite-admin
					RW+ = admin
					repo otherRepo
					# rule comment
					R  param  = user @users
					# desc comment
					desc = a desc`)
			gtl, err := Read(r)
			So(err, ShouldBeNil)
			So(gtl.IsEmpty(), ShouldBeFalse)
			res := `# group users
@users = u1 u2

# ga comment
repo gitolite-admin
    RW+   = admin

repo otherRepo
    # desc comment
    desc  = a desc
    # rule comment
    R    param = user @users

`
			So(gtl.Print(), ShouldEqual, res)
			config := gtl.GetConfigsForRepo("otherRepo")[0]
			rule := config.Rules()[0]
			So(fmt.Sprintf("%v", rule.String()), ShouldEqual, "R param = user, @users (u1, u2)")
			So(len(rule.GetUsersOrGroups()), ShouldEqual, 2)
		})

	})

	Convey("A Gitolite can read subconfs", t, func() {
		test = "ignorega"

		Convey("A Gitolite detect invalid subconf line", func() {
			r := strings.NewReader(`
						# comment
						subconf subs/*.conf
`)
			gtl, err := Read(r)
			So(err, ShouldNotBeNil)
			So(strings.HasPrefix(err.Error(), "Parse Error: Invalid subconf at line"), ShouldBeTrue)
			So(len(gtl.Subconfs()), ShouldEqual, 0)
		})

		Convey("A Gitolite detect invalid subconf line~regexp", func() {
			r := strings.NewReader(`
						# comment
						subconf "subs/(*.conf"
`)
			gtl, err := Read(r)
			So(err, ShouldNotBeNil)
			So(strings.HasPrefix(err.Error(), "Parse Error: Invalid subconf regexp"), ShouldBeTrue)
			So(len(gtl.Subconfs()), ShouldEqual, 0)
		})

		Convey("A Gitolite detect invalid subconf line", func() {
			r := strings.NewReader(`
						# comment
						subconf "subs/*.conf"
						subconf "subs2/*.conf"
						subconf "subs/*.conf"
`)
			gtl, _ := Read(r)
			So(len(gtl.Subconfs()), ShouldEqual, 2)
		})

	})

}
