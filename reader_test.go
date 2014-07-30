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
			r := strings.NewReader("  @developers     =   dilbert alice wally  \nrepo ")
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
	})

	Convey("An group name must be [a-zA-Z0-9_-]", t, func() {
		r := strings.NewReader("  @develop;ers     =   dilbert alice wally")
		gtl, err := Read(r)
		So(gtl.IsEmpty(), ShouldBeTrue)
		So(strings.Contains(err.Error(), ": Incorrect repo declaration"), ShouldBeTrue)
	})
}
