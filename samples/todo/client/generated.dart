#library('generated');
#import('seven5.dart');
#import('dart:html');
#import('jsonobject.dart');

class UrgencyLevel {
  JsonObject self;
  int get Id() => self["Id"];
  String get Name() => self["Name"];
  int get Level() => self["Level"];
  UrgencyLevel.fromJsonString(String s) {
    self=JsonObject.fromJsonString(s);
    self["isExtendable"] = false;
  }
}

interface Todo extends JsonObject{
  String Description;
  UrgencyLevel Urgency; 
  bool Done;
}

String restfulTextBlob = 
'''
<p>Everything that has 'state' in this application has a Rest backend 
implemented in Go. This is not just the ToDo items, but also the sets of names,
or <em>vocabularies</em>, such as the levels of importance.
''';

String strongTypingTextBlob =
'''
All the CSS classes and id values on the Html elements are fully cross-checked 
and typed to prevent errors.  All of these are defined <em>once</em>, in your 
Go code, and then automatically connected to the Dart front-end code.
''';

String devCycleTextBlob =
'''
Whether you are developing front-end or back-end code, the development cycle is 
always the same.  Just hit 'Reload' in your browser and everything that has 
changed gets rebuilt.
''';


class CssRule {
  static Map<String,Element> _idCache;
  static Element _byTagAndId(String query) {
    if (!_idCache.containsKey(query)) {
      _idCache[query] = document.query(query);
    }
    return _idCache[query];
  }
  static void _init() {
    if (_idCache == null) {
      _idCache = new Map<String,Element>();
    }
  }
  static Element get expContainer() {
    _init();
    return CssRule._byTagAndId("div#explanationContainer");
  }
  static Element get urgentMarker() {
    _init();
    return CssRule._byTagAndId("li#urgentMarker");
  }
  static Element get doneMarker() {
    _init();
    return CssRule._byTagAndId("li#doneMarker");
  }
}

