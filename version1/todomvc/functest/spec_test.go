package unittest

import (
	selenium "github.com/sourcegraph/go-selenium"
	"strings"
	"testing"
)

var (
	caps      selenium.Capabilities = make(selenium.Capabilities)
	serverURL                       = "http://localhost:8898/index.html"
	executor                        = "http://localhost:4444"
)

func TestPageLoad(t *testing.T) {
	wd := newRemote("PageLoad", t)
	defer wd.Quit()

	wd.Get(serverURL)
	fetchedURL := wd.CurrentURL()

	short := serverURL[0 : strings.LastIndex(serverURL, "/")+1]
	if fetchedURL != short {
		t.Fatalf("URL %s != %s", fetchedURL, short)
	}

	checkMainSectionsVisible(t, wd, false, "when the page loads")

}

func TestEnterNewTodo(t *testing.T) {
	wd := newRemote("EnterNewTodo", t)
	defer wd.Quit()

	wd.Get(serverURL)
	primaryInput := wd.Q("#new-todo")
	primaryInput.SendKeys("fleazil\n")

	if primaryInput.Text() != "" {
		t.Fatalf("input was not cleared after creating todo!")
	}
	checkMainSectionsVisible(t, wd, true, "when an item has been added")

	checkNumTodoItemsDisplayedText(t, wd, "1 item")
}

func TestMarkDone(t *testing.T) {
	wd := newRemote("MarkDone", t)
	defer wd.Quit()

	wd.Get(serverURL)
	primaryInput := wd.Q("#new-todo")
	primaryInput.SendKeys("frobnitz\n")

	//test structure with complex query
	div := wd.Q("ul#todo-list li#li-todo-0 div.view")

	label := div.Q("label")
	if label.Text() != "frobnitz" {
		t.Fatalf("unexpected text in todo item: '%s'", label.Text())
	}

	if numCheckedItems(t, wd) != 0 {
		t.Fatalf("no items should be marked completed with item just added")
	}

	toggle := div.Q("input.toggle")
	toggle.Click()

	checkNumTodoItemsDisplayedText(t, wd, "0 items")

	if numCheckedItems(t, wd) != 1 {
		t.Fatalf("clicked an item, but it was not marked done")
	}

	checkNumCompletedItemsDisplayedText(t, wd, "1")

}

func TestClearCompleted(t *testing.T) {
	wd := newRemote("ClearCompleted", t)
	defer wd.Quit()

	wd.Get(serverURL)
	primaryInput := wd.Q("#new-todo")
	primaryInput.SendKeys("bar\n")
	primaryInput.SendKeys("baz\n")

	//sanity
	checkNumTodoItemsDisplayedText(t, wd, "2 items")

	//click em both to done
	toggles := wd.QAll("input.toggle")
	for _, toggle := range toggles {
		toggle.Click()
	}

	//sanity
	checkNumCompletedItemsDisplayedText(t, wd, "2")

	//click the clear all done
	button := wd.Q("button#clear-completed")
	button.Click()

	checkMainSectionsVisible(t, wd, false, "after clearing all done items")

}

func TestMarkAllDone(t *testing.T) {
	wd := newRemote("ClearCompleted", t)
	defer wd.Quit()

	wd.Get(serverURL)
	primaryInput := wd.Q("#new-todo")
	primaryInput.SendKeys("bar\n")
	primaryInput.SendKeys("baz\n")

	wd.Q("input.toggle").Click() //pick one and click it

	checkNumCompletedItemsDisplayedText(t, wd, "1")
	checkNumTodoItemsDisplayedText(t, wd, "1 item")

	//click the mark all done button
	wd.Q("input#toggle-all").Click()
	if numCheckedItems(t, wd) != 2 {
		t.Fatalf("failed to check the other item via the mark all done button")
	}

	//second click puts them all to not done
	wd.Q("input#toggle-all").Click()
	if numCheckedItems(t, wd) != 0 {
		t.Fatalf("failed to uncheck all the items when mark all done hit and all items checked")
	}

	if wd.Q("button#clear-completed").IsDisplayed() {
		t.Fatalf("after making all the items not done, clear done items button is still visible")
	}
}

///////////////////////////////////////////////////////////////////////
//// HELPERS
///////////////////////////////////////////////////////////////////////

func newRemote(testName string, t *testing.T) selenium.WebDriverT {
	var err error
	caps["browserName"] = "chrome"
	wd, err := selenium.NewRemote(caps, executor)
	if err != nil {
		t.Fatalf("can't start session for test %s: %s", testName, err)
	}
	return wd.T(t)
}

func checkNumTodoItemsDisplayedText(t *testing.T, wd selenium.WebDriverT, expected string) {
	countDisplayed := wd.FindElement(selenium.ByCSSSelector, "#todo-count").Text()
	if countDisplayed != expected {
		t.Fatalf("unexpected todo count '%s' but expected '%s'", countDisplayed, expected)
	}
}

func checkMainSectionsVisible(t *testing.T, wd selenium.WebDriverT, isVis bool, text string) {

	for _, hidden := range []string{"#footer", "#main"} {
		element := wd.Q(hidden)
		if element.IsDisplayed() != isVis {
			msg := "not be"
			if isVis {
				msg = "be"
			}
			t.Fatalf("%s section should %s turned on %s", hidden, msg, text)
		}
	}

}

func numCheckedItems(t *testing.T, wd selenium.WebDriverT) int {
	return len(wd.QAll("li.completed"))
}

func checkNumCompletedItemsDisplayedText(t *testing.T, wd selenium.WebDriverT, expected string) {
	countDisplayed := wd.Q("span#num-completed")
	if countDisplayed.Text() != expected {
		t.Fatalf("unexpected done count '%s' but expected '%s'", countDisplayed, expected)
	}
}
