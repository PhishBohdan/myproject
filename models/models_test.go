package models

import (
	"strings"
	"testing"

	"github.com/PuerkitoBio/goquery"
	"github.com/gophish/gophish/config"
	"github.com/jinzhu/gorm"
	"gopkg.in/check.v1"
)

// Hook up gocheck into the "go test" runner.
func Test(t *testing.T) { check.TestingT(t) }

type ModelsSuite struct{}

var _ = check.Suite(&ModelsSuite{})

func (s *ModelsSuite) SetUpSuite(c *check.C) {
	config.Conf.DBName = "sqlite3"
	config.Conf.DBPath = ":memory:"
	config.Conf.MigrationsPath = "../db/db_sqlite3/migrations/"
	err := Setup()
	if err != nil {
		c.Fatalf("Failed creating database: %v", err)
	}
}

func (s *ModelsSuite) TearDownTest(c *check.C) {
	// Clear database tables between each test. If new tables are
	// used in this test suite they will need to be cleaned up here.
	db.Delete(Group{})
	db.Delete(Target{})
	db.Delete(GroupTarget{})
	db.Delete(SMTP{})
	db.Delete(Page{})

	// Reset users table to default state.
	db.Not("id", 1).Delete(User{})
	db.Model(User{}).Update("username", "admin")
}

func (s *ModelsSuite) TestGetUser(c *check.C) {
	u, err := GetUser(1)
	c.Assert(err, check.Equals, nil)
	c.Assert(u.Username, check.Equals, "admin")
}

func (s *ModelsSuite) TestPutUser(c *check.C) {
	u, err := GetUser(1)
	u.Username = "admin_changed"
	err = PutUser(&u)
	c.Assert(err, check.Equals, nil)
	u, err = GetUser(1)
	c.Assert(u.Username, check.Equals, "admin_changed")
}

func (s *ModelsSuite) TestGeneratedAPIKey(c *check.C) {
	u, err := GetUser(1)
	c.Assert(err, check.Equals, nil)
	c.Assert(u.ApiKey, check.Not(check.Equals), "12345678901234567890123456789012")
}

func (s *ModelsSuite) TestPostGroup(c *check.C) {
	g := Group{Name: "Test Group"}
	g.Targets = []Target{Target{Email: "test@example.com"}}
	g.UserId = 1
	err := PostGroup(&g)
	c.Assert(err, check.Equals, nil)
	c.Assert(g.Name, check.Equals, "Test Group")
	c.Assert(g.Targets[0].Email, check.Equals, "test@example.com")
}

func (s *ModelsSuite) TestPostGroupNoName(c *check.C) {
	g := Group{Name: ""}
	g.Targets = []Target{Target{Email: "test@example.com"}}
	g.UserId = 1
	err := PostGroup(&g)
	c.Assert(err, check.Equals, ErrGroupNameNotSpecified)
}

func (s *ModelsSuite) TestPostGroupNoTargets(c *check.C) {
	g := Group{Name: "No Target Group"}
	g.Targets = []Target{}
	g.UserId = 1
	err := PostGroup(&g)
	c.Assert(err, check.Equals, ErrNoTargetsSpecified)
}

func (s *ModelsSuite) TestGetGroups(c *check.C) {
	// Add groups.
	PostGroup(&Group{
		Name:    "Test Group 1",
		Targets: []Target{Target{Email: "test1@example.com"}},
		UserId:  1,
	})
	PostGroup(&Group{
		Name:    "Test Group 2",
		Targets: []Target{Target{Email: "test2@example.com"}},
		UserId:  1,
	})

	// Get groups and test result.
	groups, err := GetGroups(1)
	c.Assert(err, check.Equals, nil)
	c.Assert(len(groups), check.Equals, 2)
	c.Assert(len(groups[0].Targets), check.Equals, 1)
	c.Assert(len(groups[1].Targets), check.Equals, 1)
	c.Assert(groups[0].Name, check.Equals, "Test Group 1")
	c.Assert(groups[1].Name, check.Equals, "Test Group 2")
	c.Assert(groups[0].Targets[0].Email, check.Equals, "test1@example.com")
	c.Assert(groups[1].Targets[0].Email, check.Equals, "test2@example.com")
}

func (s *ModelsSuite) TestGetGroupsNoGroups(c *check.C) {
	groups, err := GetGroups(1)
	c.Assert(err, check.Equals, nil)
	c.Assert(len(groups), check.Equals, 0)
}

func (s *ModelsSuite) TestGetGroup(c *check.C) {
	// Add group.
	PostGroup(&Group{
		Name:    "Test Group",
		Targets: []Target{Target{Email: "test@example.com"}},
		UserId:  1,
	})

	// Get group and test result.
	group, err := GetGroup(1, 1)
	c.Assert(err, check.Equals, nil)
	c.Assert(len(group.Targets), check.Equals, 1)
	c.Assert(group.Name, check.Equals, "Test Group")
	c.Assert(group.Targets[0].Email, check.Equals, "test@example.com")
}

func (s *ModelsSuite) TestGetGroupNoGroups(c *check.C) {
	_, err := GetGroup(1, 1)
	c.Assert(err, check.Equals, gorm.ErrRecordNotFound)
}

func (s *ModelsSuite) TestGetGroupByName(c *check.C) {
	// Add group.
	PostGroup(&Group{
		Name:    "Test Group",
		Targets: []Target{Target{Email: "test@example.com"}},
		UserId:  1,
	})

	// Get group and test result.
	group, err := GetGroupByName("Test Group", 1)
	c.Assert(err, check.Equals, nil)
	c.Assert(len(group.Targets), check.Equals, 1)
	c.Assert(group.Name, check.Equals, "Test Group")
	c.Assert(group.Targets[0].Email, check.Equals, "test@example.com")
}

func (s *ModelsSuite) TestGetGroupByNameNoGroups(c *check.C) {
	_, err := GetGroupByName("Test Group", 1)
	c.Assert(err, check.Equals, gorm.ErrRecordNotFound)
}

func (s *ModelsSuite) TestPutGroup(c *check.C) {
	// Add test group.
	group := Group{Name: "Test Group"}
	group.Targets = []Target{
		Target{Email: "test1@example.com", FirstName: "First", LastName: "Example"},
		Target{Email: "test2@example.com", FirstName: "Second", LastName: "Example"},
	}
	group.UserId = 1
	PostGroup(&group)

	// Update one of group's targets.
	group.Targets[0].FirstName = "Updated"
	err := PutGroup(&group)
	c.Assert(err, check.Equals, nil)

	// Verify updated target information.
	targets, _ := GetTargets(group.Id)
	c.Assert(targets[0].Email, check.Equals, "test1@example.com")
	c.Assert(targets[0].FirstName, check.Equals, "Updated")
	c.Assert(targets[0].LastName, check.Equals, "Example")
	c.Assert(targets[1].Email, check.Equals, "test2@example.com")
	c.Assert(targets[1].FirstName, check.Equals, "Second")
	c.Assert(targets[1].LastName, check.Equals, "Example")
}

func (s *ModelsSuite) TestPutGroupEmptyAttribute(c *check.C) {
	// Add test group.
	group := Group{Name: "Test Group"}
	group.Targets = []Target{
		Target{Email: "test1@example.com", FirstName: "First", LastName: "Example"},
		Target{Email: "test2@example.com", FirstName: "Second", LastName: "Example"},
	}
	group.UserId = 1
	PostGroup(&group)

	// Update one of group's targets.
	group.Targets[0].FirstName = ""
	err := PutGroup(&group)
	c.Assert(err, check.Equals, nil)

	// Verify updated empty attribute was saved.
	targets, _ := GetTargets(group.Id)
	c.Assert(targets[0].Email, check.Equals, "test1@example.com")
	c.Assert(targets[0].FirstName, check.Equals, "")
	c.Assert(targets[0].LastName, check.Equals, "Example")
	c.Assert(targets[1].Email, check.Equals, "test2@example.com")
	c.Assert(targets[1].FirstName, check.Equals, "Second")
	c.Assert(targets[1].LastName, check.Equals, "Example")
}

func (s *ModelsSuite) TestPostSMTP(c *check.C) {
	smtp := SMTP{
		Name:        "Test SMTP",
		Host:        "1.1.1.1:25",
		FromAddress: "Foo Bar <foo@example.com>",
		UserId:      1,
	}
	err = PostSMTP(&smtp)
	c.Assert(err, check.Equals, nil)
	ss, err := GetSMTPs(1)
	c.Assert(err, check.Equals, nil)
	c.Assert(len(ss), check.Equals, 1)
}

func (s *ModelsSuite) TestPostSMTPNoHost(c *check.C) {
	smtp := SMTP{
		Name:        "Test SMTP",
		FromAddress: "Foo Bar <foo@example.com>",
		UserId:      1,
	}
	err = PostSMTP(&smtp)
	c.Assert(err, check.Equals, ErrHostNotSpecified)
}

func (s *ModelsSuite) TestPostSMTPNoFrom(c *check.C) {
	smtp := SMTP{
		Name:   "Test SMTP",
		UserId: 1,
		Host:   "1.1.1.1:25",
	}
	err = PostSMTP(&smtp)
	c.Assert(err, check.Equals, ErrFromAddressNotSpecified)
}

func (s *ModelsSuite) TestPostPage(c *check.C) {
	html := `<html>
			<head></head>
			<body><form action="example.com">
				<input name="username"/>
				<input name="password" type="password"/>
			</form></body>
		  </html>`
	p := Page{
		Name:        "Test Page",
		HTML:        html,
		RedirectURL: "http://example.com",
	}
	// Check the capturing credentials and passwords
	p.CaptureCredentials = true
	p.CapturePasswords = true
	err := PostPage(&p)
	c.Assert(err, check.Equals, nil)
	c.Assert(p.RedirectURL, check.Equals, "http://example.com")
	d, err := goquery.NewDocumentFromReader(strings.NewReader(p.HTML))
	c.Assert(err, check.Equals, nil)
	forms := d.Find("form")
	forms.Each(func(i int, f *goquery.Selection) {
		// Check the action has been set
		a, _ := f.Attr("action")
		c.Assert(a, check.Equals, "")
		// Check the password still has a name
		_, ok := f.Find("input[type=\"password\"]").Attr("name")
		c.Assert(ok, check.Equals, true)
		// Check the username is still correct
		u, ok := f.Find("input").Attr("name")
		c.Assert(ok, check.Equals, true)
		c.Assert(u, check.Equals, "username")
	})
	// Check what happens when we don't capture passwords
	p.CapturePasswords = false
	p.HTML = html
	p.RedirectURL = ""
	err = PutPage(&p)
	c.Assert(err, check.Equals, nil)
	c.Assert(p.RedirectURL, check.Equals, "")
	d, err = goquery.NewDocumentFromReader(strings.NewReader(p.HTML))
	c.Assert(err, check.Equals, nil)
	forms = d.Find("form")
	forms.Each(func(i int, f *goquery.Selection) {
		// Check the action has been set
		a, _ := f.Attr("action")
		c.Assert(a, check.Equals, "")
		// Check the password still has a name
		_, ok := f.Find("input[type=\"password\"]").Attr("name")
		c.Assert(ok, check.Equals, false)
		// Check the username is still correct
		u, ok := f.Find("input").Attr("name")
		c.Assert(ok, check.Equals, true)
		c.Assert(u, check.Equals, "username")
	})
	// Finally, check when we don't capture credentials
	p.CaptureCredentials = false
	p.HTML = html
	err = PutPage(&p)
	c.Assert(err, check.Equals, nil)
	d, err = goquery.NewDocumentFromReader(strings.NewReader(p.HTML))
	c.Assert(err, check.Equals, nil)
	forms = d.Find("form")
	forms.Each(func(i int, f *goquery.Selection) {
		// Check the action has been set
		a, _ := f.Attr("action")
		c.Assert(a, check.Equals, "")
		// Check the password still has a name
		_, ok := f.Find("input[type=\"password\"]").Attr("name")
		c.Assert(ok, check.Equals, false)
		// Check the username is still correct
		_, ok = f.Find("input").Attr("name")
		c.Assert(ok, check.Equals, false)
	})
}
