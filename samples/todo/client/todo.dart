#import('dart:html');
#import('generated.dart');
#import('jsonobject.dart');

void main() {
  
  addExplanation("Restful", restfulTextBlob);
  addExplanation("Strong Typing", strongTypingTextBlob);
  addExplanation("Fast Dev Cycle", devCycleTextBlob);

  //Fake data
  UrgencyLevel low = new UrgencyLevel.fromJsonString('''{"Id": 7989, "Name":"Low", "Level": 1}''');
    
  
  
  //UrgencyLevel low = obj;
  UrgencyLevel med = {"Id": 7990, "Name":"Normal", "Level": 2};
  UrgencyLevel high = {"Id": 7991, "Name":"HOLY COW", "Level": 3};
  
  Todo todo1 = {"Description" : "Buy the latest album by Death Cab For Cutie, Numbers.",
                "Urgency": med, "Done": false, "Id": 1108};
  Todo todo2 = {"Description" : "Schedule visit to dentist.",
                "Urgency": low, "Done": false, "Id": 1109};
  Todo todo3 = {"Description" : "Check on toilet paper supply.",
                "Urgency": high, "Done": false, "Id": 1110};
  Todo todo4 = {"Description" : "Get dog food (special kind, for irritable, fussy little dogs).",
                "Urgency": med, "Done": true, "Id": 1110};
  
  //should be doing a query here for Todo.Done = false, Urgency.Level==3
  List<Todo> urgentNotDone = [ todo3 ];
  
  //should be doing a query here for Todo.Done = true
  List<Todo> done = [todo4];
  
  //should be doing a query here for Todo.Done = false
  List<Todo> list = [todo1, todo2, todo3];
  
  populateSideBar(done, urgentNotDone);
}

/**
  * This method removes any little links to things that exist already
  * and then replaces them with links to these objects.
  */
void populateSideBar(List<Todo> done, List<Todo> urgent) {
  Element urgentChild = CssRule.urgentMarker;
  Element doneChild = CssRule.doneMarker;
  Element parent = urgentChild.parent;
  num urgentIndex = parent.nodes.indexOf(urgentChild);
  num doneIndex = parent.nodes.indexOf(doneChild);

  print("urgent index ${urgentIndex} and ${doneIndex}");
  //the fact that we have to know which way the nodes are organized is bad
  parent.nodes.removeRange(urgentIndex+1, doneIndex-(urgentIndex+1));
  parent.nodes.removeRange(doneIndex+1, parent.nodes.length-(doneIndex+1));

  //create li's for our new elements
  List<LIElement> urgentElem = urgent.map((t) => liElementForTodo(t));
  List<LIElement> doneElem = done.map((t) => liElementForTodo(t));
  
  parent.nodes.setRange(urgentIndex+1, urgentElem.length, urgentElem);
  parent.nodes.addAll(doneElem);
}

/**
  * Compute the <li> to use for a given todo.  
  */
LIElement liElementForTodo(Todo t) {
  LIElement li = new LIElement();
  li.text = t.Description;
  return li;
}

/**
  * Add a textual explanation area down below the main playing field.
  */
void addExplanation(String title, String blobOfText) {
  
 
  DivElement column = new DivElement();
  column.attributes["class"] = "span4";

  Element h =  new Element.tag("h3");
  h.text = title;
  
  ParagraphElement p = new ParagraphElement();
  p.innerHTML = blobOfText;
 
  column.nodes.add(h);
  column.nodes.add(p);
  
  // note that this is typed because expContainer is generated
  CssRule.expContainer.nodes.add(column);
  
  
}
