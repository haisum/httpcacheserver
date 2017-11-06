This is a very simple http cache server for files.

You can run it to speed up frequently accessed files on your local network. I personally use it for caching a maven repository. It has also been tested as yum repository cache server.

Execution is simple, you can put parameters in config.yaml or in environment variables with HCS_ as prefix followed by parameters in config.yaml.

Usage Example
--------------

If you wanted to cache http://mirror.centos.org/centos/ and put suffix as /proxy, you can access http://mirror.centos.org/centos/timestamp.txt on http://localhost:8081/proxy/timestamp.txt. Future requests will be served from cache unless file on remote host is modified.

All cache is saved in directory defined in DATA_DIR parameter. Server doesn't automatically expire any cache, so if you want to delete cache on some criteria like files which are N days old, you can safely write a bash/batch script and put it as cronjob to delete files in cache directory.

Currently it relies on Last-Modified header, and won't work if this header isn't provided.

Building
---------------

`go build -o httpcacheserver main.go`

If you don't have go installed, you can download pre-compiled binaries from releases tab on Github.com

Executing
-------------

####Linux

nohup ./httpcacheserver &

####Windows

httpcacheserver.exe

