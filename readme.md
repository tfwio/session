status: in development  
*valid sessoins have been broken in the last few commits*


A session package intended to provide some foundation which may be cusomized and
implemented, perhaps useful in [github.com/gin-gonic/gin] middleware aside any other heuristic it can be wired to.

This package provides a secure logon session by utilizing a sqlite3 database via [GORM],
so easily conforms to other data-systems.

users: id name salt hash

sessions: id userid sessid created expires cli-key

login logout stat register *!unregister*

Filter URI in middleware

**LIMITATIONS**

Each of these limitations is a blessing in disguise 😁

- freshly brewed.
- [*todo/feature*] One session on one client (browser session / IP) is allowed per User once initial session is created.  (can easily be modified)  
  Will likely fix this soon.


**EXAMPLES**


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
[service-config]:               https://github.com/tfwio/session/blob/2c2adb376cde8c0b31d269f8686df51d0f64eb62/examples/srv/main.go#L14
[crypt-override]:               https://github.com/tfwio/session/blob/7c101cae41533a59124cac9b1664e5deb354b429/crypt.go#L16 "crypt.go OverrideCrypto(…)"
[ClientIP]:                     https://github.com/gin-gonic/gin/blob/f98b339b773105aad77f321d0baaa30475bf875d/context.go#L690
[GORM]:                         https://github.com/jinzhu/gorm
[github.com/gin-gonic/gin]:     https://github.com/gin-gonic/gin
[GORM]:                         https://github.com/jinzhu/gorm
[unsafe-handlers]:              https://github.com/tfwio/session/blob/053b1d9438caa8bac618b7c6a42f9756a518ab82/examples/srv/conf.go#L71
