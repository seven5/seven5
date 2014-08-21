package seven5

import (
	"errors"
	"fmt"
	"github.com/coocood/qbs"
	_ "github.com/lib/pq"
	"net/http"
	"testing"
)

const (
	none        = 0
	force_error = 1
	force_panic = 2
)

/*---- type of actual DB action ----*/
type House struct {
	Id      int64
	Address string
	Zip     int64 /*0->99999, inclusive*/
}

/*---- wire type for the tests ----*/
type HouseWire struct {
	Id      int64
	Addr    string
	ZipCode int64
}

type testObj struct {
	testCallCount int
	failPost      int
}

/*these funcs are use to test that if you meet the QBSRest interfaces you
can be wrapped by the qbs code in Seven5 */
func (self *testObj) IndexQbs(pb PBundle, q *qbs.Qbs) (interface{}, error) {
	self.testCallCount++
	return &HouseWire{}, nil
}
func (self *testObj) FindQbs(id int64, pb PBundle, q *qbs.Qbs) (interface{}, error) {
	self.testCallCount++
	return &HouseWire{}, nil
}
func (self *testObj) DeleteQbs(id int64, pb PBundle, q *qbs.Qbs) (interface{}, error) {
	self.testCallCount++
	return &HouseWire{}, nil
}
func (self *testObj) PutQbs(id int64, value interface{}, pb PBundle, q *qbs.Qbs) (interface{}, error) {
	self.testCallCount++
	return &HouseWire{}, nil
}
func (self *testObj) PostQbs(value interface{}, pb PBundle, q *qbs.Qbs) (interface{}, error) {
	self.testCallCount++
	in := value.(*HouseWire)
	house := &House{Address: in.Addr, Zip: in.ZipCode}
	if _, err := q.Save(house); err != nil {
		return nil, err
	}

	//simulate an error AFTER the save!
	switch self.failPost {
	case force_panic:
		panic("testing panic handling")
	case force_error:
		return nil, errors.New("testing error handling")
	}

	return &HouseWire{Id: house.Id, Addr: house.Address, ZipCode: house.Zip}, nil
}

/*-------------------------------------------------------------------------*/
/*                                 TEST CODE                               */
/*-------------------------------------------------------------------------*/
const (
	APP_NAME = "testapp"
)

func checkNumberHouses(T *testing.T, store *QbsStore, expected int) {
	houses := []*House{}
	err := store.Q.FindAll(&houses)
	if err != nil {
		T.Fatalf("Error during find: %s", err)
	}
	if len(houses) != expected {
		T.Errorf("Wrong number of houses! expected %d but got %d", expected, len(houses))
	}
}

func TestTxn(T *testing.T) {
	//raw, mux := setupDispatcher()
	store := setupTestStore()

	obj := &testObj{}
	wrapped := QbsWrapAll(obj, store)

	for _, choice := range []int{force_panic, force_error} {
		obj.failPost = choice
		checkNumberHouses(T, store, 0)
		_, err := wrapped.Post(&HouseWire{Id: 0, Addr: "123 evergreen terrace", ZipCode: 98607}, nil)
		if err == nil {
			T.Fatalf("expected an error but didn't get one!")
		}
		e, ok := err.(*Error)
		if !ok {
			T.Fatalf("unexpected error type %T", err)
		}
		if e.StatusCode != http.StatusInternalServerError {
			T.Error("Wrong error code, expected %d but got %d", 500, e.StatusCode)
		}
		checkNumberHouses(T, store, 0)
	}
}

func setupDispatcher() (*RawDispatcher, *ServeMux) {

	io := NewRawIOHook(&JsonDecoder{}, &JsonEncoder{}, nil)
	raw := NewRawDispatcher(io, nil, nil, NewSimpleTypeHolder(), "/rest")

	serveMux := NewServeMux()
	serveMux.Dispatch("/rest/", raw)

	return raw, serveMux
}

func setupTestStore() *QbsStore {
	dsn := ParamsToDSN("seven5test", "", "")
	return NewQbsStore(dsn)
}

func checkNetworkCalls(T *testing.T, portSpec string, serveMux *ServeMux, obj *testObj) {
	if obj.testCallCount != 0 {
		T.Fatalf("sanity check at start failed: %d", obj.testCallCount)
	}

	go func() {
		http.ListenAndServe(portSpec, serveMux)
	}()

	client := new(http.Client)

	messageData := [][]string{
		[]string{"GET", fmt.Sprintf("http://localhost%s/rest/house", portSpec), ""},
		[]string{"GET", fmt.Sprintf("http://localhost%s/rest/house/1", portSpec), ""},
		[]string{"DELETE", fmt.Sprintf("http://localhost%s/rest/house/1", portSpec), ""},
		[]string{"POST", fmt.Sprintf("http://localhost%s/rest/house", portSpec), "{}"},
		[]string{"PUT", fmt.Sprintf("http://localhost%s/rest/house/1", portSpec), "{}"},
	}
	for i, callCount := range []int{1, 2, 3, 4, 5} {
		req := makeReq(T, messageData[i][0], messageData[i][1], messageData[i][2])
		resp, err := client.Do(req)
		checkResponse(T, err, resp)
		if obj.testCallCount != callCount {
			T.Fatalf("did not call QBS level resource (expected %d calls but found %d)", callCount, obj.testCallCount)
		}
	}

}

func TestWrappingAll(T *testing.T) {
	raw, mux := setupDispatcher()
	store := setupTestStore()

	obj := &testObj{}
	raw.Resource("house", &HouseWire{}, QbsWrapAll(obj, store))
	checkNetworkCalls(T, ":8991", mux, obj)

	store.Q.WhereEqual("Zip", 0).Delete(&House{})
}

func TestWrappingSeparate(T *testing.T) {
	raw, mux := setupDispatcher()
	store := setupTestStore()

	obj := &testObj{}

	raw.ResourceSeparate("house", &HouseWire{},
		QbsWrapIndex(obj, store),
		QbsWrapFind(obj, store),
		QbsWrapPost(obj, store),
		QbsWrapPut(obj, store),
		QbsWrapDelete(obj, store))

	checkNetworkCalls(T, ":8992", mux, obj)

	store.Q.WhereEqual("Zip", 0).Delete(&House{})

}

func checkResponse(T *testing.T, err error, resp *http.Response) {
	if err != nil {
		T.Fatalf("failed on %s with error %v", "GET", err)
	}
	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		T.Fatalf("failed on %s with status %d", "GET", resp.StatusCode)
	}
}
