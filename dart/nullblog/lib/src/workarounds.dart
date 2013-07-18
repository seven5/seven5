import 'dart:html';

//
// All of these classes are here because we can't get properties to work
// "right" in mock objects so we have do wrappers and create functions
// that are called by the implementation.
//


class MyDocument implements Document {
  Document wrapped;
  MyDocument(Document w) {
    wrapped=w;
  }
  Element query(String selectors) {
    return new MyElement(wrapped.query(selectors));
  }
}

class MyLocation implements Location {
  Location wrapped;
  MyLocation(Location w) {
    wrapped=w;
  }
  String fullUri() {
    return wrapped.href;
  }
  void replace(String url) {
    wrapped.replace(url);
  }
  void goto(String url) {
    wrapped.href = url;
  }
  Map<String,String> queryHash() {
    Uri uri = Uri.parse(wrapped.href);
    if (uri.query=="") {
      return {};
    }
    List<String> raw = uri.query.split("&");
    Object pieces = raw.map((e) => e.split("=")); 
    Map<String,String> params = {}; 
    pieces.forEach((List<String> piece) => params[piece[0]] = piece[1]);  
    return params;
  }
  void noSuchMethod(Invocation mirror) {
    print('You tried to use a non-existent member of MyLocation: ${mirror.memberName}');
  }
}

class MyWindow implements Window {
  Window wrapped;
  MyWindow(Window w) {
    wrapped=w;
  }
  Document doc() {
    return new MyDocument(wrapped.document);
  }
  Location loc() {
    return new MyLocation(wrapped.location);
  }
  void noSuchMethod(Invocation mirror) {
    print('You tried to use a non-existent member of MyWindow: ${mirror.memberName}');
  }

}
class MyElement implements Element {
  Element wrapped;
  MyElement(Element w) {
    wrapped=w;
  }
  CssClassSet classSet() {
    return wrapped.classes;
  }
  void noSuchMethod(Invocation mirror) {
    print('You tried to use a non-existent member of MyElement: ${mirror.memberName}');
  }
}
