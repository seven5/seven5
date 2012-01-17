# Seven5: Pontificate

<nav>
    <ul>
        <li>[Intro](index.html)</li>
        <li>[Install](install.html)</li>
        <li>[Develop](develop.html)</li>
        <li>[Pontificate](pontificate.html)</li>
    </ul>
</nav>

Here you will find snippets of text explaining why Seven5 is the cranky, [opinionated](http://gettingreal.37signals.com/ch04_Make_Opinionated_Software.php), single-minded beast that it is.  If this were Halo, Seven5 would be the father of three with the sniper rifle making life miserable for the sword wielding 12 year olds camping by the lifts.

## Convention, not configuration

Seven5 uses conventions, not configuration, to manage how your web application will work.  Because we've chosen one way to code and one way to deploy, most of the configuration jackassery goes away.  During development you will never write a configuration file.  During deployment you will only have to place your creds in predefined places and Bob's your uncle.

Conventions also allow us to stop the insanity of repeating ourselves in each layer of the stack.  So many web frameworks think that it's normal to code your data models in multiple places: ORM, API and Javascript.  Seven5 unifies the definition of data, API, events and the rest of the system and then flows that information out where necessary.

If you want flexibility, go write Apache configuration files.


## Data structures belong in RAM, not on disk

We avoid disk storage for dynamic data because disk is slow, expensive to manage, and generally more trouble than it's worth in the age of server machines with metric tons of RAM.  Also, it's a fricking spinning platter of metal.  WTF?!? 

97% of the world doesn't need relational databases and the other 1% screws it up.  Because of snazzy ORMs that hide SQL, developers end up with relational databases simply because that's the easiest path in the web framework.

Seven5 makes a different strategy easy.  Your data model is just, well, your data model.  You use go's data structures and store them, unmodified, in a distributed flock of RAM called [memcached](http://memcached.org/) using [gobs](http://blog.golang.org/2011/03/gobs-of-data.html) because [Rob Pike](http://en.wikipedia.org/wiki/Rob_Pike) is a righteous dude.

If you're paranoid about your entire cluster crashing then rest assured, it's trivial to provision redundant memcacheds in different locales.  In the future, we'll get around to [hacking up something](http://code.google.com/p/leveldb-go/) that uses the spinning platters of metal as a deep backup.

## Design is for designers. Programming is for programmers.

There was a time when the base materials of the web were simple enough that the "front end" of a web app could be produced by people who did not understand the "back end".  This seemed like a good idea because real programmers were snobbish about crappy browsers and foisting off quirks mode junk seemed like a good idea.

Those days are dead and buried.

Today's browsers are more standardized so the pain of quirks is fading.  Today's standards are more complex and most designers would rather design than learn about CSS selectors and event binding.  Today's pages are programs and client side libraries like jQuery and YUI require honest to god programming.

Seven5 embraces this by unifying the entire stack around languages and tools made for programmers.  Let designers design and programmers program!

## Naming

The framework is called Seven5 because the originator lives in Paris, France. All the postal codes for Paris, proper, begin with 75.  Besides, names don't matter that much.

The use of the strange pronunciation of guise is because it sounds cooler. Plus, the originator lives very close to the residence (compound?) of the  [House de Guise](http://en.wikipedia.org/wiki/House_of_Guise) which is  pronounced in this way.   The English word that is spelled the same way comes from the rumor that a dis_guise_ was used by the Duc an attempt to mask his involvement in the attempted assassination of  [Gaspard de Coligny](http://en.wikipedia.org/wiki/Gaspard_de_Coligny) that lead directly to the  [St. Bartholomew's Day Massacre](http://en.wikipedia.org/wiki/St._Bartholomew's_Day_massacre). Web frameworks  may educate in many ways.

## Why Mongrel2

### Production Reasons

* [Mongrel2](http://www.mongrel2.org) is derived from mongrel.  Both of them have extremely well tested and secure http handling code.  Both are known to perform well under high load and to pass [valgrind](http://www.valgrind.org), so they do not leak memory.  It's solid.

* Mongrel2 is friendly for deployment/operations and can be easily configured to work in a cluster.  Mongrel2 can also handle having clusters, not necessarily in the same configuration, that handle the requests for one or more applications deployed on the cluster.  It's scalable.

### Testing Reasons

* Test the "front door" not some other path.  In other words, the best tests  use the code path that is as close---or better yet identical too--the code path  that is used by the end-user.  Unit tests in Seven5 code through the exact same dispatching (sometimes called "routing") as a request in a production deployment, even in a clustered deployment.

* Mongrel is easy to configure and control programmatically.  Seven5  exploits this ability to allow the server to be configured based on its own conventions of how to develop a web application.  During development it  should never be necessary to touch a configuration file.  Seven5 also uses this ability to programmatically start or restart mongrel2 as needed to run the developer's web application.
