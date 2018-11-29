package models

import (
	"testing"

	"github.com/gophish/gophish/config"
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
	db.Delete(Result{})
	db.Delete(MailLog{})
	db.Delete(Campaign{})

	// Reset users table to default state.
	db.Not("id", 1).Delete(User{})
	db.Model(User{}).Update("username", "admin")
}

func (s *ModelsSuite) createCampaignDependencies(ch *check.C, optional ...string) Campaign {

	//Add a dummy public key to the admin user
	pubKey := &PublicKey{Id: 1, FriendlyName: "Happy", UserId: 1,
		PubKey: `-----BEGIN PUBLIC KEY-----
MIIBIjANBgkqhkiG9w0BAQEFAAOCAQ8AMIIBCgKCAQEAz8qUODbqjWxcL8eNngjC
fwO6bstHOt6p8EvHahem6JQ/VeIdJ4h7Hy0eTxm68sXKvliWrs6J3uJUAAZlZqX5
E9uMaSjiaF+aLVjQOj+fqmh/+UnpZUa/p2WtWy1YyZuAZ9o4EcbeUFokkS8oIXfW
vyPrE5ggAUNiW4p/4iHltjqKyt8z2cids36j09OLz0hnGxzAq4PQvdYnW0OyLkbq
FwvpiR8/9JY4O7pM4dUaQBQhvj+ahbuYhdO+tsnE7cRMOLNXfc8vDtdTY08BfL5Z
vFsuNexQlF1DnL5VIETx9WHmbT77A00VJp3VeUxADpYoSyrKQ5settc+dSFvp7kP
qwIDAQAB
-----END PUBLIC KEY-----`,
	}
	ch.Assert(PostPublicKey(pubKey), check.Equals, nil)

	// we use the optional parameter to pass an alternative subject
	group := Group{Name: "Test Group"}
	group.Targets = []Target{
		Target{BaseRecipient: BaseRecipient{Email: "test1@example.com", FirstName: "First", LastName: "Example"}},
		Target{BaseRecipient: BaseRecipient{Email: "test2@example.com", FirstName: "Second", LastName: "Example"}},
		Target{BaseRecipient: BaseRecipient{Email: "test3@example.com", FirstName: "Second", LastName: "Example"}},
		Target{BaseRecipient: BaseRecipient{Email: "test4@example.com", FirstName: "Second", LastName: "Example"}},
	}
	group.UserId = 1
	ch.Assert(PostGroup(&group), check.Equals, nil)

	// Add a template
	t := Template{Name: "Test Template"}
	if len(optional) > 0 {
		t.Subject = optional[0]
	} else {
		t.Subject = "{{.RId}} - Subject"
	}
	t.Text = "{{.RId}} - Text"
	t.HTML = "{{.RId}} - HTML"
	t.UserId = 1
	ch.Assert(PostTemplate(&t), check.Equals, nil)

	// Add a landing page
	p := Page{Name: "Test Page"}
	p.HTML = "<html>Test</html>"
	p.UserId = 1
	ch.Assert(PostPage(&p), check.Equals, nil)

	// Add a sending profile
	smtp := SMTP{Name: "Test Page"}
	smtp.UserId = 1
	smtp.Host = "example.com"
	smtp.FromAddress = "test@test.com"
	ch.Assert(PostSMTP(&smtp), check.Equals, nil)

	c := Campaign{Name: "Test campaign"}
	c.UserId = 1
	c.Template = t
	c.Page = p
	c.SMTP = smtp
	c.PublicKeyId = 1
	c.Groups = []Group{group}
	return c
}

func (s *ModelsSuite) createCampaign(ch *check.C) Campaign {
	c := s.createCampaignDependencies(ch)
	// Setup and "launch" our campaign
	ch.Assert(PostCampaign(&c, c.UserId), check.Equals, nil)
	return c
}
