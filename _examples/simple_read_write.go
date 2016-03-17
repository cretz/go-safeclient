package main

import (
	"fmt"
	"github.com/cretz/go-safeclient/client"
	"io/ioutil"
	"strings"
)

// This is a simple example that writes "Hello, World!" to a file and reads it back out.
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
