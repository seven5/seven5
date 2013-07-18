import 'dart:io';
import 'package:polymer/component_build.dart';

// Ref: http://www.dartlang.org/articles/dart-web-components/tools.html
main() {
  build(new Options().arguments, ['web/article.html']);
}
