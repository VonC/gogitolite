package gogitolite

import (
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
			So(gtl.NbGroupRepos(), ShouldEqual, 1)
		})
	})

	Convey("An reader can read repo rules", t, func() {
		Convey("single rule", func() {
			r := strings.NewReader(
				`repo arepo1
				   RW+ = user1`)
			gtl, err := Read(r)
			So(err, ShouldBeNil)
			So(gtl.IsEmpty(), ShouldBeFalse)
			So(gtl.NbUsers(), ShouldEqual, 1)
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
			So(rules[0].String(), ShouldEqual, "RW+ master user1")
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
			So(strings.Contains(err.Error(), "already used user group at line"), ShouldBeTrue)
			So(gtl.IsEmpty(), ShouldBeFalse)
			So(gtl.NbUsers(), ShouldEqual, 1)
			So(gtl.NbGroupUsers(), ShouldEqual, 1)
		})
		Convey("A user name shouldn't be part of a repo group", func() {
			r := strings.NewReader(
				`@arepogrp = repo1 repo2 repo3
				 repo repo1
				   RW+ = repo2`)
			gtl, err := Read(r)
			So(err, ShouldNotBeNil)
			So(strings.Contains(err.Error(), "already used repo group at line"), ShouldBeTrue)
			So(gtl.IsEmpty(), ShouldBeFalse)
			So(gtl.NbUsers(), ShouldEqual, 1)
		})

	})

	/*
			@project = module1 module2

			repo gitolite-admin
		      RW+     =   gitoliteadm @almadmins
		      RW                                = projectowner
		      RW VREF/NAME/conf/subs/project    = projectowner
		      -  VREF/NAME/                     = projectowner

		    repo module1
		      desc = module1 repo for repos group project1
		      RW+ = projectowner @almadmins

		    subconf "subs/*.conf"

		    more subs/project.conf

		    repo @project
		      RW = projectowner @almadmins
		      RW = otheruser1 otheruser2...

	*/
}
