# go-safeclient - SafeClient CLI and Go Library

**This project is alpha quality and the APIs may change without notice. Use with caution!**

This project, go-safeclient, provide both a CLI and a Go library for using the
[SAFE Launcher API](https://maidsafe.readme.io/docs/introduction).

**WARNING: Many of the CLI command names and functions may not work as one might expect a POSIX system to work. For
example, in SAFE moving a directory to a directory moves it UNDER that directory, it does not just rename it. Similarly,
you can't move a file and change the name at the same time since the destination of the move MUST be the directory it
will be moved under.**

## Command Line Interface

The CLI provides the ability to use SAFE from the command line. Releases for some operating systems may be found in the
[releases](https://github.com/cretz/go-safeclient/releases) section. Otherwise, follow instructions for building below.

### Building

To build the project, make sure your [GOPATH](https://github.com/golang/go/wiki/GOPATH) environment variable is set to
an appropriate directory. Then:

    go get -u github.com/cretz/go-safeclient

Once it has been fetched, navigate to the directory it is in (i.e. `$GOPATH/src/github.com/cretz/go-safeclient`) and
run:

    go build

This will create an executable in the local directory called `go-safeclient` (`go-safeclient.exe` on Windows).

### Usage

Below is the dump from `go-safeclient help`:

    Usage:
      go-safeclient [command]
    
    Available Commands:
      cp               Copy file
      cpdir            Copy directory
      dnsaddservice    Add DNS Service
      dnscreatename    Create DNS Name
      dnsdeletename    Delete DNS Name
      dnsdeleteservice Delete DNS service
      dnsfile          Get file at specific path for DNS name and service name
      dnsnames         Get all DNS names
      dnsregister      Register DNS
      dnsservicedir    Get DNS service dir for the given name and service
      dnsservices      Get all DNS services for the given name
      fetch            Fetch file contents
      ls               Fetch directory information
      mkdir            Create directory
      mod              Change file name and/or metadata
      moddir           Change directory name and/or metadata
      mv               Move file
      mvdir            Move directory
      ping             Do simple ping to make sure app is registered
      put              Put file contents
      rm               Delete file
      rmdir            Delete directory
      touch            Create empty file
    
    Flags:
      -c, --config string   config file (default "conf.json")
      -h, --help            help for go-safeclient
      -v, --verbose         show debug output

For information about an individual command, run `go-safeclient help [command]`. The easiest way to get started is just
to run:

    go-safeclient ping

The `-c` option specifies the file to use for configuration. By default it is `conf.json` in the local directory. If
this file does not exist (i.e. upon initial execution), the program will attempt to authenticate with a running
[SAFE Launcher](https://maidsafe.readme.io/docs/getting-started) and write the configuration to the specified file.

Here are commands to create new safenet site assuming "mysite" isn't already registered (tested on Windows):

    go-safeclient mkdir /mysitedir
    go-safeclient touch /mysitedir/index.html
    echo "Hello World!" | go-safeclient put /mysitedir/index.html
    go-safeclient dnsregister mysite www /mysitedir

The site can now be reached at http://www.mysite.safenet (or safe://www.mysite if you are using the `safe://` protocol).

## Library

Documentation for the client library can be found [here](https://godoc.org/github.com/cretz/go-safeclient/client). The
documentation is currently sparse but will be filled in as the project matures. Here is a simple example of writing
"Hello, World!" to a file and reading it back (which can also be done by running
`go run _examples/simple_read_write.go`):

```go
package main

import (
    "fmt"
    "github.com/cretz/go-safeclient/client"
    "io/ioutil"
    "strings"
)

func main() {
    // Create the client
    myclient := client.NewClient(client.Conf{})
    err := myclient.EnsureAuthed(client.AuthInfo{
        App: client.AuthAppInfo{
            Name:    "My Application",
            ID:      "my.application.id.mysite.safenet",
            Version: "0.1.0",
            Vendor:  "myname",
        },
    })
    if err != nil {
        panic(err)
    }
    // Note, the myclient.Conf structure should be saved here to reuse authentication in the future

    // Create a file
    err = myclient.CreateFile(client.CreateFileInfo{FilePath: "/myfile.txt"})
    if err != nil {
        panic(err)
    }

    // Write to it
    err = myclient.WriteFile(client.WriteFileInfo{
        FilePath: "/myfile.txt",
        Contents: ioutil.NopCloser(strings.NewReader("Hello, World!")),
    })
    if err != nil {
        panic(err)
    }

    // Read it back out and print to stdout
    rc, err := myclient.GetFile(client.GetFileInfo{FilePath: "/myfile.txt"})
    if err != nil {
        panic(err)
    }
    defer rc.Close()
    contents, err := ioutil.ReadAll(rc)
    if err != nil {
        panic(err)
    }
    fmt.Printf("Contents of /myfile.txt: %v\n", string(contents))
}
```

Many useful client functions accept a `XXXInfo` struct. This is to help with backwards compatibility as new features are
added to different calls.

### Integration Tests

Included in this project is a suite of integration tests. Currently it is small but it will grow over time to cover the
entire spectrum of API calls. I can be executed by running the following in the base folder of the project:

    go test ./integration -tags integration

Add `-v` to get verbose output to watch it as it runs with debug HTTP information. Note, this makes many files and
directories which will cost safecoin if executed in a production environment.

### Contributing

Before submitting a pull request, please make sure that `go fmt`, `go vet`, and `golint` pass along with the entire
integration suite.

## License
   
The MIT License (MIT)

Copyright (c) 2016 Chad Retz

Permission is hereby granted, free of charge, to any person obtaining a copy
of this software and associated documentation files (the "Software"), to deal
in the Software without restriction, including without limitation the rights
to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
copies of the Software, and to permit persons to whom the Software is
furnished to do so, subject to the following conditions:

The above copyright notice and this permission notice shall be included in all
copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
SOFTWARE.