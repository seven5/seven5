package store

import (
	"fmt"
	"github.com/bradfitz/gomemcache/memcache"
	"launchpad.net/gocheck"
	"math/rand"
	//"os"
	"reflect"
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
	Title    string
	YYYYMMDD string `seven5key:"AggMonth"`
	Id       uint64 `seven5order:"none"`
}

//AggMonth makes sure that it only returns the VALUE of the year and month rather than the data
//proper so that an index is computed based on this value.
func (self BlarghParst) AggMonth() string {
	return string(self.YYYYMMDD[0:6])
}

//Make sure we meet the Lesser interface that allows us to always have sorted results.
func (self BlarghParst) Less(i reflect.Value, j reflect.Value) bool {
	var x, y *BlarghParst
	//this converts to our type, which is not known to the infrastructure
	reflect.Indirect(reflect.ValueOf(&x)).Set(i)
	reflect.Indirect(reflect.ValueOf(&y)).Set(j)
	//conversion accepted
	return x.YYYYMMDD < y.YYYYMMDD
}

//This demonstrates to use a key "raw" without computing any function of it.  The extra index
//is based on the value of Foo. 
type sample2 struct {
	Foo string `seven5key:"Foo"`
	Id  uint64 `seven5order:"none"`
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
	if err := self.store.Write(s); err != nil {
		c.Fatalf("failed to write a sample1:%v", err)
	}
	return s
}

//basic read/write test at the level of store.T (in this case, self.store)
func (self *MemcachedSuite) TestBasicStoreWithId(c *gocheck.C) {
	iansmith := "iansmith"
	yr := 1970
	pwd := "fart"

	t1 := self.WriteSample1(iansmith, yr, pwd, c)
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
	p2 := &BlarghParst{mostly, "20111203", uint64(0)}
	p3 := &BlarghParst{jim, "20111214", uint64(0)}
	p4 := &BlarghParst{xmas, "20111225", uint64(0)}
	p5 := &BlarghParst{newyear, "20120101", uint64(0)}

	c.Assert(self.store.Write(p1), gocheck.Equals, nil)
	c.Assert(self.store.Write(p2), gocheck.Equals, nil)
	c.Assert(self.store.Write(p3), gocheck.Equals, nil)
	c.Assert(self.store.Write(p4), gocheck.Equals, nil)
	c.Assert(self.store.Write(p5), gocheck.Equals, nil)

	//we turned off all keys
	hits := make([]*BlarghParst, 0, 10)
	err := self.store.FindAll(&hits, uint64(0))
	c.Check(len(hits), gocheck.Equals, 0)

	//four posts in dec 2011 
	hits = make([]*BlarghParst, 0, 5)
	err = self.store.FindByKey(&hits, "AggMonth", "201112", uint64(0))

	c.Assert(err, gocheck.Equals, nil)
	c.Check(len(hits), gocheck.Equals, 4)

	//order is increasing by date
	c.Check(hits[0].Title, gocheck.Equals, mostly)
	c.Check(hits[1].Title, gocheck.Equals, jim)
	c.Check(hits[2].Title, gocheck.Equals, quick)
	c.Check(hits[3].Title, gocheck.Equals, xmas)

	//no overflow?
	hits = make([]*BlarghParst, 0, 1)
	err = self.store.FindByKey(&hits, "AggMonth", "201112", uint64(0))
	c.Assert(err, gocheck.Equals, nil)
	c.Check(len(hits), gocheck.Equals, 1)

	//nothing in dec 2012
	hits = make([]*BlarghParst, 0, 5)
	err = self.store.FindByKey(&hits, "AggMonth", "201212", uint64(0))
	c.Assert(err, gocheck.Equals, nil)
	c.Check(len(hits), gocheck.Equals, 0)

	//look for all the values
	
	uniqueValues:=make([]ValueInfo,0,10)
	example:=new(BlarghParst)
	err = self.store.UniqueKeyValues(example,"AggMonth",&uniqueValues,uint64(0))
	c.Assert(err, gocheck.Equals, nil)
	c.Assert(len(uniqueValues),gocheck.Equals,2)
	c.Check(uniqueValues[0].Value,gocheck.Not(gocheck.Equals),uniqueValues[1].Value)
	if uniqueValues[0].Value!="201201" && uniqueValues[0].Value!="201112"{
		c.Fatalf("bad key found in uniqueValues[0]:%s",uniqueValues[0].Value)
	}
	if uniqueValues[1].Value!="201201" && uniqueValues[1].Value!="201112"{
		c.Fatalf("bad key found in uniqueValues[1]:%s",uniqueValues[1].Value)
	}
	
}

//test deleting works and that the indexes get updated properly
func (self *MemcachedSuite) TestDeleteItems(c *gocheck.C) {
	t1 := self.WriteSample1("iansmith", 1970, "fart", c)
	c.Assert(t1.Id, gocheck.Not(gocheck.Equals), 0)
	t2 := self.WriteSample1("trevorsmith", 1972, "yech", c)
	c.Assert(t2.Id, gocheck.Not(gocheck.Equals), 0)

	//all is on, so see if can get them all
	hits := make([]*sample1, 0, 5)
	err := self.store.FindAll(&hits, uint64(0))
	c.Check(err, gocheck.Equals, nil)
	c.Check(len(hits), gocheck.Equals, 2)

	//load a copy of t2
	t3 := new(sample1)
	err = self.store.FindById(t3, t2.Id)

	c.Check(err, gocheck.Equals, nil)
	c.Check(t3.Id, gocheck.Equals, t2.Id)
	c.Check(t3.Birth_year, gocheck.Equals, t2.Birth_year)
	c.Check(t3.Username, gocheck.Equals, t2.Username)

	//check we can get it by key
	hits = make([]*sample1, 0, 1)
	err = self.store.FindByKey(&hits, "Birth_year", "1972", uint64(0))
	c.Assert(err, gocheck.Equals, nil)
	c.Assert(1, gocheck.Equals, len(hits))
	c.Check(hits[0].Username, gocheck.Equals, "trevorsmith")

	//check we can't find it by wrong key
	hits = make([]*sample1, 0, 1)
	err = self.store.FindByKey(&hits, "Birth_year", "1973", uint64(0))
	c.Assert(err, gocheck.Equals, nil)
	c.Assert(0, gocheck.Equals, len(hits))

	//check we can't delete stuff not there
	t2.Id = 429
	err = self.store.Delete(t2)
	c.Check(err, gocheck.Equals, memcache.ErrCacheMiss)

	//check that we cannot write stuff that is not there
	t2.Id = 429
	err = self.store.Write(t2)
	c.Check(err, gocheck.Equals, memcache.ErrCacheMiss)

	//delete trev
	t2.Id = uint64(2)
	err = self.store.Delete(t2)
	c.Check(err, gocheck.Equals, nil)

	//now can't find trev
	err = self.store.FindById(t3, uint64(2))
	c.Check(err, gocheck.Equals, memcache.ErrCacheMiss)

	//still can find ian
	err = self.store.FindById(t3, t1.Id)
	c.Check(err, gocheck.Equals, nil)
	c.Check(t3.Username, gocheck.Equals, "iansmith")

	//still can find ian by year
	hits = make([]*sample1, 0, 1)
	err = self.store.FindByKey(&hits, "Birth_year", "1970", uint64(0))
	c.Check(err, gocheck.Equals, nil)
	c.Check(1, gocheck.Equals, len(hits))
	c.Check(hits[0].Username, gocheck.Equals, "iansmith")

	//but not trevor by year
	hits = make([]*sample1, 0, 1)
	err = self.store.FindByKey(&hits, "Birth_year", "1972", uint64(0))
	c.Check(err, gocheck.Equals, nil)
	c.Check(0, gocheck.Equals, len(hits))
}

//
// ORDERING TEST
//

type lifo struct {
	Name string
	Id   uint64 `seven5order:"lifo"`
}

type fifo struct {
	Name string
	Id   uint64 `seven5order:"fifo"`
}

type neither struct {
	Name string
	Id   uint64
}

//test ordering works properly, if we force it with seven5order
func (self *MemcachedSuite) TestOrderOfItems(c *gocheck.C) {

	stackItem := &lifo{"abc", 0}
	queueItem := &fifo{"abc", 0}
	item := &neither{"abc", 0}

	c.Assert(self.store.Write(stackItem), gocheck.Equals, nil)
	c.Assert(self.store.Write(queueItem), gocheck.Equals, nil)
	c.Assert(self.store.Write(item), gocheck.Equals, nil)

	stackItem = &lifo{"def", 0}
	queueItem = &fifo{"def", 0}
	item = &neither{"def", 0}

	c.Assert(self.store.Write(stackItem), gocheck.Equals, nil)
	c.Assert(self.store.Write(queueItem), gocheck.Equals, nil)
	c.Assert(self.store.Write(item), gocheck.Equals, nil)

	stackItem = &lifo{"ghi", 0}
	queueItem = &fifo{"ghi", 0}
	item = &neither{"ghi", 0}

	c.Assert(self.store.Write(stackItem), gocheck.Equals, nil)
	c.Assert(self.store.Write(queueItem), gocheck.Equals, nil)
	c.Assert(self.store.Write(item), gocheck.Equals, nil)

	stack := make([]*lifo, 0, 10)
	queue := make([]*fifo, 0, 10)
	dontknow := make([]*neither, 0, 10)

	c.Assert(self.store.FindAll(&stack, uint64(0)), gocheck.Equals, nil)
	c.Assert(self.store.FindAll(&queue, uint64(0)), gocheck.Equals, nil)
	c.Assert(self.store.FindAll(&dontknow, uint64(0)), gocheck.Equals, nil)

	c.Check(len(stack), gocheck.Equals, 3)
	c.Check(len(queue), gocheck.Equals, 3)
	c.Check(len(dontknow), gocheck.Equals, 3)

	c.Check(stack[0].Name, gocheck.Equals, "ghi")
	c.Check(stack[1].Name, gocheck.Equals, "def")
	c.Check(stack[2].Name, gocheck.Equals, "abc")

	c.Check(queue[0].Name, gocheck.Equals, "abc")
	c.Check(queue[1].Name, gocheck.Equals, "def")
	c.Check(queue[2].Name, gocheck.Equals, "ghi")

	//can be in order, but must appear exactly once
	c.Check(isValidForDontKnow(dontknow[0].Name, true), gocheck.Equals, true)
	c.Check(isValidForDontKnow(dontknow[1].Name, true), gocheck.Equals, true)
	c.Check(isValidForDontKnow(dontknow[2].Name, true), gocheck.Equals, true)
	c.Check(dontknow[0].Name, gocheck.Not(gocheck.Equals), dontknow[1].Name)
	c.Check(dontknow[0].Name, gocheck.Not(gocheck.Equals), dontknow[2].Name)
	c.Check(dontknow[1].Name, gocheck.Not(gocheck.Equals), dontknow[2].Name)

	//delete last item from each one
	stackItem.Id = uint64(3)
	c.Assert(self.store.Delete(stackItem), gocheck.Equals, nil)
	queueItem.Id = uint64(3)
	c.Assert(self.store.Delete(queueItem), gocheck.Equals, nil)
	item.Id = uint64(3)
	c.Assert(self.store.Delete(item), gocheck.Equals, nil)

	c.Assert(self.store.FindAll(&stack, uint64(0)), gocheck.Equals, nil)
	c.Assert(self.store.FindAll(&queue, uint64(0)), gocheck.Equals, nil)
	c.Assert(self.store.FindAll(&dontknow, uint64(0)), gocheck.Equals, nil)

	//note: reading these into positions #3 and #4 since first three are already filled
	//from last call to findAll!
	c.Check(len(stack), gocheck.Equals, 5)
	c.Check(len(queue), gocheck.Equals, 5)
	c.Check(len(dontknow), gocheck.Equals, 5)

	c.Check(stack[3].Name, gocheck.Equals, "def")
	c.Check(stack[4].Name, gocheck.Equals, "abc")

	c.Check(queue[3].Name, gocheck.Equals, "abc")
	c.Check(queue[4].Name, gocheck.Equals, "def")

	c.Check(isValidForDontKnow(dontknow[3].Name, false), gocheck.Equals, true)
	c.Check(isValidForDontKnow(dontknow[4].Name, false), gocheck.Equals, true)
	c.Check(dontknow[3].Name, gocheck.Not(gocheck.Equals), dontknow[4].Name)

}

func isValidForDontKnow(name string, includeGHI bool) bool {
	if name == "abc" || name == "def" {
		return true
	}
	if includeGHI {
		if name == "ghi" {
			return true
		}
	}
	return false
}

//test that init creates the necessary structures
func (self *MemcachedSuite) TestInit(c *gocheck.C) {
	self.store.Init(&sample1{})
	hits := make([]*sample1, 0, 1)
	err := self.store.FindByKey(&hits, "Birth_year", "1970", uint64(0))
	c.Check(err, gocheck.Equals, nil)
	c.Check(0, gocheck.Equals, len(hits))
	err = self.store.FindAll(&hits, uint64(0))
	c.Check(err, gocheck.Equals, nil)
	c.Check(0, gocheck.Equals, len(hits))
}

type youzer struct {
	Name    string
	Friends []uint64 //we could do a pointers here but that implies "early" fetching (heh!)
	Id      uint64
}

type freundreck struct {
	To    uint64 `seven5key:"To"`
	Owner uint64 //note, this is the SENDER of the freundreck
	Id    uint64 
}

//http://www.youtube.com/watch?v=EcHP1tWWEvI&feature=related
func (self *MemcachedSuite) TestOwner(c *gocheck.C) {
	lenny := &youzer{"Lenny", []uint64{}, uint64(0)}
	carl := &youzer{"Carl", []uint64{}, uint64(0)}
	homer := &youzer{"Homer", []uint64{}, uint64(0)}
	moe := &youzer{"Moe", []uint64{}, uint64(0)}
	duffman := &youzer{"Duffman", []uint64{}, uint64(0)}

	c.Assert(self.store.Write(lenny), gocheck.Equals, nil)
	c.Assert(self.store.Write(carl), gocheck.Equals, nil)
	c.Assert(self.store.Write(homer), gocheck.Equals, nil)
	c.Assert(self.store.Write(moe), gocheck.Equals, nil)
	c.Assert(self.store.Write(duffman), gocheck.Equals, nil)
	
	//lenny are carl are friends (?)
	lenny.Friends=[]uint64{carl.Id}
	carl.Friends=[]uint64{lenny.Id}

	//moe and homer are friends, more or less
	homer.Friends=[]uint64{moe.Id,lenny.Id}
	moe.Friends=[]uint64{homer.Id, lenny.Id, carl.Id}

	//note that we don't write the friends until we have established the Id fields in the structs
	c.Assert(self.store.Write(lenny), gocheck.Equals, nil)
	c.Assert(self.store.Write(carl), gocheck.Equals, nil)
	c.Assert(self.store.Write(homer), gocheck.Equals, nil)
	c.Assert(self.store.Write(moe), gocheck.Equals, nil)
	
	//duffman wants to get friends by spamming, capitalist pig that he is 
	req1:=&freundreck{homer.Id,duffman.Id,uint64(0)}
	req2:=&freundreck{moe.Id,duffman.Id,uint64(0)}
	req3:=&freundreck{lenny.Id,duffman.Id,uint64(0)}
	
	//he sends the spam
	c.Assert(self.store.Write(req1), gocheck.Equals, nil)
	c.Assert(self.store.Write(req2), gocheck.Equals, nil)
	c.Assert(self.store.Write(req3), gocheck.Equals, nil)
	
	//check that it is present
	targets:=make([]ValueInfo,0,5)
	c.Assert(self.store.UniqueKeyValues(req1, "To",  &targets, duffman.Id),gocheck.Equals,nil)
	c.Check(len(targets),gocheck.Equals,3)
	c.Check(targets[0].Value,gocheck.Not(gocheck.Equals),targets[1].Value)
	c.Check(targets[1].Value,gocheck.Not(gocheck.Equals),targets[2].Value)
	c.Check(targets[0].Value,gocheck.Not(gocheck.Equals),targets[2].Value)
	
	//so, we have 3 pending friend reqs from duffman, none from home
	hitsDuffman:=make([]*freundreck,0,5)
	hitsHomer:=make([]*freundreck,0,5)
	c.Assert(self.store.FindAll(&hitsDuffman,duffman.Id),gocheck.Equals,nil)
	c.Assert(self.store.FindAll(&hitsHomer,homer.Id),gocheck.Equals,nil)
	
	//homer has none pending, but duffman does
	c.Check(len(hitsHomer),gocheck.Equals,0)
	c.Check(len(hitsDuffman),gocheck.Equals,3)
	
	//but does anybody want to be freunds with homey?
	hitsHomer=make([]*freundreck,0,5)
	c.Assert(self.store.FindByKey(&hitsHomer,"To",fmt.Sprintf("%d",homer.Id),uint64(0)),gocheck.Equals,nil)

	//yes, it's the duffman, oh yeah
	c.Assert(len(hitsHomer),gocheck.Equals,1)
	c.Check(hitsHomer[0].To, gocheck.Equals, homer.Id)
	c.Check(hitsHomer[0].Owner, gocheck.Equals, duffman.Id)
	
}


//
// Benchmarks
//

var letter = []string{"a", "b", "c", "d", "e", "f", "g", "h", "i", "j", "k", "l", "m", "n", "o", "p", "q", "r", "s", "t", "u", "v", "w", "x", "y", "z"}

func randomKey() string {
	return letter[rand.Intn(len(letter))] + letter[rand.Intn(len(letter))] + letter[rand.Intn(len(letter))] + letter[rand.Intn(len(letter))] //+letter[rand.Intn(len(letter))]
}

func sampleData(size int) []*sample2 {
	result := make([]*sample2, size, size)
	for i := 0; i < size; i++ {
		k := randomKey()
		result[i] = &sample2{k, 0}
	}
	return result
}

func BenchmarkWriteSpeed(b *testing.B) {
	b.StopTimer()
	store := &MemcacheGobStore{memcache.New(LOCALHOST)}
	data := sampleData(b.N)
	fmt.Printf("Write speed test: %d items...\n", b.N)
	b.StartTimer()
	for i := 0; i < b.N; i++ {
		if err := store.Write(data[i]); err != nil {
			b.Fatalf("unable to write sample data!%v", err)
		}
	}
}

var haveWrittenSampleData = false

func BenchmarkSelectSpeed(b *testing.B) {
	b.StopTimer()
	store := &MemcacheGobStore{memcache.New(LOCALHOST)}
	if !haveWrittenSampleData {
		haveWrittenSampleData = true
		size := 5000
		fmt.Printf("constructing sample data set of %d items...\n", size)
		data := sampleData(size)
		if err := store.BulkWrite(data); err != nil {
			b.Fatalf("unable to write sample data!%v\n", err)
		}
	}
	target := randomKey()
	fmt.Printf("benchmarking search: %d searches\n", b.N)
	b.StartTimer()
	for i := 0; i < b.N; i++ {
		hits := make([]*sample2, 0, 5)
		if err := store.FindByKey(&hits, "Foo", target, uint64(0)); err != nil {
			b.Fatalf("unable to read sample data!%v\n", err)
		}
	}
}

func BenchmarkWriteOverhead(b *testing.B) {
	b.StopTimer()
	store := &MemcacheGobStore{memcache.New(LOCALHOST)}
	item := &memcache.Item{Key: "key", Value: []byte("0")}
	b.StartTimer()
	for i := 0; i < b.N; i++ {
		store.Client.Set(item)
	}
}

func BenchmarkReadOverhead(b *testing.B) {
	b.StopTimer()
	store := &MemcacheGobStore{memcache.New(LOCALHOST)}
	item := &memcache.Item{Key: "key", Value: []byte("0")}
	store.Client.Set(item)
	b.StartTimer()
	for i := 0; i < b.N; i++ {
		_, _ = store.Client.Get("key")
	}
}
