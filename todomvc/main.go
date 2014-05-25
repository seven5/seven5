package main

import (
	// JQUERY: Jquery allows many non type-safe operations.
	// Uses of jquery are marked in the code as they are suspect.
	"github.com/gopherjs/jquery"
	//"honnef.co/go/js/console"
	s5 "github.com/seven5/seven5/client"
	"strings"
)

var (
	//These are static references to the dom. They each specify a
	//single element. Maybe should be all upper case?  Note that
	//it is preferred to access constrainable parts of the dom
	//that is fixed (in the page) through these objects.
	sectionMain    = s5.NewHtmlId("section", "main")
	footer         = s5.NewHtmlId("footer", "footer")
	primaryInput   = s5.NewHtmlId("input", "new-todo")
	listContainer  = s5.NewHtmlId("ul", "todo-list")
	pluralSpan     = s5.NewHtmlId("span", "plural")
	todoCount      = s5.NewHtmlId("span", "todo-count")
	clearCompleted = s5.NewHtmlId("button", "clear-completed")
	numCompleted   = s5.NewHtmlId("span", "num-completed")
	toggleAll      = s5.NewHtmlId("input", "toggle-all")

	//css classes used in the view
	view      = s5.NewCssClass("view")
	toggle    = s5.NewCssClass("toggle")
	destroy   = s5.NewCssClass("destroy")
	edit      = s5.NewCssClass("edit")
	completed = s5.NewCssClass("completed")

	//note: this is the same "name" as the field in the todo model but because
	//note: of types you can't bodge it up.  Having the same "name" with
	//note: different, but coordinated, types FTW!
	editing = s5.NewCssClass("editing")
)

const (
	CUTSET = " \n\t" //for trimming
)

////////////////////////////////////////////////////////////////////////
/// TODO MODEL: MODELS SINGLE TODO ITEM
////////////////////////////////////////////////////////////////////////

type todo struct {
	modName s5.ModelName
	name    s5.StringAttribute
	done    s5.BooleanAttribute
	editing s5.BooleanAttribute
}

//Id is used to decide which todo item is which.
func (self *todo) Id() string {
	return self.modName.Id()
}

//Equal is used to compare two todo items.  They are the same if they
//have the same Id.
func (self *todo) Equal(e s5.Equaler) bool {
	if e == nil {
		return false
	}
	other := e.(*todo)
	return self.Id() == other.Id()
}

//newTodo creates a new todo list item for a given string.  Note that the
//three primary elements are Attributes so the constraint system can operate
//on them.  The value provided is used to initialize the displayed text.
func newTodo(raw string) *todo {
	result := &todo{
		name:    s5.NewStringSimple(raw),
		done:    s5.NewBooleanSimple(false),
		editing: s5.NewBooleanSimple(false),
	}
	result.modName = s5.NewModelName(result)
	return result
}

////////////////////////////////////////////////////////////////////////
/// TODO APP
////////////////////////////////////////////////////////////////////////

//note: attributes at this level other than the main collection are here
//note: so that they can be constrained in the Start() method and then
//note: ignored for the rest of the program.  These attributes represent
//note: things that are state "above the level" of a single todo item.
type todoApp struct {
	//list of the todo elements, initially empty
	todos *s5.Collection

	//number of items in the list that are not currently done.
	numNotDone s5.IntegerAttribute
	//string that is either "" or "s" to make the plural of the number
	//of items on the display work out right.
	plural s5.StringAttribute
	//boolean that is true if there are some items in the list that
	//are done
	someDone s5.BooleanAttribute
	//number of elements that are done
	numDone s5.IntegerAttribute

	//true if this object is currently being edited
	editing s5.BooleanAttribute
}

//Start is called by concorde once the DOM is fully loaded.  Most of the
//application intialization code should be put in here.
func (self *todoApp) Start() {

	//Setup an event handler for the primary input field. The called func
	//creates model instance and puts in the list of todos.
	// JQUERY: Any use of jquery is suspect as it allows many non type-safe operations.
	primaryInput.Event(s5.CHANGE, func(j jquery.JQuery, event jquery.Event) {
		if !self.createTodo(j.Val()) {
			event.PreventDefault()
		}
	})

	//We need to attach the self.numNotDone string to the proper place
	//in the DOM.  Note that we use the lower-level jquery selector
	//(Select().Children())) plus NewTextAttr() we needed the string
	//child of span#todo-count (so we can't use HtmlId) and because
	//that object is directly in the HTML file.
	// JQUERY: Any use of jquery is suspect as it allows many non type-safe operations.
	todoCountSelect := todoCount.Select().Children("strong")
	s5.Equality(c.NewTextAttr(todoCountSelect), self.numNotDone)

	//We need to attach the self.plural string to the proper place
	//in the dom.
	s5.Equality(pluralSpan.TextAttribute(), self.plural)

	//These two calls attach the inverse of the empty attribute derived
	//from the model collection the display property (turning them on
	//when the list is not empty).  Note that we can't use the simpler
	//"Equality()" because we want to invert the value. The BooleanInverter
	//is a built in constraint function.
	sectionMain.DisplayAttribute().Attach(
		s5.NewBooleanInverter(self.todos.EmptyAttribute()))
	footer.DisplayAttribute().Attach(
		s5.NewBooleanInverter(self.todos.EmptyAttribute()))

	//This connects the display property of the clearCompleted to the boolean
	//that is true if some of the elements are done.  This effectively
	//turns on the button when there are some elements in the list and
	//some of those elements are done.
	s5.Equality(clearCompleted.DisplayAttribute(), self.someDone)

	//This connects the display in the button to the number of done
	//elements.  Note that this wont be visible if there are no
	//done elements.
	s5.Equality(numCompleted.TextAttribute(), self.numDone)

	//This is the event handler for click on the clearCompleted
	//dom element. We just walk the list of objects building a kill list,
	//then we destroy all the items in the kill list
	//JQUERY: Neither of the jquery params are used.
	clearCompleted.Event(c.CLICK, func(jquery.JQuery, jquery.Event) {
		all := self.todos.All()
		if len(all) == 0 {
			return
		}
		dead := make([]s5.Model, len(all))
		ct := 0
		for _, model := range all {
			if model.(*todo).done.Value() {
				dead[ct] = model
				ct++
			}
		}
		for _, d := range dead {
			self.todos.Remove(d)
		}
	})

	//toggleAll's behavior is to toggle any items that are not already
	//marked done, unless they are all marked done in which they should all
	//be umarked
	//JQUERY: Neither of the jquery params are used.
	toggleAll.Event(s5.CLICK, func(jquery.JQuery, jquery.Event) {
		desired := true
		//Compare the output of the constraints to see if all are done
		if self.todos.LengthAttribute().Value() == self.numDone.Value() {
			desired = false
		}
		for _, m := range self.todos.All() {
			m.(*todo).done.Set(desired)
		}
	})

	//These are discussed below. These are constraints that depend
	//on *all* the values in the list.
	self.dependsOnAll()
}

//called from the UI when the user hits return in the primary type-in
//field.  We return false if we want the input to be ignored.
func (self *todoApp) createTodo(v string) bool {
	result := strings.Trim(v, CUTSET)
	if result == "" {
		return false
	}
	primaryInput.Select().SetVal("")
	//note: we just push it into the list and let the constraint system
	//note: take over in terms of updating the display
	todo := newTodo(result)
	self.todos.PushRaw(todo)
	return true
}

//newApp creates a new instance of the application object, properly
//initialized
func newApp() *todoApp {
	result := &todoApp{}
	//init the list, setting our own object as the joiner (we meet
	//the interface Joiner)
	result.todos = s5.NewList(result)

	//create initial values of attributes
	result.numNotDone = s5.NewIntegerSimple(0)
	result.plural = s5.NewStringSimple("")
	result.someDone = s5.NewBooleanSimple(false)
	result.numDone = s5.NewIntegerSimple(0)

	//done create app object
	return result
}

//helper function for getting the done attribute out of our model
func (self *todoApp) pullDone(m c.Model) s5.Attribute {
	return m.(*todo).done
}

//This function is called from Start(). It creates the constraints that
//are functions of *all* the list elements.  Note that these are not
//used at startup because there are no list elements to depend on yet.
//The initial values (above) of the attributes are the startup values.
func (self *todoApp) dependsOnAll() {

	//
	// NUMBER OF NOT DONE ELEMENTS
	//
	self.todos.AllFold(
		self.numNotDone, //object to be constrained
		0,               //initial value for iterative folding
		self.pullDone,   //done attribute from the model

		//this function is called repeatedly with the *values* extracted
		//from the done attributes in our list. It sums the number of
		//elements that are not done. It holds state in the first return
		//param, the second is the final result on the last iter.
		func(prev interface{}, curr s5.Equaler) (interface{}, s5.Equaler) {
			p := prev.(int)
			if !curr.(s5.BoolEqualer).B {
				p++
			}
			return p, s5.IntEqualer{p}
		},
		c.IntEqualer{0}, //if we transition to an empty list, what result do we want?
	)

	//
	// STRING FOR PLURALIZATION OF THE NUMBER DONE
	//
	self.todos.AllFold(
		self.plural,   //object to be constrained
		0,             //initial value for iterative folding
		self.pullDone, //operating on the done attribute in the model

		//"s" if there is exactly one element not done, otherwise ""
		func(prev interface{}, curr s5.Equaler) (interface{}, s5.Equaler) {
			p := prev.(int)
			if !curr.(s5.BoolEqualer).B {
				p++
			}
			s := "s"
			if p == 1 {
				s = ""
			}
			return p, s5.StringEqualer{s}
		},
		s5.StringEqualer{"s"}, // what to show if we transititon to empty list
	)

	//
	// ARE ANY TODO ITEMS DONE?
	//
	self.todos.AllFold(
		self.someDone, //object to be constrained
		0,             //initial value for iterative folding
		self.pullDone, //done attribute in the model

		//return true if there is at least one done item
		func(prev interface{}, curr s5.Equaler) (interface{}, s5.Equaler) {
			p := prev.(int)
			if curr.(s5.BoolEqualer).B {
				p++
			}
			return p, s5.BoolEqualer{p > 0}
		},
		s5.BoolEqualer{false}, //what to do if we transition to empty list
	)

	//
	// NUMBER OF DONE ITEMS
	//
	self.todos.AllFold(
		self.numDone,  //object to be constrained
		0,             //initial value for iterative folding
		self.pullDone, //we operate on the done field

		//total up the number of items that are marked as done
		func(prev interface{}, curr s5.Equaler) (interface{}, s5.Equaler) {
			p := prev.(int)
			if curr.(s5.BoolEqualer).B {
				p++
			}
			return p, s5.IntEqualer{p}
		},
		s5.IntEqualer{0}, //result if we transition to an empty list of todos
	)
}

//Add() is the method that is called in response to an element being
//added to the collection (self.todos).  This is the "magic" turns an
//instance of todo into a view.
func (self *todoApp) Add(length int, newObj c.Model) {
	model := newObj.(*todo)

	//note: There are two legal things that can be passed to any
	//note: of the tag creation methods.  Sadly, there is no way
	//note: to typecheck these until runtime (it is checked then).
	//note:
	//note: The two legal things are some type of option, such as
	//note: Class() or ModelId() that affect the resulting tag.
	//note: One of the common types of "options" is something creates
	//note: a constraint between a dom "piece" of the tag being
	//note: being constructed and some value, usually in the model.
	//note:
	//note: The other leagl item is another tag, that will be added
	//note: as a child of the one it is neted in.   This lack of type
	//note: safety has been chosen for convenience of notation.
	tree :=
		s5.LI(
			//LI: Pass in a model ID to generate unique id for this tag,
			//LI: and make easy to remove the whole subtree by id.
			s5.ModelId(model),
			//LI: constraint that toggles the completed property
			s5.CssExistence(completed, model.done),
			//LI: constraint that toggles the editing property
			s5.CssExistence(editing, model.editing),
			s5.DIV(
				//DIV: just one CSS class to make it look nice
				s5.Class(view),
				s5.INPUT(
					//INPUT: has a CSS class "toggle" to make it look nice
					s5.Class(toggle),
					//INPUT: we force the "type" of this to be the constant "checkbox" (possibly overkill)
					s5.HtmlAttrEqual(s5.TYPE, c.NewStringSimple("checkbox")),
					//INPUT: make the checked attr be equal to the model's done
					s5.PropEqual(s5.CHECKED, model.done),
					//INPUT: when clicked, it toggles the value on the model
					s5.Event(s5.CHANGE, func(ignored jquery.JQuery, e jquery.Event) {
						model.done.Set(!model.done.Value())
					}),
				),
				s5.LABEL(
					//LABEL: We use a constraint to bind the name attribute of the
					//LABEL: model to the label's displayed text.
					s5.TextEqual(model.name),
					//LABEL: Double clicking on the label causes edit mode
					s5.Event(s5.DBLCLICK, func(ignored jquery.JQuery, ignored2 jquery.Event) {
						model.editing.Set(true)
						//XXX UGH, don't have a handle to the input object
						in := s5.HtmlIdFromModel("INPUT", model).Select()
						in.SetVal(model.name.Value())
						in.Focus()
						in.Select()
					}),
				),
				s5.BUTTON(
					//BUTTON: destroy class makes it look nice
					s5.Class(destroy),
					//BUTTON: click function that calls remove on the list
					//BUTTON: element that was used to create this whole structure
					//JQUERY: Neither of the jquery params are used.
					s5.Event(s5.CLICK, func(jquery.JQuery, jquery.Event) {
						//note: we are calling remove on the *collection* which
						//note: will end up calling the Remove() method of our
						//note: joiner.  If we don't tell the collection that the
						//note: model was removed, we could end up with a display
						//note: that doesn't correctly reflect the constraints
						//note: state (since these constraints would have dependencies
						//note: one items no longer visible).
						self.todos.Remove(model)
					}),
				), //BUTTON
			), //DIV
			s5.INPUT(
				//INPUT: Use a model to make this input easy to find
				s5.ModelId(model),
				//INPUT: edit CSS class to make it look nice
				s5.Class(edit),
				//INPUT: wire the placeholder to be name of the model... (overkill?)
				s5.HtmlAttrEqual(s5.PLACEHOLDER, model.name),
				//INPUT:the spec calls for escape to cancel editing with no change
				//INPUT:and for return to commit the changes EXCEPT if the
				//INPUT:user edited out all the text, then we should delete the
				//INPUT:whole thing
				//JQUERY: This uses the jquery selector to get the value of the input.
				//JQUERY: This uses the event object to get the keyboard code.
				s5.Event(c.KEYDOWN, func(j jquery.JQuery, e jquery.Event) {
					//note: This type of "event handler" is the glue that
					//note: connects a user action to something that manipulates
					//note: the model.  Most event handlers do not need to
					//note: manipulate the view as well (although that can be done
					//note: through the j parameter) because they have constraints
					//note: that connect the model to the view.  This event
					//note: handler touches the view (j) primarily because it needs
					//note: to manipulate the focus, which is not expressed
					//note: in constraints.
					switch e.Which {
					case 13:
						v := strings.Trim(j.Val(), CUTSET)
						//check for the special case of making a name==""
						if v == "" {
							self.todos.Remove(model)
						} else {
							//just reset the model name and it propagates to display
							model.name.Set(v)
						}
						j.Blur()
						fallthrough
					case 27:
						model.editing.Set(false)
						primaryInput.Select().Focus()
					}
				}),
			), //INPUT
		).Build()

	listContainer.Select().Append(tree)
}

//Remove is called when the oldObj is removed from the collection. It
//just looks up the view (via the id of the model) and then removes it
//from the display.
func (self *todoApp) Remove(IGNORED int, oldObj s5.Model) {
	model := oldObj.(*todo)
	finder := s5.HtmlIdFromModel("li", model)
	finder.Select().Remove()
}

//Go-level entry point, normally code is put in the Start() method that
//needs to manipulate the UI.  This is called _before_ the DOM is fully
//loaded.
func main() {
	s5.Main(newApp())
}
