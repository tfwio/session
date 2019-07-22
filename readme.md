[working-example]: https://github.com/tfwio/sekhem/blob/cd0c5c5021683d424ff9b351b7a6258f7f2e5bde/fsindex/config/serve.logon.go
[crypt.cli/sess.go]: crypt.cli/sess.go

A session package intended to provide a skeleton which may be cusomized and
implemented, perhaps useful in github.com/gogonic/gin middleware among other things.

In its current form, this package provides a secure logon session by utilizing a
sqlite3 database via GORM, which makes it simple to migrate to other database systems.

**example: gin-gonic middleware**

Though embedded in an existing server, the following example allows
for the following file to be compiled in using a Golang build tag `session`
when the executable is compiled.  (I.E. `GO111MODULES=on go build <...> -tag 'session <…>' -o srv.exe`)

A working middleware can be found here:  
[github.com/tfwio/sekhem/fsindex/config/serve.logon.go][working-example]


**example CLI executable**

Another example implementation can be found in the [crypt.cli/sess.go] CLI app.

This example can be compiled witout using GO111MODULES.  
To compile the CLI app, run the following from the root directory.

```bash
go build crypt.cli/*
```


