package gogitolite

import (
	"strings"
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func TestRead(t *testing.T) {

	Convey("An empty reader means no repos", t, func() {
		gtl := Read(nil)
		So(gtl.IsEmpty(), ShouldBeTrue)
		r := strings.NewReader("")
		gtl = Read(r)
		So(gtl.IsEmpty(), ShouldBeTrue)
	})

	Convey("An reader can read groups", t, func() {
		r := strings.NewReader("  @developers     =   dilbert alice wally  ")
		gtl := Read(r)
		So(gtl.IsEmpty(), ShouldBeFalse)
	})

}
