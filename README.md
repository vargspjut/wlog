# wlog
A simple golang logger with log-level, hooks and structured logging capabilities. Wlog also supports split-output to both stdout and optionally a file.

Available log-levels:
- Debug
- Info
- Warning
- Error
- Fatal
  
## Installation
```
go get github.com/vargspjut/wlog
```

## Usage
```golang
package main

import "github.com/vargspjut/wlog"

func main() {

  // Set log level to Debug (most verbose). Default is Info
  wlog.SetLogLevel(wlog.Dbg)

  // Log some messages with different log-levels
  wlog.Debug("This is a debug message")
  wlog.Debugf("This is a %s", "formatted debug message")
  wlog.Info("This is an info message")
  wlog.Warning("This is a warning message")
  wlog.Error("This is an error message")
  wlog.Fatal("This is a fatal message that will call os.exit")  
  wlog.Fatalf("This is a fatal message %s", "that won't be printed")
}
```

The output would look similar to this
```
2019-03-10 20:57:04:280279 DBG This is a debug message
2019-03-10 20:57:04:280394 DBG This is a formatted debug message
2019-03-10 20:57:04:280402 NFO This is an info message
2019-03-10 20:57:04:280408 WRN This is a warning message
2019-03-10 20:57:04:280414 ERR This is an error message
2019-03-10 20:57:04:280421 FTL This is a fatal message that will call os.exit
```

### Logging hooks
A logging hook is a function callback that can be used to perform common tasks when a logging event is triggered. You may install any number of hooks per logging level.

```golang
// Code left out for brevity

// Install a hook to catch all Info messages
wlog.InstallHook(wlog.Nfo, func(timestamp time.Time, l wlog.LogLevel, msg string){
  // Perform some action here for all Info log events
})

// Install a hook to catch all Error messages
wlog.InstallHook(wlog.Err, func(timestamp time.Time, l wlog.LogLevel, msg string){
  // Perform some action here for all Error log events
})

wlog.Info("This is an Info log entry. Hook installed")
wlog.Error("This is an Error log entry. Hook installed")
wlog.Warning("This is a Warning log entry. No hooks installed")
```

### Structured logs
*wlog* provides support for structured logs allowing additional fields to be added to a mutable `Logger` instance. This includes the default logger and any new loggers created with wlog.New(...). You can also create a scoped logger based on any Logger instance that will inherit the fields from the parent Logger. See the code snippet below:

```golang
package main

import "github.com/vargspjut/wlog"

func main() {

    // The default logger
    wlog.SetLogLevel(wlog.Nfo)
    wlog.SetFormatter(wlog.JSONFormatter{})
    wlog.SetFields(wlog.Fields{"userId": "dd18f2b6-35df-11ea-bb24-c0b88337ca26"})

    // Create a scoped logger based on the default logger
    scopedLogger1 := wlog.WithScope(wlog.Fields{"field1": "field1_value"})

    scopedLogger1.Info("This is an log entry")
    scopedLogger1.Info("This is another log entry")

    // Create a new scoped logger based on a another scoped logger
    scopedLogger2 := scopedLogger1.WithScope(wlog.Fields{"field2": "field2_value"})

    scopedLogger2.Info("This scoped logger is inherited from scopedLogger1")

    // Create a new mutable logger
    logger := wlog.New(nil, wlog.Nfo, true)
    logger.SetFormatter(wlog.JSONFormatter{})
    logger.SetFields(wlog.Fields{"tenantId": "aa18f2b6-35df-11ea-bb24-c0b88337ca26"})

    logger.Info("This is a new logger with no relation to above loggers")

    // Create a new scoped logger based on logger
    scopedLogger3 := logger.WithScope(wlog.Fields{"field3": "field3_value"})

    scopedLogger3.Info("This scoped logger is inherited from logger")
}
```
Here, *wlog* is set the log level `INFO` and after that we invoke the method `SetFields` passing `wlog.Fields` which is an alias type to `map[string]interface{}`. `SetFields` will attach `Fields` to the `Logger` default scope. Next, the `WithScope` method is called receiving `Fields` as arguments, this will return a new scope with `Fields` attached to it, this new scope will contain its own list of `Fields`, plus 
the fields previously added to the default scope. The rest of the example demonstrates variants on this behaviour. Child scopes use the `JSONFormatter` by default.
 
Running this code you should a similar output to this:

```json
{"field1":"field1_value","level":"Info","msg":"This is a log entry","timestamp":"2020-01-23 09:57:54:157141","userId":"dd18f2b6-35df-11ea-bb24-c0b88337ca26"}
{"field1":"field1_value","level":"Info","msg":"This is another log entry","timestamp":"2020-01-23 09:57:54:157273","userId":"dd18f2b6-35df-11ea-bb24-c0b88337ca26"}
``` 
Note that we call the method `Info` two times with different messages, however, the field `userId` added to the default logger sticks with the scoped logger and will be part of any log entry written by that logger or any of its descendants. This behaviour was inspired by the great library [logrus](https://github.com/sirupsen/logrus)

The `JsonFormatter` has support for compact property names, this can be achieved by setting its
property `Compact` to `true`, like so:

```golang
wlog.SetLogLevel(wlog.Nfo)
wlog.SetFormatter(wlog.JSONFormatter{Compact: true})
wlog.SetGlobalFields(wlog.Fields{"userId": "dd18f2b6-35df-11ea-bb24-c0b88337ca26"})

wlog.Info("This is a log entry")
```
Output:
```json
{"@l":"Info","@m":"This is a log entry","@t":"2020-02-05 12:19:30:163927","userId":"dd18f2b6-35df-11ea-bb24-c0b88337ca26"}
```
Note that, the default fields `level`, `timestamp` and `message` are renamed to its first letter prefixed
with a `@` symbol. The `Compact` property in the `JsonFormatter` is optional and it is set to `false`
by default.

### Field mapping
When using JSON compact format it is possible to customize the name of fields using the `FieldMapping`, eg:

```go 
wlog.SetFormatter(wlog.JSONFormatter{Compact: true, FieldMapping: FieldMapping{"username", "un"})
scopedLogger = wlog.WithFields(wlog.Fields{"username": "test"})
scopedLogger.Info("This is a log entry")
``` 
Output:
```json
{"@l":"Info","@m":"This is a log entry","@t":"2020-02-05 12:19:30:163927","un":"test"}
```
In the example above if the option `Compact` was set to `false`, then the field would be named `username`.

Note: When creating `FieldMapping`, the name of field can't be prefixed with the symbol `@`, since it is reserved
for default fields like `@t` (timestamp), `@l` (level) and `@m` (message). 

To add new mappings after the logger initialization, use the `SetFieldMapping` function, eg:

```go
wlog.SetFormatter(wlog.JSONFormatter{Compact: true, FieldMapping: FieldMapping{"username", "un"})
scopedLogger = wlog.WithFields(wlog.Fields{"username": "test", "firstname": "John", "lastname": "Smith"})

wlog.SetFieldMapping(wlog.FieldMapping{"firstname": "fn", "lastname": "ln"})

scopedLogger.Info("This is a log entry")
```

Output:
```json
{"@l":"Info","@m":"This is a log entry","@t":"2020-02-05 12:19:30:163927","un":"test", "fn": "John", "ln":  "Smith"}
``` 

Fields that are not mapped will be shown in a non-compact manner.

## Test
```
go test
```
