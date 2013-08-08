---
layout: book
chapter: Storing articles in the database with Hood
---

### Goal: Understand how to store and retrieve values to and from a database
After you've read this chapter, you should be able to use [Hood](https://github.com/eaigner/hood) to model your server-side resources and store and retrieve them from the database.  You can use the code you checked out previously from the branch "code-book-1" for this chapter.

### Theory: See-kret! This isn't really part of _Seven5_
_Seven5_ is a web development tool, not a database tool. If you want to skip this chapter entirely because you prefer alternative database (or non-relational) storage strategies for the client side, the author will not be bothered.  For server-side go development, there are tools like [Jet](https://github.com/eaigner/jet), [modsql](https://github.com/kless/modsql?source=cc) and the go library's [built-in support for SQL](http://golang.org/pkg/database/sql/) that may interest you.

This chapter has been included because of the strange fixation many developers have about storage, particularly for doing web development. "What? No database mapping?" is a sad but common refrain.  Other web toolkits' tight integration with databases and more specifically [ORMs](http://www.codinghorror.com/blog/2006/06/object-relational-mapping-is-the-vietnam-of-computer-science.html), seems to have caused many people to think that this is integral to the web toolkit.  The author does not agree.  See [Active Record in Rails](http://guides.rubyonrails.org/active_record_basics.html) and [Django's Models](https://docs.djangoproject.com/en/dev/topics/db/models/) for examples of toolkits that have strongly connected the storage models to the toolkit.  However because the desire to have _some_ storage story is so strong among so many, this chapter has been created and the "typical" development model altered to include it.


### Practice: Installing the hood tool
Assuming you have set your `GOPATH` as was discussed previously, you should be able to install the tool portion of hood like this:

```
$ go get github.com/eaigner/hood
$ cd /tmp/book/go/src/github.com/eaigner/hood
$ go run cmd/gen/templates.go
$ go build -o /tmp/book/go/bin/hood github.com/eaigner/hood/cmd
```

You can test that you have installed hood's tool `hood` correctly with `which hood`:

```
$ which hood
/tmp/book/go/bin/hood
```

### Practice: Setting up your first migration
Hood expects you to have a file in `/tmp/book/db/config.json` that specifies the properties to connect to your database.   The default configuration of _Seven5_ as you checked it out from the git repository *does* put this `config.json` into the source code repository.  This is by design as _Seven5_ keeps things that are *not* meant to be in source code control in environment variables.  You should not put any passwords or usernames 

If you want to create a "blank" configuration file, you can create the file `config.json` by running the following command in the root directory of the project.  _This will destroy (overwrite) any configuration currently in the file!_  We have included the `cd` command a reminder that this `hood` commands should be run at the top of the project :

```
$ cd /tmp/book
$ echo "danger: overwrites config file"
danger: overwrites config file
$ hood create:config
2013/08/04 12:11:54 created db configuration 'db/config.json'
```

If you are using the default configuration provided from the git repository, this is the contents of the configuration file:

```
{
  "development": {
    "driver": "postgres",
    "source": "dbname=${NULLBLOG_DBNAME} user=${NULLBLOG_USER} sslmode=disable"
  },
  "production": {
    "driver": "postgres",
    "source": "dbname=${NULLBLOG_DBNAME} user=${NULLBLOG_USER} password=${NULLBLOG_PASS} sslmode=verify-full"
  },
  "test": {
    "driver": "postgres",
    "source": "dbname=${NULLBLOG_DBNAME} user=${NULLBLOG_USER} dbname=nullblog sslmode=disable"
  }
}
```

The default environment is "development" if you don't specify.

### Practice: Environment variables for database configuration
The _Seven5_ practice is store the database, username, and password for accessing the relational store as environment variables.  The following are typical environment variable settings for the `nullblog` application when doing development:

```
$ export NULLBLOG_USER=postgres
$ export NULLBLOG_PASS=seekret
$ export NULLBLOG_DBNAME=nullblog
```

You may need to adjust these depending on how you have set up your local database installation.  You will need to have these set to run the commands below.

### Theory: Environment variable names
_Seven5_ expects that the application name prefixes all environment variable names.  This is to allow easy combination and co-development of different projects from the same shell.  This expectation makes environment variable names a bit longer, but it is also expected that developers will keep all their environment variable settings together in simple shell scripts.  Thus, this is less of an issue than it would be were these typed on the command line. The author recommends the use of scripts like `enable-nullblog` that can be used with `source` in a shell to enable a particular project.

### Practice: Database creation
Most relational database systems, including Postgres and MySQL, expect database creation to be rare enough that it can be done "out of band".  From the command line, you'll need to create a database on your local machine for development on the nullblog project.  With Postgres, the command looks like this:

```
$ createdb -U postgres -O postgres nullblog
```

The command above sets the user and owner of the database to be one named "postgres", which is the default for a simple postgres installation.  You may need to modify this for your local database configuration.

### Theory: The type name "Article"
In [chapter 1](http://localhost:4000/seven5/nullblog.html), we introduced the wire type `ArticleWire` and the resource type `ArticleResource`.  Neither of these were named "Article" because it was reserved to be used for database storage.  For the server side, the "base" name of the noun is expected to be something that we stored in a relational database.  For this section of the book, the definition is `Article` in `/tmp/book/go/src/nullblog/nullblog.go`

### Practice: Creating the first migration
A database `migration` is a change to the schema (or other part of the data model) of a database that is both programmatically defined and reversible.  Typically, the operation "migrate" is used to logically forward to newer versions of the data model and "rollback" is used for the reverse.  The nomenclature used by hood is very similar to that used by [Active Record Migrations](http://guides.rubyonrails.org/migrations.html) from Rails.

We can create a new, empty migration by running this command in the project root directory:

```

$ hood create:migration CreateArticleTable
2013/08/04 12:55:59 created migration '/Users/iansmith/codebook/db/migrations/init.go'
```

The result can be see in `/tmp/book/db/migrations`.  For this chapter, you probably don't want to have two "CreateArticleTable" migrations.  We will assume that you are using the one checked out from the repository, `/tmp/book/db/migrations/1375646159_CreateArticleTable.go`.  The "up" and "down" methods migrate or rollback a particular operation.  In this case this is the creation of the `Article` table.

The code for the forward and backward migrations are simple:

```
func (m *M) CreateArticleTable_1375646159_Up(hd *hood.Hood) {
	hd.CreateTable(&nullblog.Article{})
}

func (m *M) CreateArticleTable_1375646159_Down(hd *hood.Hood) {
    hd.DropTable(&nullblog.Article{})
}
```

So, we are effectively setting the forward migration to create a table defined by the fields of the `Article` type and the backward migration to delete that same table.

### Practice: Running the first migration

You can run the migration to create the necessary database table(s) like this:

```
$ hood db:migrate
2013/08/04 15:37:38 applying migrations...
2013/08/04 15:37:38 applying CreateArticleTable_1375646159_Up...
2013/08/04 15:37:38 applied 1 migrations
2013/08/04 15:37:38 generating new schema... /tmp/book/db/schema.go
2013/08/04 15:37:38 wrote schema /tmp/book/db/schema.go
2013/08/04 15:37:38 done.
```

The are a number of `hood` commands to manipulate the database, moving it forward or backward by some number of migrations. These are detailed [elsewhere](https://github.com/eaigner/hood#migrations) but the two most important ones are the one shown above, `hood db:migrate` and the reverse `hood db:rollback`.

### Practice: Using the database

With all the preliminaries completed, we can now update our implementation of `ArticleResource` to use `Article` and the now defined database plus table(s). The code for the new implementations of `Index` and `Find` are replacing our prior, hard coded implementations in `/tmp/book/go/src/nullblog/nullblog.go`:





