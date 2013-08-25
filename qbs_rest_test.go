package seven5

import (
	"fmt"
	"github.com/coocood/qbs"
	_ "github.com/lib/pq"
	"net/http"
	"os"
	"strings"
	"testing"
)

/*---- type of actual DB action ----*/
type House struct {
	Address string
	Zip     int /*0->99999, inclusive*/
}

/*---- wire type for the tests ----*/
type HouseWire struct {
	Id Id
}

type testObj struct {
	testCallCount int
}

/*these funcs are use to test that if you meet the QBSRest interfaces you
can be wrapped by the qbs code in Seven5 */
func (self *testObj) Index(pb PBundle, q *qbs.Qbs) (interface{}, error) {
	self.testCallCount++
	return &HouseWire{}, nil
}
func (self *testObj) Find(id Id, pb PBundle, q *qbs.Qbs) (interface{}, error) {
	self.testCallCount++
	return &HouseWire{}, nil
}
func (self *testObj) Delete(id Id, pb PBundle, q *qbs.Qbs) (interface{}, error) {
	self.testCallCount++
	return &HouseWire{}, nil
}
func (self *testObj) Put(id Id, value interface{}, pb PBundle, q *qbs.Qbs) (interface{}, error) {
	self.testCallCount++
	return &HouseWire{}, nil
}
func (self *testObj) Post(value interface{}, pb PBundle, q *qbs.Qbs) (interface{}, error) {
	self.testCallCount++
	return &HouseWire{}, nil
}

/*-------------------------------------------------------------------------*/
/*                                 TEST CODE                               */
/*-------------------------------------------------------------------------*/
const (
	APP_NAME = "testapp"
)

func TestTxn(T *testing.T) {
}

func setupDispatcher() (*RawDispatcher, *ServeMux) {

	io := NewRawIOHook(&JsonDecoder{}, &JsonEncoder{}, nil)
	raw := NewRawDispatcher(io, nil, nil, NewSimpleTypeHolder(), "/rest")

	serveMux := NewServeMux()
	serveMux.Dispatch("/rest/", raw)

	return raw, serveMux
}

func setupTestStore(name string) *QbsStore {
	os.Setenv(strings.ToUpper(name)+"_DBNAME", "seven5test") //so you don't need it to run tests
	env := NewEnvironmentVars(name)
	return NewQbsStore(env)
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
	store := setupTestStore(APP_NAME)

	obj := &testObj{}
	raw.Resource("house", &HouseWire{}, QbsWrapAll(obj, store))
	checkNetworkCalls(T, ":8991", mux, obj)

}

func TestWrappingSeparate(T *testing.T) {
	raw, mux := setupDispatcher()
	store := setupTestStore(APP_NAME)

	obj := &testObj{}

	raw.ResourceSeparate("house", &HouseWire{},
		QbsWrapIndex(obj, store),
		QbsWrapFind(obj, store),
		QbsWrapPost(obj, store),
		QbsWrapPut(obj, store),
		QbsWrapDelete(obj, store))

	checkNetworkCalls(T, ":8992", mux, obj)

}

func checkResponse(T *testing.T, err error, resp *http.Response) {
	if err != nil {
		T.Fatalf("failed on %s with error %v", "GET", err)
	}
	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		T.Fatalf("failed on %s with status %d", "GET", resp.StatusCode)
	}
}
