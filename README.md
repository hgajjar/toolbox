# Toolbox
This repository contains reimplementation of some Spryker commands using Golang.

## Available commands
### Queue worker
This command works in same way as Spryker's `console queue:worker` command. However, there are some differences:
 * It uses AMQP protocol to connect to RabbitMq for checking the pending queues, which is very reliable and much faster.
 * By default it keeps running until all queues are empty. But it also has support for daemon-mode, which keeps it running in background and picking up any new messages in queues.
 * It has an inbuilt yml configuration that contains list of queues that it uses to "watch". Check toolbox.yml for reference (it can also be overwritten at runtime using -c flag).
```
go run main.go queue:worker -h
Usage:
  toolbox queue:worker [flags]

Flags:
  -d, --daemon-mode   Keep queue workers running in daemon mode.
  -h, --help          help for queue:worker

Global Flags:
  -c, --config string   config file (default is toolbox.yml in current dir)
  -v, --verbose         verbose output
```

### Sync data
This command works in same wasy as Spryker's `console sync:data` command. Some differences are:
 * Again it uses AMQP protocol for RabbitMq connection, so publishing messages is very fast. Also, it has a mechanism to "wait" if RabbiMq blocks the producer channel temporarily (it does this when you push a lot of messages and it needs time to persist them on disk to save memory).
 * It has an inbuilt yml configuration that contains list of "resources" that it uses to sync data. Check toolbox.yml for reference (it can also be overwritten at runtime using -c flag).
 * It provides a flag `-q` to also run the queue worker in background while doing the sync. So you can do complete synchronization using one command.
```
go run main.go sync:data -h
Usage:
  toolbox sync:data [flags]

Flags:
  -h, --help               help for sync:data
  -i, --ids string         Defines ids for entities which should be exported, if there is more than one, use comma to separate them.
                           If not, full export will be executed.
  -r, --resource string    Defines which resource(s) should be exported, if there is more than one, use comma to separate them.
                           If not, full export will be executed.
                           	
  -q, --run-queue-worker   Run queue workers in the background.

Global Flags:
  -c, --config string   config file (default is toolbox.yml in current dir)
  -v, --verbose         verbose output
```
