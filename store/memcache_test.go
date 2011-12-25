package store

import (
	"github.com/bradfitz/gomemcache"
	"launchpad.net/gocheck"
	"testing"
	//	"fmt"
	//	"os"
)

// Hook up gocheck into the default gotest runner.
func Test(t *testing.T) { gocheck.TestingT(t) }

// This is the "suite" structure for objects that need to survive the whole of the tests
// in this file.
type MemcachedSuite struct {
	store T
}

var s = &MemcachedSuite{}

// hook up suite to gocheck
var _ = gocheck.Suite(s)

type sample1 struct {
	Username   string
	Birth_year int
	Password   string
	Id         uint64
}

type sample2 struct {
	Username string `seven5key:"true"`
	Password string
	Id       uint64
}

func (self *MemcachedSuite) SetUpSuite(c *gocheck.C) {
	self.store = &MemcacheGobStore{memcache.New(LOCALHOST)}
}

func (self *MemcachedSuite) TearDownSuite(c *gocheck.C) {
}

func (self *MemcachedSuite) SetUpTest(c *gocheck.C) {
	err := self.store.DestroyAll(LOCALHOST)
	if err != nil {
		c.Fatal("unable to setup test and clear memcached:%s\n", err)
	}
}

func (self *MemcachedSuite) TestMemcacheLevelSetupWorks1(c *gocheck.C) {
	m:=self.store.(*MemcacheGobStore)
	
	item := &memcache.Item{Key: "fart", Value: []byte("fartvalue")}
	err := m.Set(item)
	if err != nil {
		c.Fatal("unable to set test value (fart) to check suite is working ok")
	}
	it, err := m.Get("fart")
	if err != nil {
		c.Fatal("unable to get test value (fart) to check suite is working ok")
	}
	if it.Key != "fart" || string(it.Value) != "fartvalue" {
		c.Fatal("unable to correctly read test value (fart) to check suite is working ok")
	}
}

func (self *MemcachedSuite) TestMemcacheLevelSetupWorks2(c *gocheck.C) {
	m:=self.store.(*MemcacheGobStore)
	
	_, err := m.Get("fart")
	if err != memcache.ErrCacheMiss {
		c.Fatal("expected all the cache entries to be cleared before each test!")
	}
}

func (self *MemcachedSuite) TestBasicStoreWithId(c *gocheck.C) {
	iansmith := "iansmith"
	yr := 1970
	pwd := "fart"

	t1 := &sample1{Username: iansmith, Birth_year: yr, Password: pwd}
	t2 := new(sample1)

	err := self.store.Write(t1)

	c.Check(err, gocheck.Equals, nil)
	c.Check(t1.Id, gocheck.Equals, uint64(1))

	err = self.store.FindById(t2, 1)
	c.Check(err, gocheck.Equals, nil)

	c.Check(t2.Id, gocheck.Equals, uint64(1))
	c.Check(t2.Username, gocheck.Equals, iansmith)
	c.Check(t2.Birth_year, gocheck.Equals, yr)
	c.Check(t2.Password, gocheck.Equals, pwd)
}

func (self *MemcachedSuite) TestExtraKeyNames(c *gocheck.C) {
	iansmith := "iansmith"
	pwd := "fart"

	s := &sample2{Username: iansmith, Password: pwd}

	keys := GetStructKeys(s)
	c.Check(len(keys), gocheck.Equals, 1)
	c.Check(keys[0], gocheck.Equals, "Username")

	err := self.store.Write(s)
	c.Check(err, gocheck.Equals, nil)
	c.Check(s.Id, gocheck.Equals, uint64(1))

	example := &sample2{}
	err = self.store.FindByKey(example, "Username", "foo")

	c.Check(err, gocheck.Equals, NO_SUCH_KEY)

	err = self.store.FindByKey(example, "Username", iansmith)
	c.Check(err, gocheck.Equals, nil)
	c.Check(s.Username, gocheck.Equals, iansmith)
	c.Check(s.Password, gocheck.Equals, pwd)

}
