package main

import (
	"bufio"
	"bytes"
	"io/ioutil"
	"os"
	"strings"
	"testing"
	. "github.com/smartystreets/goconvey/convey"
)

var (
	gitoliteconf_bad string
	gitoliteconf     string
	bout             *bytes.Buffer
	wout             *bufio.Writer
	berr             *bytes.Buffer
	werr             *bufio.Writer
)

func init() {
	gitoliteconf_bad = `test`
	gitoliteconf = `
		@project = module1 module2
		@almadmins = admin1 admin2

		repo gitolite-admin
	      RW+     =   gitoliteadm @almadmins @alm2
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
  RW = user3 HBu1
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
	bout = bytes.NewBuffer(nil)
	wout = bufio.NewWriter(bout)
	out()
	oerr()
	sout = wout
	berr = bytes.NewBuffer(nil)
	werr = bufio.NewWriter(berr)
	serr = werr
}

func resetStds() {
	bout = bytes.NewBuffer(nil)
	sout.Reset(bout)
	berr = bytes.NewBuffer(nil)
	serr.Reset(berr)
}

func flushStds() {
	sout.Flush()
	serr.Flush()
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

		Convey("Default usage", func() {
			args = []string{"-h"}
			main()
			flushStds()
			So(r, ShouldBeNil)
			So(bout.String(), ShouldEqual, "")
			So(berr.String(), ShouldEqual, `Usage: gogitolite.exe [opts] gitolite.conf
Options:
  -audit=false: print user access audit
  -list=false: list projects
  -print=false: print config
  -v=false: verbose, display filenames read
`)
			resetStds()
		})
		Convey("Error if no file", func() {
			args = []string{"-v", "-audit"}
			main()
			flushStds()
			So(r, ShouldBeNil)
			So(bout.String(), ShouldEqual, "")
			So(berr.String(), ShouldEqual, "One gitolite.conf file expected")
			resetStds()
		})
		Convey("Error if unknown file", func() {

			args = []string{"-audit", "unknownFile"}
			main()
			flushStds()
			So(r.gtl, ShouldBeNil)
			So(bout.String(), ShouldEqual, `Read file 'unknownFile'
`)
			So(berr.String(), ShouldEqual, `ERR open unknownFile: The system cannot find the file specified.
`)
			resetStds()
		})
		Convey("Error if bad gitolite-admin config file content", func() {

			r := strings.NewReader(gitoliteconf_bad)
			gtl, err := getGtl2(r, nil)
			flushStds()
			So(err, ShouldNotBeNil)
			So(gtl, ShouldBeNil)
			So(bout.String(), ShouldEqual, ``)
			So(berr.String(), ShouldEqual, `ERR Parse Error: group or repo expected after line 1 ('test')
`)
			resetStds()
		})

		Convey("Detects one project with several admins and users", func() {

			args = []string{"-v", "-audit", "_tests/p1/conf/gitolite.conf"}
			main()
			So(r, ShouldNotBeNil)
			gtl, err := getGtlFromFile("_tests/p1/conf/gitolite.conf", r.gtl)
			flushStds()
			So(err, ShouldBeNil)
			So(gtl, ShouldNotBeNil)
			So(bout.String(), ShouldEqual, `Read file '_tests/p1/conf/gitolite.conf'
Visited: subs/project.conf _tests\p1\conf\subs\project.conf
Visited: subs/projectbad.conf _tests\p1\conf\subs\projectbad.conf
@alm2,,gitolite-admin,system
HBu1,,@project,system
HBu1,,module1,system
HBu1,,module2,system
admin1,,gitolite-admin,system
admin2,,gitolite-admin,system
gitoliteadm,,gitolite-admin,user
projectowner1,,gitolite-admin,system
projectowner2,,gitolite-admin,system
pu1,,@project,user
pu1,,module1,user
pu1,,module2,user
user1,,module1,user
user11,,module1,user
user2,,module1,user
user2,,module2,user
user21,,module2,user
user3,,@project,user
user3,,module1,user
user3,,module2,user
`)
			So(berr.String(), ShouldEqual, `ERR Parse Error: group or repo expected after line 2 ('repo')
Ignore subconf file: subs/projectbad.conf _tests\p1\conf\subs\projectbad.conf because of err 'Parse Error: group or repo expected after line 2 ('repo')'
`)
			resetStds()
		})

		Convey("List one project with several admins and users", func() {

			args = []string{"-v", "-list", "_tests/p1/conf/gitolite.conf"}
			main()
			flushStds()
			So(r, ShouldNotBeNil)
			So(bout.String(), ShouldEqual, `Read file '_tests/p1/conf/gitolite.conf'
Visited: subs/project.conf _tests\p1\conf\subs\project.conf
Visited: subs/projectbad.conf _tests\p1\conf\subs\projectbad.conf
@alm2,,gitolite-admin,system
HBu1,,@project,system
HBu1,,module1,system
HBu1,,module2,system
admin1,,gitolite-admin,system
admin2,,gitolite-admin,system
gitoliteadm,,gitolite-admin,user
projectowner1,,gitolite-admin,system
projectowner2,,gitolite-admin,system
pu1,,@project,user
pu1,,module1,user
pu1,,module2,user
user1,,module1,user
user11,,module1,user
user2,,module1,user
user2,,module2,user
user21,,module2,user
user3,,@project,user
user3,,module1,user
user3,,module2,user
NbProjects: 1
project project, admins: projectowner1, projectowner2, members: user1, user11, user2, pu1, user21
`)
			So(berr.String(), ShouldEqual, `ERR Parse Error: group or repo expected after line 2 ('repo')
Ignore subconf file: subs/projectbad.conf _tests\p1\conf\subs\projectbad.conf because of err 'Parse Error: group or repo expected after line 2 ('repo')'
`)
			resetStds()
		})
	})

	Convey("Prints configs", t, func() {
		Convey("Print a gitolite config", func() {

			args = []string{"-print", "_tests/p1/conf/gitolite.conf"}
			main()
			So(r, ShouldNotBeNil)
			gtl, err := getGtlFromFile("_tests/p1/conf/gitolite.conf", r.gtl)
			flushStds()
			So(err, ShouldBeNil)
			So(gtl, ShouldNotBeNil)
			So(bout.String(), ShouldEqual, `Read file '_tests/p1/conf/gitolite.conf'
Visited: subs/project.conf _tests\p1\conf\subs\project.conf
Visited: subs/projectbad.conf _tests\p1\conf\subs\projectbad.conf
@alm2,,gitolite-admin,system
HBu1,,@project,system
HBu1,,module1,system
HBu1,,module2,system
admin1,,gitolite-admin,system
admin2,,gitolite-admin,system
gitoliteadm,,gitolite-admin,user
projectowner1,,gitolite-admin,system
projectowner2,,gitolite-admin,system
pu1,,@project,user
pu1,,module1,user
pu1,,module2,user
user1,,module1,user
user11,,module1,user
user2,,module1,user
user2,,module2,user
user21,,module2,user
user3,,@project,user
user3,,module1,user
user3,,module2,user
NbProjects: 1
project project, admins: projectowner1, projectowner2, members: user1, user11, user2, pu1, user21


@project = module1 module2

@almadmins = admin1 admin2


repo gitolite-admin
        RW+                              = gitoliteadm @almadmins @alm2
        RW                               = projectowner1 projectowner2
        RW   VREF/NAME/conf/subs/project = projectowner1 projectowner2
        -    VREF/NAME/                  = projectowner1 projectowner2



@project = module1 module2


repo module1
        RW    = user1 user11 user2

repo module2
        RW    = user2 user21

repo @project
        RW    = pu1

`)
			So(berr.String(), ShouldEqual, `ERR Parse Error: group or repo expected after line 2 ('repo')
Ignore subconf file: subs/projectbad.conf _tests\p1\conf\subs\projectbad.conf because of err 'Parse Error: group or repo expected after line 2 ('repo')'
`)
			resetStds()
		})
	})
}
