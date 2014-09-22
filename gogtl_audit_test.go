package main

import (
	"io/ioutil"
	"os"
	"strings"
	"testing"
	. "github.com/smartystreets/goconvey/convey"
)

var gitoliteconf_bad string
var gitoliteconf string

func init() {
	gitoliteconf_bad = `test`
	gitoliteconf = `
		@project = module1 module2
		@almadmins = admin1 admin2

		repo gitolite-admin
	      RW+     =   gitoliteadm @almadmins
	      RW                                = projectowner1 projectowner2
	      RW VREF/NAME/conf/subs/project    = projectowner1 projectowner2
	      -  VREF/NAME/                     = projectowner1 projectowner2

	    repo module1
	      RW = user1 user11 user2
	    repo module2
	      RW = user2 user21
	    repo @project
	      RW = pu1
      subconf "subs/*.conf"
`
	if err := os.MkdirAll("_tests/p1/conf/subs", 0755); err != nil {
		panic(err)
	}
	if err := ioutil.WriteFile("_tests/p1/conf/gitolite.conf", []byte(gitoliteconf), 0644); err != nil {
		panic(err)
	}
	var projectconf = `
repo @project
  RW = user3
`
	if err := ioutil.WriteFile("_tests/p1/conf/subs/project.conf", []byte(projectconf), 0644); err != nil {
		panic(err)
	}
	var projectbadconf = `
repo
  RW = user3bad
`
	if err := ioutil.WriteFile("_tests/p1/conf/subs/projectbad.conf", []byte(projectbadconf), 0644); err != nil {
		panic(err)
	}
}

/*
   subconf "subs/*.conf"

   more subs/project.conf

   repo @project
     RW = projectowner @almadmins
     RW = otheruser1 otheruser2...
*/
func TestProject(t *testing.T) {

	Convey("Detects projects", t, func() {

		Convey("Error if no file", func() {

			args = []string{"-v", "-audit"}
			main()
			So(r, ShouldBeNil)
		})
		Convey("Error if unknown file", func() {

			args = []string{"-audit", "unknownFile"}
			main()
			So(r.gtl, ShouldBeNil)
		})
		Convey("Error if bad gitolite-admin config file content", func() {

			r := strings.NewReader(gitoliteconf_bad)
			gtl, err := getGtl2(r, nil)
			So(err, ShouldNotBeNil)
			So(gtl, ShouldBeNil)
		})

		Convey("Detects one project with several admins and users", func() {

			args = []string{"-v", "-audit", "_tests/p1/conf/gitolite.conf"}
			main()
			So(r, ShouldNotBeNil)
			gtl, err := getGtlFromFile("_tests/p1/conf/gitolite.conf", r.gtl)
			So(err, ShouldBeNil)
			So(gtl, ShouldNotBeNil)
		})

		Convey("List one project with several admins and users", func() {

			args = []string{"-v", "-list", "_tests/p1/conf/gitolite.conf"}
			main()
			So(r, ShouldNotBeNil)
		})
	})
}
