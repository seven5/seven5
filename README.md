# Seven5

## An opinionated web framework written in go.

Design
------------

This is the seven5 layer. Seven5 provides a "rough passthrough" for implementations that need access to the raw mongrel2 layer.  These two passthroughs are the interfaces `Httpified` and `Jsonified`.  These interfaces allow you to process messages that are very close to the raw
mongrel2 protocol--just a few things were already parsed out rather than expecting every
developer to do that.  The `Httpified` and `Jsonified` interfaces must be implemented by the developer. See `echo_rawhttp.go` and `chat_jsonservice.go` for examples.

Introductions, installation notes, and much pontification can be found at the [Seven5 Site](http://seven5.github.com/seven5/).
