package reader

import (
	"bufio"
	"fmt"
	"io"
	"regexp"
	"strings"

	"github.com/VonC/gogitolite/gitolite"
)

type content struct {
	s             *bufio.Scanner
	l             int
	gtl           *gitolite.Gitolite
	currentConfig *gitolite.Config
}

type stateFn func(*content) (stateFn, error)

var test = ""
var currentComment = &gitolite.Comment{}

// Read a gitolite config file
func Read(r io.Reader) (*gitolite.Gitolite, error) {
	return Update(r, nil)
}

// Update a gitolite config file
func Update(r io.Reader, gtl *gitolite.Gitolite) (*gitolite.Gitolite, error) {
	res := gitolite.NewGitolite(gtl)
	if r == nil {
		return res, nil
	}
	s := bufio.NewScanner(r)
	s.Scan()
	c := &content{s: s, gtl: res, l: 1}
	var state stateFn
	var err error
	for state, err = readEmptyOrCommentLines(c); state != nil && err == nil; {
		state, err = state(c)
	}
	if err == nil && test != "ignorega" && gtl == nil {
		configs := res.GetConfigsForRepo("gitolite-admin")
		err = checkConfigRead(configs)
		/*
			if !rule.HasAnyUserOrGroup() {
				err = fmt.Errorf("First rule for gitolite-admin repo must have at least one user or group of users")
			}
		*/
	}
	//fmt.Printf("\nGitolite res='%v'\n", res)
	return res, err
}

func checkConfigRead(configs []*gitolite.Config) error {
	var err error
	if len(configs) != 1 {
		err = fmt.Errorf("There must be one and only gitolite-admin repo config")
		return err
	}
	config := configs[0]
	if len(config.Rules()) == 0 {
		err = fmt.Errorf("There must be at least one rule for gitolite-admin repo config")
		return err
	}
	rule := config.Rules()[0]
	if rule.Access() != "RW+" || rule.Param() != "" {
		err = fmt.Errorf("First rule for gitolite-admin repo config must be 'RW+', empty param, instead of '%v'-'%v'", rule.Access(), rule.Param())
		return err
	}
	return nil
}

// ParseError indicates gitolite.conf parsing error
type ParseError struct {
	msg string
}

func (pe ParseError) Error() string {
	return fmt.Sprintf("Parse Error: %s", pe.msg)
}

var readEmptyOrCommentLinesRx = regexp.MustCompile(`(?m)^\s*?$|^\s*?#(.*?)$`)
var readSubconfLinesRx = regexp.MustCompile(`(?m)^\s*?subconf\s+"(.*.conf)"\s*?$`)

func readEmptyOrCommentLines(c *content) (stateFn, error) {
	t := c.s.Text()
	for keepReading := true; keepReading; {
		res := readEmptyOrCommentLinesRx.FindStringSubmatchIndex(t)
		//fmt.Println(res, ">'"+t+"'")
		if res == nil {
			res := readSubconfLinesRx.FindStringSubmatchIndex(t)
			if res == nil {
				if strings.HasPrefix(strings.TrimSpace(t), "subconf") {
					return nil, ParseError{msg: fmt.Sprintf("Invalid subconf at line %v ('%v')", c.l, t)}
				}
				return readRepoOrGroup, nil
			}
			err := c.gtl.AddSubconf(t[res[2]:res[3]])
			if err != nil {
				return nil, ParseError{msg: fmt.Sprintf("Invalid subconf regexp:\n%v at line %v ('%v')", err.Error(), c.l, t)}
			}
		} else {
			currentComment.AddComment(t)
			//fmt.Println("\nCMT: ", currentComment, "\nGTL: ", c.gtl)
		}
		if !c.s.Scan() {
			keepReading = false
		} else {
			c.l = c.l + 1
			t = c.s.Text()
		}
	}
	if c.gtl.IsEmpty() {
		return nil, ParseError{msg: fmt.Sprintf("comment, group or repo expected at line %v ('%v')", c.l, t)}
	}
	return nil, nil
}

var readRepoOrGroupRx = regexp.MustCompile(`^\s*?(repo |@)`)

func readRepoOrGroup(c *content) (stateFn, error) {
	t := strings.TrimSpace(c.s.Text())
	res := readRepoOrGroupRx.FindStringSubmatchIndex(t)
	if res == nil {
		return nil, ParseError{msg: fmt.Sprintf("group or repo expected after line %v ('%v')", c.l, t)}
	}
	prefix := t[res[2]:res[3]]
	if prefix == "@" {
		return readGroup, nil
	}
	return readRepo, nil
}

var readGroupRx = regexp.MustCompile(`(?m)^\s*?(@[a-zA-Z0-9_-]+)\s*?=\s*?((?:@?[a-zA-Z0-9\._-]+\s*?)+)$`)

func readGroup(c *content) (stateFn, error) {
	t := strings.TrimSpace(c.s.Text())
	res := readGroupRx.FindStringSubmatchIndex(t)
	//fmt.Println(res, "'"+t+"'")
	if len(res) == 0 {
		return nil, ParseError{msg: fmt.Sprintf("Incorrect group declaration at line %v ('%v')", c.l, t)}
	}
	//fmt.Println(res, "'"+c.s+"'", "'"+c.s[res[2]:res[3]]+"'", "'"+c.s[res[4]:res[5]]+"'")
	grpname := t[res[2]:res[3]]
	grpmembers := strings.Split(strings.TrimSpace(t[res[4]:res[5]]), " ")
	// http://cats.groups.google.com.meowbify.com/forum/#!topic/golang-nuts/-pqkICuokio
	//fmt.Printf("'%v'\n", grpmembers)
	if err := c.gtl.AddUserOrRepoGroup(grpname, grpmembers, currentComment); err != nil {
		return nil, ParseError{msg: fmt.Sprintf("%v at line %v ('%v')", err.Error(), c.l, t)}
	}
	currentComment = &gitolite.Comment{}

	// fmt.Println("'" + c.s + "'")
	if !c.s.Scan() {
		return nil, nil
	}
	c.l = c.l + 1
	return readEmptyOrCommentLines, nil
}

var readRepoRx = regexp.MustCompile(`(?m)^\s*?repo\s*?((?:@?[a-zA-Z0-9\._-]+\s*?)+)$`)

func readRepo(c *content) (stateFn, error) {
	t := strings.TrimSpace(c.s.Text())
	//fmt.Println(res, "'"+t+"'")
	res := readRepoRx.FindStringSubmatchIndex(t)
	if len(res) == 0 {
		return nil, ParseError{msg: fmt.Sprintf("Incorrect repo declaration at line %v ('%v')", c.l, t)}
	}
	rpmembers := strings.Split(strings.TrimSpace(t[res[2]:res[3]]), " ")
	seen := map[string]bool{}
	for _, val := range rpmembers {
		if _, ok := seen[val]; !ok {
			seen[val] = true
		} else {
			return nil, ParseError{msg: fmt.Sprintf("Duplicate repo element name '%v' at line %v ('%v')", val, c.l, t)}
		}
	}
	var config *gitolite.Config
	if cfg, err := c.gtl.AddConfig(rpmembers, currentComment); err == nil {
		config = cfg
	} else {
		return nil, ParseError{msg: fmt.Sprintf("%v\nAt line %v ('%v')", err.Error(), c.l, t)}
	}
	currentComment = &gitolite.Comment{}

	if !c.s.Scan() {
		return nil, nil
	}
	c.l = c.l + 1
	c.currentConfig = config
	return readRepoRules, nil
}

var readRepoRuleRx = regexp.MustCompile(`(?m)^\s*?([^@=]+)\s*?=\s*?((?:@?[a-zA-Z0-9_.-]+\s*?)+)(#.*?)?$`)
var repoRulePreRx = regexp.MustCompile(`(?m)^([RW+-]+?)\s*?(?:\s([a-zA-Z0-9_.\-/]+))?$`)
var repoRuleDescRx = regexp.MustCompile(`(?m)^desc\s*?=\s*?(\S.*?)$`)

func readRepoRulesDesc(c *content, config *gitolite.Config, t string) (bool, error) {
	res := repoRuleDescRx.FindStringSubmatchIndex(t)
	//fmt.Println(res, ">0'"+t+"'")
	if res == nil || len(res) == 0 {
		return false, nil
	}
	if err := config.SetDesc(strings.TrimSpace(t[res[2]:res[3]]), currentComment); err != nil {
		return true, ParseError{msg: fmt.Sprintf("%v, line %v ('%v')", err.Error(), c.l, t)}
	}
	currentComment = &gitolite.Comment{}
	return true, nil
}

func readRepoRulesComment(t string) (bool, error) {
	res := readEmptyOrCommentLinesRx.FindStringSubmatchIndex(t)
	if res == nil || len(res) == 0 {
		return false, nil
	}
	currentComment.AddComment(t)
	return true, nil
}

func readRepoRuleGroupUsers(rule *gitolite.Rule, username string, c *content, t string) error {
	if err := c.gtl.AddUserOrGroupToRule(rule, username); err != nil {
		return ParseError{msg: fmt.Sprintf("%v\nAt line %v (%v)", err.Error(), c.l, t)}
	}
	return nil
}

func readRepoRuleUsers(rule *gitolite.Rule, post string, c *content, t string) error {
	users := strings.Split(post, " ")
	for _, username := range users {
		if !strings.HasPrefix(username, "@") {
			if err := c.gtl.AddUserOrGroupToRule(rule, username); err != nil {
				return ParseError{msg: fmt.Sprintf("%v\nAt line %v (%v)", err.Error(), c.l, t)}
			}
		} else {
			if err := readRepoRuleGroupUsers(rule, username, c, t); err != nil {
				return err
			}
		}
	}
	return nil
}

func readRepoRule(c *content, config *gitolite.Config, t string) (bool, error) {
	res := readRepoRuleRx.FindStringSubmatchIndex(t)
	if res == nil || len(res) == 0 {
		return false, nil
	}
	pre := strings.TrimSpace(t[res[2]:res[3]])
	post := strings.TrimSpace(t[res[4]:res[5]])
	if res[6] > -1 {
		//fmt.Printf("\nreadRepoRuleRx res='%v'\n", res)
		currentComment.SetSameLine(strings.TrimSpace(t[res[6]:res[7]]))
	}

	respre := repoRulePreRx.FindStringSubmatchIndex(pre)
	fmt.Printf("\nrespre='%v' for '%v'\n", respre, pre)
	if respre == nil {
		return true, ParseError{msg: fmt.Sprintf("Incorrect access rule '%v' at line %v ('%v')", pre, c.l, t)}
	}
	access := pre[respre[2]:respre[3]]
	param := ""
	if respre[4] > -1 {
		param = pre[respre[4]:respre[5]]
	}
	rule := gitolite.NewRule(access, param, currentComment)
	err := readRepoRuleUsers(rule, post, c, t)
	if err != nil {
		return true, err
	}
	c.gtl.AddRuleToConfig(rule, config)
	currentComment = &gitolite.Comment{}

	if strings.HasPrefix(param, "VREF/NAME/conf/subs/") {
		repogrpname := "@" + param[len("VREF/NAME/conf/subs/"):]
		grp := c.gtl.GetGroup(repogrpname)
		if grp == nil {
			c.gtl.AddUserOrRepoGroup(repogrpname, nil, nil)
			/*
				if err != nil {
					return true, err
				}*/
			grp = c.gtl.GetGroup(repogrpname)
		}
		//fmt.Printf("Group '%v' as repo\n", repogrpname)
		if err = grp.MarkAsRepoGroup(); err != nil {
			return true, err
		}
	}

	return true, nil
}

func readRepoRules(c *content) (stateFn, error) {
	t := strings.TrimSpace(c.s.Text())
	//fmt.Printf("readRepoRules '%v'\n", t)
	//rules := []*gitolite.Rule{}
	config := c.currentConfig
	for keepReading := true; keepReading; {
		//fmt.Printf("readRepoRules '%v'\n", t)
		lineProcessed, err := readRepoRulesDesc(c, config, t)
		if !lineProcessed {
			lineProcessed, err = readRepoRulesComment(t)
		}
		if !lineProcessed {
			lineProcessed, err = readRepoRule(c, config, t)
		}
		if err != nil {
			//fmt.Printf("readRepoRules ERR '%v'\n", err)
			return nil, err
		}
		if !lineProcessed {
			if len(config.Rules()) == 0 {
				return nil, ParseError{msg: fmt.Sprintf("At least one access rule expected at line %v ('%v')", c.l, t)}
			}
			break
		}
		if !c.s.Scan() {
			keepReading = false
			return nil, nil
		}
		c.l = c.l + 1
		t = strings.TrimSpace(c.s.Text())
	}
	return readEmptyOrCommentLines, nil
}
