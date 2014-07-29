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
		gtl, _ = Read(r)
		So(gtl.IsEmpty(), ShouldBeTrue)
	})

	Convey("An reader can read groups", t, func() {
		r := strings.NewReader("  @developers     =   dilbert alice wally  \nrepo ")
		gtl, err := Read(r)
		So(err, ShouldBeNil)
		So(gtl.IsEmpty(), ShouldBeFalse)
		So(gtl.NbGroup(), ShouldEqual, 1)
	})

	Convey("An group name must be [a-zA-Z0-9_-]", t, func() {
		r := strings.NewReader("  @develop;ers     =   dilbert alice wally")
		gtl, err := Read(r)
		So(gtl.IsEmpty(), ShouldBeTrue)
		So(err, ShouldNotBeNil)
	})
}
