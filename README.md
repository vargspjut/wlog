# wlog
A simple golang logger with log-level and structured logging capabilities. Wlog also supports split-output to both stdout and optionally a file.

Version 1.0.2 introduces logging hooks. See below for an example on how to install and use hooks.

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
*wlog* provides support for structured logs allowing the addition of fields that will "stick" with the log. When using the`wlog.WithFields` method the log entries will automatically be serialized to JSON and the string passed to logging methods like `wlog.info` will be included in the JSON object as well. See the code snippet below:

```golang
package main

import "github.com/vargspjut/wlog"

func main() {

    wlog.SetLogLevel(wlog.Nfo)
    wlog.WithFields(wlog.Fields{"userId": "dd18f2b6-35df-11ea-bb24-c0b88337ca26"})

    logger.Info("This is a log entry")
    logger.Info("This is another log entry")

}
```
Here *wlog* is set the log level to `INFO` and after that we invoke the method `WithFields` passing `wlog.Fields` which is a alias type to `map[string]interface{}`. Running this code you should see the output below:

```json
{"level":"Info","msg":"This is a log entry","timestamp":"2020-01-14 10:49:03:627880","userId":"dd18f2b6-35df-11ea-bb24-c0b88337ca26"}
{"level":"Info","msg":"This is another log entry","timestamp":"2020-01-14 10:49:03:627981","userId":"dd18f2b6-35df-11ea-bb24-c0b88337ca26"}
``` 
Note that we call the method `Info` two times with different messages, however, the field `userId` added to the log sticks with the logger and it is part of the log entry at all times. This behaviour was inspired by the great library [logrus](https://github.com/sirupsen/logrus)
## Test
```
go test
```
