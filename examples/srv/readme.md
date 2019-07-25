


example: **gin-gonic middleware**

If implemented as middleware, session functionality can easily snap-in to any
[github.com/gin-gonic/gin] server.

compile with the bash helper

```bash
./do srv
# or standard way
go build ./examples/srv

# and run it
./srv
```

The example uses `127.0.0.1:5500` as the default host/port and serves the following URIs
using `context.Any("/login/", handlerFunc)` (heuristic) so you can use either GET or PUT
to test the example, for example:

- http://localhost/register/?user=admin&pass=password  
  http://localhost/register/?user=admin&pass=password&keep=1 or  
  http://localhost/register/?user=admin&pass=password&keep=true
  create user and a new session for the user, or fails if we allready have the user record.
- http://localhost/login/?user=admin&pass=password  
- http://localhost/login/?user=admin&pass=password&keep=1 or  
- http://localhost/login/?user=admin&pass=password&keep=true  
  logs client in, creates session or fails if the password isn't valid.
- http://localhost/logout/  
  logs client out, destroys session or fails if client is not logged in.
- http://localhost/stat/  
  Shows login status for when logged in or not. *Note that when you've logged in
  with `…&keep=true`, you will notice the created/expires date updated.*
- http://localhost/index/  
  This is protected, requiring client to be logged in.

Each of the above is intended for XHR/JSON response and will have a JSON response.

In the example, stat is a special case that we're protecting or requiring a login
session in order to allow content to be served, however in practice I wouldn't block
this particular URI so I can get a valid response using XHR/JSON.

Note that in the example we're using a list of "[Unsafe][unsafe-handlers]" handlers.  
There is a provided mechanism for checking for "unsafe" URI.  The default heuristic
uses a simple regular expression and a list of uri start-paths for example our
callback function checks the incoming URI, and is looking to return true for "unsafe".