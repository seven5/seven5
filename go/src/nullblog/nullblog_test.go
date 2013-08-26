package nullblog

import (
	"bytes"
	_ "github.com/lib/pq"
	"github.com/seven5/seven5"
	"net/http"
	"testing"
)

func checkIndexLength(T *testing.T, underTest seven5.RestIndex, expected int) []*ArticleWire {
	raw, err := underTest.Index(nil)
	if err != nil {
		T.Fatalf("unexpected error in Index: %s", err)
	}
	result := raw.([]*ArticleWire)
	if len(result) != expected {
		T.Errorf("unexpected number of results, expected %d but got %d", expected, len(result))
	}
	return result
}

func TestResourceIndexQbs(T *testing.T) {

	store := seven5.NewLocalhostEnvironment("nullblog", true /*test*/).GetQbsStore()
	underTest := seven5.QbsWrapIndex(&ArticleResource{}, store)

	seven5.WithEmptyQbsStore(store, &Migrate{}, func() {
		checkIndexLength(T, underTest, 0)

		//put content in DB by hand
		a := &Article{}
		a.Author = "Ian Smith"
		var b bytes.Buffer
		for i := 0; i < 1000; i++ {
			b.WriteString("this is a test.")
		}
		a.Content = b.String()

		if _, err := store.Q.Save(a); err != nil {
			T.Fatalf("unexpected error saving content (%s)", err)
		}

		//test that content is returned
		checkIndexLength(T, underTest, 1)

		//add more content
		a = &Article{}
		a.Author = "Joe Blow"
		a.Content = "Bad Web Frameworks... blah blah blah"

		if _, err := store.Q.Save(a); err != nil {
			T.Fatalf("unexpected error saving content (%s)", err)
		}

		//test that content is returned
		checkIndexLength(T, underTest, 2)

	})
}

func TestResourceFindQbs(T *testing.T) {

	store := seven5.NewLocalhostEnvironment("nullblog", true /*test*/).GetQbsStore()
	underTest := seven5.QbsWrapFind(&ArticleResource{}, store)

	seven5.WithEmptyQbsStore(store, &Migrate{}, func() {
		//NOTE: no connection to the database here! this is the code under test!
		result, err := underTest.Find(-22, nil)
		if err == nil {
			T.Fatalf("expected error but didn't receive one from Find!")
		}

		http_error := err.(*seven5.Error)
		if http_error.StatusCode != http.StatusBadRequest {
			T.Errorf("wrong error code, expect BadRequest but got %d", http_error.StatusCode)
		}

		//write some content in the db
		a := &Article{}
		ian := "Ian Smith"
		shirt := "a bit shirty"
		a.Author = ian
		a.Content = shirt

		result_id, err := store.Q.Save(a)
		if err != nil {
			T.Fatalf("unexpected error in Save (%s)", err)
		}

		//NOTE: no connection to the database here! this is the code under test!
		result, err = underTest.Find(seven5.Id(result_id), nil)
		if err != nil {
			T.Fatalf("unexpected error in Find (%s)", err)
		}
		wire := result.(*ArticleWire)
		if string(wire.Author) != ian || string(wire.Content) != shirt {
			T.Errorf("wrong content received from Find (%+v)")
		}

		result, err = underTest.Find(seven5.Id(result_id-1), nil)
		if err == nil {
			T.Fatalf("expected error but didn't get it from Find")
		}

		http_error = err.(*seven5.Error)
		if http_error.StatusCode != http.StatusBadRequest {
			T.Errorf("wrong error code, expect BadRequest but got %d", http_error.StatusCode)
		}

	})

}
