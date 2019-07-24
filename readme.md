

A session package intended to provide some foundation which may be cusomized and
implemented, perhaps useful in [github.com/gin-gonic/gin] middleware aside any other heuristic it can be wired to.

This package provides a secure logon session by utilizing a sqlite3 database via [GORM],
so easily conforms to other data-systems.

**FEATURES**

- User salt and hash table generation
- Validate User password
- Client (browser) session create, destroy and validate
- Easily customize data source, system and salt-size in
  [configs.go::SetDefaults(…)][setdefaults] or just
  [crypt.go::Override(…)][crypt-override] hash memory, time and key-length.
- Client (browser) sessions have two expiration options: when browser session ends (browser is closed) and "keep alive" (as shown in the example gin web-app).
- The crypto incorporated is powerful and presently (201907) concurrent
  with todays standards.
- As provided in the gin web-app example, `Session` "KeepAlive" setting is automatially
  maintained using the following heuristic:  
  Use xhr/json to check the client session status (`/stat/`) on app-launch.  
  *...A client web-app will check/call this on each page-load in order to
  know which to render: login, logout or register menu-option(s) or form(s).*
- More to come…

**LIMITATIONS**

Each of these limitations is a blessing in disguise 😁

- freshly brewed.
- [*todo/feature*] One session on one client (browser session / IP) is allowed per User once initial session is created.  (can easily be modified)  
  Will likely fix this soon.


**EXAMPLES**


example: **gin-gonic middleware**

If implemented as middleware, session functionality can easily snap-in to any
[github.com/gin-gonic/gin] server in as little as one source file…

compile with the bash helper

```bash
./do gin
```

or just bash

```bash
go build ./examples/srv
```

The example uses `127.0.0.1:5500` as the default host/port and serves the following URIs
using `context.Any("/login/", handlerFunc)` (heuristic) so you can use either GET or PUT
to test the example, for example:

- http://localhost/register/?user=admin&pass=password  
  create user and a new session for the user, or fails if we allready have the user record.
- http://localhost/login/?user=admin&pass=password  
  loggs client in, creates session or fails if the password isn't valid.
- http://localhost/logout/  
  loggs client out, destroys session or fails if client is not logged in.
- http://localhost/stat/  
  Shows login status for when logged in or not. *Note that when you've logged in
  with `…&keep=true`, you will notice the created/expires date updated.*
- http://localhost/index/  
  This is protected, requiring client to be logged in.

Each of the above is intended for XHR/JSON response and will have a JSON response.

In the example, stat is a special case that we're protecting or requiring a login
session in order to allow content to be served, however in practice I wouldn't block
this particular URI so I can get a valid response using XHR/JSON.

example: **CLI executable**

Another example implementation can be found in the [crypt.cli/sess.go] CLI app.

The CLI app is generally useful for looking at and configuring or testing User
salt/hash settings, verifying a generated user-password, etc;
not intended for use with or as a companion to the web-app.

To compile:
```bash
go build ./examples/cli
```

Or use the build helper (bash) script `./do cli`

```bash
./do cli
```

The tool allows you to (1) `create` a user/password which includes generation of
a session for the user, (2) `validate` the user/password and (3) `list` all
sessions including the user owning the session.

Apparently if you attempt to create the same user twice, it will fail to generate the user however creates a new session.  
The cli tool was written just to test creation and validation of user passwords as well as [GORM].

For example, after you compile the CLI app...


**generate a user**

```bash
./cli create -u admin -p password
```
generates the following output
```text
{1 admin dW6tmIxySUoV9SJr5OJ9aYH1b35QzuqpSoxo1KmHVIw33FpdM6asZ+Q0uYEFZ2fb K0Bt+bF4qiTNYMkrJ3tF7RdxGBjuV0zsDpV4htJ2B+U=}
success: false; session={1 1 cli-example-app 2019-07-24 08:35:09.4055506 -0500 CDT m=+2.414109701 2020-01-24 08:35:09.4055506 -0600 CST RFNaT0JhQTNpM2xsRno0NmFkeURqdDJKeHpqRGxZQU8veUpjWnJocXkzSjYrR1BVUG8wejQ2QklJbzZ0NThSRA== dW5rbm93bi1jbGllbnQ= false}
```
**verify user password**
```bash
./cli validate -u admin -p doomedtofail
```
outputs
```text
Result: false
```
Lets try the working "password"
```bash
./cli validate -u admin -p password
```
outputs the following success
```text
Result: true
```
Finally, we can call the following to show each user-session stored in the sessions table including the user name.
```bash
./cli list
```

[crypt.cli/sess.go]:            crypt.cli/sess.go
[setdefaults]:                  https://github.com/tfwio/session/blob/bb6bd69e91f3ca4ef880e5f216fe25c3febd5912/configs.go#L27
[crypt-override]:               https://github.com/tfwio/session/blob/16e442ee2d7bb51873e2741dd5aa98f0751abbe4/crypt.go#L20
[ClientIP]:                     https://github.com/gin-gonic/gin/blob/f98b339b773105aad77f321d0baaa30475bf875d/context.go#L690
[GORM]:                         https://github.com/jinzhu/gorm
[github.com/gin-gonic/gin]:     https://github.com/gin-gonic/gin
[working-example]:              https://github.com/tfwio/sekhem/blob/cd0c5c5021683d424ff9b351b7a6258f7f2e5bde/fsindex/config/serve.logon.go
[GORM]:                         https://github.com/jinzhu/gorm