#library('modena');
#source("munger.dart");

/**
 * A Munger knows how to turn some json into HTML.  Some Mungers take an amount of time
 * to run because they do animations or other actions that are not instantaneous.  
 */
interface Munger {
	/**
	  * Munger gets passed a wad of json (loaded from the server) plus a root element of the DOM
	  * to place it's content into.  The root element is always empty but is alreayd correctly sized. 
	  */
	Munger(Map<String,Object> jsonData, DivElement root);
	/**
	  * Animated Mungers or mungers which display lots of text, and thus take a while to read, should
	  * return a value here in milliseconds that they would prefer to have their display up.
	  */
	int preferredRunTime();
	/**
	  * The run method is called with a time (in ms) that is the total run time of the munger,
	  * the "current start time" (inclusive) of the interval of time being represented by this
	  * segment and an end point.  In math terms, the interval is [start, endExclusive).  A munger
	  * can and should assume that it will receive at least two calls to run one with the start
	  * time of zero (ms) and the other with the start time equal to total time.  Thus, a silly
	  * munger that returns a preferredRunTime of 50ms might get a sequence of calls like this
	  * run(50,0,50)  --- covers the whole time period except the shutdown moment
	  * run(50,50,50) --- shutdown message
	  *
    * A more sensible example might be something like:
    * run(10000,0,250)
    * run(10000,250, 500)
    * ... etc at 250ms intervals until ...
    * run(10000,10000,10000)
    *
    * There is never any interval of the time period that is not covered by a call to run, although
    * it may represent 0ms.  Intervals are not guaranteed to be equally spaced.  Munger authors should
    * react to the interval presented, not anticipate future intervals.
    *
    * Mungers should clean up their DOM nodes at the point of receiving the last interval, run(x,x,x).
    */
	void run(int totalTime, int start, int endExclusive);
}

/**
  * The driver class handles managing mungers plus timing them so that they can be bundled off stage when
  * they are done.
  */
class Driver { //heh: ferrari, maseratti...
	const int INTERVAL = 250;//ms
	/**
	  * Given a list of json objects, display with some mungers.  
	  */
	void runWithContent(List<Map<String,Object>> jsonArray) {
		DivElement root = document.query("root");
		
		for (Map<String,Object> jsonData in jsonArray) {
			QuoteMunger munger = new QuoteMunger(jsonData,root);
			int length = munger.preferredLength();
			if (length<1000) {
				stupidMunger(length,munger);
				continue;
			}
			launchMunger()
		}
	}
}