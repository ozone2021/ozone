# Commands

## run

```bash
ozone run [-d -c <context>] <...runnables>
```

-d = detached headless mode, you must set a context if using this mode.

### Non headless mode

The run command in non-headless mode interacts with the daemon, which stores the cache hashes.

It is an interactive bubbletea(interactive CLI framework) app. The runResult is displayed and updated in realtime as the 
build is running.

### Headless mode

No local caching. This is meant for use with CI/CD.

# TODO example image

The controller ozone-lib/run/runapp_controller/controller.go that looks after the run_app wraps
the bubbletea app.

It also contains a logRegistrationServer. The logRegistrationServer is a grpc server that listens for new log_apps connecting
via unix pipe that is created in /tmp/ozone/<hash_of_working_dir>/socks/log-registration.sock

After connecting, the runResult is updated in the new log_app in realtime.


## logs

The log app is run using this command:

```bash
ozone logs
```

The main build app must be running in interactive mode (non-headless), and not be closed.

If the build app is closed, you can leave the log app running and it will reconnect when the build app is restarted.

#### Why does this exist?

When we build docker images in parallel, we don't want the output all spewing out to the same terminal windows, as this 
is hard to read.

The log_app acts as a multiplexer for the logs, and allows you to read each build log separately, as the logs are all
sent to log files in /tmp/ozone/<hash_of_working_dir>/<run_id>/logs/

It basically runs an equivalent to the linux tail command on whichever log file is selected.

