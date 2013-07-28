import 'dart:async';
import 'dart:html';
import 'package:polymer/polymer.dart';
import 'package:observe/observe.dart';
import 'package:nullblog/src/nullblog.dart';


//This is the true implementation of the code that does the work for displaying an article
class ArticleDiv extends PolymerElement with ObservableMixin {
  
	//pulled from the server
  //@observable article obj;
	
	static final String rawHtml = '''
  <template id="article-div">
      <div class="article-div-main">  <!--see article.css-->
        <hr/>
        <p class="lead">{{ Content }}</p>
        <h4 class="author">Written By {{ Author }} -- (Id: {{ Id }})</h4>
      </div>
  </template>        
	''';
	
	static final Element htmlContent = new Element.html(rawHtml);
	
	static final Element invocation = new Element.html("<template id='invoke-article-div' ref='article-div' syntax='fancy' bind>");
	
}