# wlog
A simple golang logger with log-level capabilities. Wlog also supports split-output to both stdout and optionally a file. 


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

## Test
```
go test
```
