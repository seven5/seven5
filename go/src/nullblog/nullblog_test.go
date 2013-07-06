package nullblog

import (
	"net/http"
	"testing"
	"github.com/seven5/seven5"
)

//trivial test of the Index method
func TestIndex(T *testing.T) {
	underTest := new(ArticleResource)

	result, err := underTest.Index(nil)
	if err != nil {
		T.Fatalf("unexpected error in Index: %s", err)
	}
	articles := result.([]*ArticleWire)
	if len(articles) != 2 {
		T.Fatalf("unexpected size of articles returned from Index: %d", len(articles))
	}
}

func TestFind(T *testing.T) {

	underTest := new(ArticleResource)

	result, err := underTest.Find(0, nil)
	if err != nil {
		T.Fatalf("unexpected error in Find: %s", err)
	}
	article := result.(*ArticleWire)
	if article.Id != 0 || article.Author == "" || article.Content == "" {
		T.Fatalf("unexpected content in Find %+v", article)
	}

	result, err = underTest.Find(2890, nil)
	if result != nil {
		T.Fatalf("expected error but got result in Find %+v", result)
	}
	if err == nil {
		T.Fatalf("nothing returned from Find (two nils)!")
	}

	http_error := err.(*seven5.Error)
	if http_error.StatusCode != http.StatusBadRequest {
		T.Errorf("wrong error code, expect BadRequest but got %d", http_error.StatusCode)
	}

}
