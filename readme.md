

A session package intended to provide some foundation which may be cusomized and
implemented, perhaps useful in [github.com/gin-gonic/gin] middleware among other things.

In its current form, this package provides a secure logon session by utilizing a
sqlite3 database via GORM, which makes it simple to migrate to other database systems.


**FEATURES**

- User salt and hash table generation
- User password verification
- Store random generated salt-like string to sessions table and http.Cookie to client (browser)
- Validate User password
- Validate Session using http.Cookie from client-browser
- Easily customize data source, system and salt-size in
  [configs.go::SetDefaults(…)][setdefaults] or just
  [crypt.go::Override(…)][crypt-override] hash memory, time and key-length.
- The crypto incorporated is powerfull and presently (201907) concurrent to todays standards.
- More to come…

**LIMITATIONS**

Each of these limitations is a blessing in disguise 😁

- freshly brewed.
- [*todo/feature*] Only one session on one client (browser session / IP) is allowed.  
  This may actually be good for some situations.  
  Perhaps we'll include optional types of sessions using `go build -tags <…>`.
- [gin-gonic/gin][github.com/gin-gonic/gin] dependency, even for the CLI executable?  
  Again, perhaps we'll use build tags in the future to integrate other HTTP services.
  Gin however provides [ClientIP()][ClientIP] simply and easily which is useful for
  validating our sessions when in web-service.
- Sessions are set by default to last 12h from the time of login.


**EXAMPLES**


example: **gin-gonic middleware**

If implemented as middleware, session functionality can easily snap-in to any
[github.com/gin-gonic/gin] server in as little as one source file…

Though embedded in an existing server, the following example allows for the following
file to be compiled in using a Golang build tag `session` when the executable is compiled.
(I.E. `GO111MODULES=on go build <...> -tag 'session <…>' -o srv.exe`)

A working middleware can be found here:  
[github.com/tfwio/sekhem/fsindex/config/serve.logon.go][working-example]

example: **CLI executable**

Another example implementation can be found in the [crypt.cli/sess.go] CLI app.  
Its generally useless with exception to observing and testing customization to User and
Session heuristics, development or customization.

This example can be compiled witout using GO111MODULES.  
To compile the CLI app, run the following from the root directory.

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



[crypt.cli/sess.go]:            crypt.cli/sess.go
[setdefaults]:                  https://github.com/tfwio/session/blob/bb6bd69e91f3ca4ef880e5f216fe25c3febd5912/configs.go#L27
[crypt-override]:               https://github.com/tfwio/session/blob/16e442ee2d7bb51873e2741dd5aa98f0751abbe4/crypt.go#L20
[ClientIP]:                     https://github.com/gin-gonic/gin/blob/f98b339b773105aad77f321d0baaa30475bf875d/context.go#L690
[GORM]:                         https://github.com/jinzhu/gorm
[github.com/gin-gonic/gin]:     https://github.com/gin-gonic/gin
[working-example]:              https://github.com/tfwio/sekhem/blob/cd0c5c5021683d424ff9b351b7a6258f7f2e5bde/fsindex/config/serve.logon.go
