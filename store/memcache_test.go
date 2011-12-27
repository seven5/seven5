package store

import (
	"fmt"
	"github.com/bradfitz/gomemcache"
	"launchpad.net/gocheck"
	"math/rand"
	//"os"
	"testing"
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

//we use the extra key "Birth_year" to test that our delete routines correctly update indexes.
type sample1 struct {
	Username   string
	Birth_year int `seven5key:"Birth_year"`
	Password   string
	Id         uint64
}

//the AggMonth method is used to demonstrate how to compute more complex key values to 
//allow you to get something like a complex query from this pile of hacks.
type BlarghParst struct {
	Title string
	YYYYMMDD  string `seven5key:"AggMonth"`
	Id    uint64
}

//AggMonth makes sure that it only returns the VALUE of the year and month rather than the data
//proper so that an index is computed based on this value.
func (self BlarghParst) AggMonth() string {
	return string(self.YYYYMMDD[0:6])
}

//This demonstrates to use a key "raw" without computing any function of it.  The extra index
//is based on the value of Foo.
type sample2 struct {
	Foo string `seven5key:"Foo"`
	Id  uint64
}

//we create a conn to the memcached at start of the suite
func (self *MemcachedSuite) SetUpSuite(c *gocheck.C) {
	self.store = &MemcacheGobStore{memcache.New(LOCALHOST)}
}

//no need to destry the connection, the program is ending anyway
func (self *MemcachedSuite) TearDownSuite(c *gocheck.C) {
}

//before each test we destroy all data in memcached.  if memcached is connected to a terminal
//(foreground) it will generate some bells to tell you that this is happening (annoying)
func (self *MemcachedSuite) SetUpTest(c *gocheck.C) {
	err := self.store.(*MemcacheGobStore).DestroyAll(LOCALHOST)
	if err != nil {
		c.Fatal("unable to setup test and clear memcached:%s\n", err)
	}
}

//test that we can work at the memcached level, since the rest of tests are at the store.T
//level
func (self *MemcachedSuite) TestMemcacheLevelSetupWorks1(c *gocheck.C) {
	m := self.store.(*MemcacheGobStore)

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

//make sure the store is empty at start
func (self *MemcachedSuite) TestMemcacheLevelSetupWorks2(c *gocheck.C) {
	m := self.store.(*MemcacheGobStore)

	_, err := m.Get("fart")
	if err != memcache.ErrCacheMiss {
		c.Fatal("expected all the cache entries to be cleared before each test!")
	}
}

//utility routine that is used a few places to write an instance of sample1
//note that the parameters to the storage layer are a pointer to the structure, even in cases
//where the structure is not modified.  
func (self *MemcachedSuite) WriteSample1(user string, yr int, pwd string, c *gocheck.C) *sample1 {
	s := &sample1{Username: user, Birth_year: yr, Password: pwd}
	if err:=self.store.Write(s); err!=nil {
		c.Fatalf("failed to write a sample1:%v",err)
	}
	return s
}

//basic read/write test at the level of store.T (in this case, self.store)
func (self *MemcachedSuite) TestBasicStoreWithId(c *gocheck.C) {
	iansmith := "iansmith"
	yr := 1970
	pwd := "fart"

	t1:= self.WriteSample1(iansmith,yr,pwd,c)
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

//create a set of blog posts on various dates and show that they get aggregated together
//into bunches based on the AggMonth method.
func (self *MemcachedSuite) TestExtraKeyNames(c *gocheck.C) {
	quick := "the quick and the dead"
	mostly := "the mostly dead"
	jim := "it's dead, jim"
	xmas := "death in december"
	newyear := "NYE is the amateur hour of drunks"

	p1 := &BlarghParst{quick, "20111222", uint64(0)}
	p2 := &BlarghParst{mostly, "20111223", uint64(0)}
	p3 := &BlarghParst{jim, "20111224", uint64(0)}
	p4 := &BlarghParst{xmas, "20111225", uint64(0)}
	p5 := &BlarghParst{newyear, "20120101", uint64(0)}

	c.Check(self.store.Write(p1), gocheck.Equals, nil)
	c.Check(self.store.Write(p2), gocheck.Equals, nil)
	c.Check(self.store.Write(p3), gocheck.Equals, nil)
	c.Check(self.store.Write(p4), gocheck.Equals, nil)
	c.Check(self.store.Write(p5), gocheck.Equals, nil)

	//four posts in 2011 12
	hits := make([]*BlarghParst, 0, 5)
	err := self.store.FindByKey(&hits, "AggMonth", "201112")

	c.Check(err, gocheck.Equals, nil)
	c.Check(len(hits), gocheck.Equals, 4)
	//order of return in the slice is not guaranteed!
	for i:=0;i<len(hits);i++ {
		p:=hits[i]
		switch p.Id {
		case 1:
			c.Check(p.Title,gocheck.Equals,quick)
		case 2:
			c.Check(p.Title,gocheck.Equals,mostly)
		case 3:
			c.Check(p.Title,gocheck.Equals,jim)
		case 4:
			c.Check(p.Title,gocheck.Equals,xmas)
			c.Check(p.YYYYMMDD, gocheck.Equals,"20111225")
		default:
			c.Fatalf("didn't expect that id in the list! %d",p.Id)
		}
	}

	//no overflow?
	hits = make([]*BlarghParst, 0, 1)
	err = self.store.FindByKey(&hits, "AggMonth", "201112")
	c.Check(err, gocheck.Equals, nil)
	c.Check(len(hits), gocheck.Equals, 1)

	//nothing in dec 2012
	hits = make([]*BlarghParst, 0, 5)
	err = self.store.FindByKey(&hits, "AggMonth", "201212")
	c.Check(err, gocheck.Equals, nil)
	c.Check(len(hits), gocheck.Equals, 0)

}

//test deleting works and that the indexes get updated properly
func (self *MemcachedSuite) TestDeleteItems(c *gocheck.C) {
	t1:= self.WriteSample1("iansmith",1970,"fart",c)
	t2:= self.WriteSample1("trevorsmith",1972,"yech",c)
	
	//load a copy of t2
	t3:=new (sample1)
	err:=self.store.FindById(t3,t2.Id)
	
	c.Check(err,gocheck.Equals,nil)
	c.Check(t3.Id,gocheck.Equals,t2.Id)
	c.Check(t3.Birth_year,gocheck.Equals,t2.Birth_year)
	c.Check(t3.Username,gocheck.Equals,t2.Username)
	
	//check we can get it by key
	hits:=make([]*sample1,0,1)
	err=self.store.FindByKey(&hits, "Birth_year","1972")
	c.Check(err,gocheck.Equals,nil)
	c.Assert(1,gocheck.Equals,len(hits))
	c.Check(hits[0].Username,gocheck.Equals,"trevorsmith")

	//check we can't find it by wrong key
	hits=make([]*sample1,0,1)
	err=self.store.FindByKey(&hits, "Birth_year","1973")
	c.Check(err,gocheck.Equals,nil)
	c.Assert(0,gocheck.Equals,len(hits))

	//check we can't delete stuff not there
	err=self.store.DeleteById(t2,429)
	c.Check(err,gocheck.Equals,memcache.ErrCacheMiss)
	
	//delete trev
	err=self.store.DeleteById(t2,t2.Id)
	c.Check(err,gocheck.Equals,nil)
	
	//now can't find trev
	err=self.store.FindById(t3,t2.Id)
	c.Check(err,gocheck.Equals,memcache.ErrCacheMiss)
	
	//still can find ian
	err=self.store.FindById(t3,t1.Id)
	c.Check(err,gocheck.Equals,nil)
	c.Check(t3.Username,gocheck.Equals,"iansmith")
	
	//still can find ian by year
	hits=make([]*sample1,0,1)
	err=self.store.FindByKey(&hits, "Birth_year","1970")
	c.Check(err,gocheck.Equals,nil)
	c.Check(1,gocheck.Equals,len(hits))
	c.Check(hits[0].Username,gocheck.Equals,"iansmith")

	//but not trevor by year
	hits=make([]*sample1,0,1)
	err=self.store.FindByKey(&hits, "Birth_year","1972")
	c.Check(err,gocheck.Equals,nil)
	c.Check(0,gocheck.Equals,len(hits))
}

//
// Benchmarks
//

var letter = []string{"a", "b", "c", "d", "e", "f", "g", "h", "i", "j", "k", "l", "m", "n", "o", "p", "q", "r", "s", "t", "u", "v", "w", "x", "y", "z"}

func randomKey() string {
	return letter[rand.Intn(len(letter))] + letter[rand.Intn(len(letter))]  + letter[rand.Intn(len(letter))]
}

func sampleData(size int) []sample2 {
	result := make([]sample2, size, size)
	for i := 0; i < size; i++ {
		k := randomKey()
		result[i] = sample2{k, 0}
	}
	return result
}

func BenchmarkWriteSpeed(b *testing.B) {
	b.StopTimer()
	store := &MemcacheGobStore{memcache.New(LOCALHOST)}
	data := sampleData(b.N)
	fmt.Printf("Write speed test: %d items...\n",b.N)
	b.StartTimer()
	for i := 0; i < b.N; i++ {
		if err := store.Write(&data[i]); err != nil {
			b.Fatalf("unable to write sample data!%v", err)
		}
	}
}

var haveWrittenSampleData = false

func BenchmarkSelectSpeed(b *testing.B) {
	b.StopTimer()
	store := &MemcacheGobStore{memcache.New(LOCALHOST)}
	if !haveWrittenSampleData {
		haveWrittenSampleData=true
		size:=10000
		fmt.Printf("constructing sample data set of %d items...\n",size)
		data := sampleData(size)
		for i := 0; i < size; i++ {
			if err := store.Write(&data[i]); err != nil {
				b.Fatalf("unable to write sample data!%v\n", err)
			}
		}
	}
	target:=randomKey()
	b.StartTimer()
	fmt.Printf("benchmarking search: %d searches\n",b.N)
	for i := 0; i < b.N; i++ {
		hits := make([]*sample2, 0, 5)
		if err := store.FindByKey(&hits, "Foo", target); err != nil {
			b.Fatalf("unable to read sample data!%v\n", err)
		}
	}
}
