package gogitolite

import (
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func TestRead(t *testing.T) {

	Convey("An empty reader means no repos", t, func() {
		gtl := Read(nil)
		So(gtl.IsEmpty(), ShouldBeTrue)
	})
}
