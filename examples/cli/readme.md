

example **CLI executable**
============================

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
[crypt-override]:               https://github.com/tfwio/session/blob/7c101cae41533a59124cac9b1664e5deb354b429/crypt.go#L16 "crypt.go OverrideCrypto(…)"
[GORM]:                         https://github.com/jinzhu/gorm
