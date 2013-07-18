import 'dart:html';
import 'package:unittest/mock.dart';
import 'package:nullblog/src/workarounds.dart';

class MockElement extends Mock implements Element{}
class MockCssClassSet extends Mock implements CssClassSet{}
class MockDocument extends Mock implements Document{}
class MockWindow extends Mock implements Window{}
class MockLocation extends Mock implements Location{}
class MockEvent extends Mock implements Event{}


MockDocument prepMockDocument(MockWindow window) {
  MockDocument result = new MockDocument();
  window.when(callsTo('doc')).alwaysReturn(result);
  return result;
}

MockLocation prepMockLocation(MockWindow window) {
  MockLocation result = new MockLocation();
  window.when(callsTo('loc')).alwaysReturn(result);
  return result;
}

MockElement prepMockQuery(MockDocument doc, String query){
  MockElement elem = new MockElement();
  doc.when(callsTo('query', query)).alwaysReturn(elem);
  return elem;  
}

MockCssClassSet prepMockCssSet(MockElement elem){
  MockCssClassSet cl = new MockCssClassSet();
  elem.when(callsTo('classSet')).alwaysReturn(cl);
  return cl;
}
