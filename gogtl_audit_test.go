package main

import (
	"io/ioutil"
	"os"
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

	Convey("Detects projects", t, func() {

		Convey("Detects one project with several admins and users", func() {
			var gitoliteconf = `
		@project = module1 module2

		repo gitolite-admin
	      RW+     =   gitoliteadm @almadmins
	      RW                                = projectowner1 projectowner2
	      RW VREF/NAME/conf/subs/project    = projectowner1 projectowner2
	      -  VREF/NAME/                     = projectowner1 projectowner2

	    repo module1
	      RW = user1 user11 user2
	    repo module2
	      RW = user2 user21
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

			main("-v", "-audit", "_tests/p1/conf/gitolite.conf")
		})
	})
}
