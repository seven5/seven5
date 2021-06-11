Every method, variable, class, interface, type and so on in Seven5 is designed such 
that code that is _intended_ to be the public API starts with an uppercase letter.  E.g.  `foo()` is
a function that is not intended to called by user level code whereas `Foo()` is part of the public
API.

At the top level, you can import the faux namespaces like `S5Render` from `render`, `S5Log`
from `log` and so forth.  (It is a custom to use `S5Err` for the error namespace,
not `S5Error`.) Each of these only exports the public API so you are 
always "in bounds" if you limit your use to these.  These expose names like `S5Render.Rect` for
the class `Rect`; if you prefer shorter notation you can alias these names
with `import Rect = S5Render.Rect` after importing the whole `S5Render` package
or you can just import `Rect` directly.

Examples:
```typescript
import * as S5Render from '../render';
import Rect = S5Render.Rect;
import {Rect} from '../render';
```

Seven5 does not prevent you from using `foo()`. There are no private methods, classes,
interfaces, or variables.  Having private entities would prevent you 
from doing something clever--particularly something clever that we have never 
thought of before.  However, the warning above still stands, that `foo()` is an 
internal function and it might change in the future--and likely without warning.

The only possible exception to this rule is a case where some function, method, etc
is marked `private` because it is _already known_ that this is going to change
soon.  In other words, it's preventing from using an API that will be out of date
shortly.  At the time of writing, there are none of these in the code.

If you use an IDE that has autocompletion, you may see properties like `_foo` and
`_bar` on some classes.  These properties are the implementation of the methods
`Foo()`/`SetFoo()` and `Bar()/SetBar()`.  Unless you are doing something really
exotic, it's best to use the upper case accessors, like `mybaz.Foo` or `mybaz.Bar`.

-------------------------------------------------------------------

When doing imports within any portion of the toolkit (like "render" or "evalvite")
we strive to only import _other_ entire portions but import specific types
from the same portion.  These typically differ in the path because on variety
uses '..' (other portion) and the other uses '.'.

Example, `interactor/defaultInteractor`:
```typescript
import * as S5Render from '../render';  //different part of toolkit
import interactor from './interactor'   //same part of toolkit
```
