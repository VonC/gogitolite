package gogitolite

import (
	"strings"
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

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
				`@grp1 rep1 rep2
				 repo  rep1 rep2 rep1`)
			gtl, err := Read(r)
			So(err, ShouldBeNil)
			So(gtl.NbGroupRepos(), ShouldEqual, 1)
		})
	})
}
