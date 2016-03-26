// +build integration

package integration

import (
	"github.com/cretz/go-safeclient/client"
	"github.com/stretchr/testify/require"
	"io/ioutil"
	"path"
	"strings"
	"testing"
)

func TestSimpleDNS(t *testing.T) {
	// First, create a name
	name := randomName()
	require.NoError(t, safeClient.DNSCreateName(name))
	defer safeClient.DNSDeleteName(name)

	// Make sure name is there
	names, err := safeClient.DNSNames()
	require.NoError(t, err)
	require.Contains(t, names, name)

	// But there are no services
	names, err = safeClient.DNSServices(name)
	require.NoError(t, err)
	require.Empty(t, names)

	// Create a temp dir
	dirName := randomName()
	dirPath := "/" + dirName
	require.NoError(t, safeClient.CreateDir(client.CreateDirInfo{DirPath: dirPath}))
	defer safeClient.DeleteDir(client.DeleteDirInfo{DirPath: dirPath})

	// Create a temp file
	fileName := randomName() + ".js"
	filePath := path.Join(dirPath, fileName)
	require.NoError(t, safeClient.CreateFile(client.CreateFileInfo{FilePath: filePath}))
	defer safeClient.DeleteFile(client.DeleteFileInfo{FilePath: filePath})
	require.NoError(t, safeClient.WriteFile(client.WriteFileInfo{
		FilePath: filePath,
		Contents: ioutil.NopCloser(strings.NewReader("Some Content")),
	}))

	// Add a service to that dir and make sure it's there
	serviceName := randomName()
	require.NoError(t, safeClient.DNSAddService(client.DNSAddServiceInfo{
		Name:        name,
		ServiceName: serviceName,
		HomeDirPath: dirPath,
	}))
	names, err = safeClient.DNSServices(name)
	require.NoError(t, err)
	require.Equal(t, []string{serviceName}, names)

	// And check the dir
	dir, err := safeClient.DNSServiceDir(name, serviceName)
	require.NoError(t, err)
	require.Len(t, dir.Files, 1)
	require.Equal(t, fileName, dir.Files[0].Name)
	require.Empty(t, dir.SubDirs)
	require.Equal(t, dirName, dir.Info.Name)

	// Try to get the whole file
	file, err := safeClient.DNSFile(client.DNSFileInfo{
		Name:     name,
		Service:  serviceName,
		FilePath: "/" + fileName,
	})
	require.NoError(t, err)
	require.Equal(t, dir.Files[0], file.Info)
	require.Equal(t, "application/javascript", file.ContentType)
	requireReadCloserEqualsString(t, "Some Content", file.Body)

	// Now delete the service and make sure it's gone
	require.NoError(t, safeClient.DNSDeleteService(name, serviceName))
	// TODO: Ug, https://maidsafe.atlassian.net/browse/CS-63
	//names, err = safeClient.DNSServices(name)
	//require.NoError(t, err)
	//require.Empty(t, names)
	names, err = safeClient.DNSNames()
	require.NoError(t, err)
	require.NotContains(t, names, name)

	// Do a "register" which should create a new DNS name and service at the same time
	newName := randomName()
	newServiceName := randomName()
	require.NoError(t, safeClient.DNSRegister(client.DNSRegisterInfo{
		Name:        newName,
		ServiceName: newServiceName,
		HomeDirPath: dirPath,
	}))
	defer safeClient.DNSDeleteName(name)

	// Check the name
	names, err = safeClient.DNSNames()
	require.NoError(t, err)
	require.Contains(t, names, newName)

	// Check the dir
	dir, err = safeClient.DNSServiceDir(newName, newServiceName)
	require.NoError(t, err)
	require.Len(t, dir.Files, 1)
	require.Equal(t, fileName, dir.Files[0].Name)
	require.Empty(t, dir.SubDirs)
	require.Equal(t, dirName, dir.Info.Name)

	// Now just get "me Con" out of the file
	file, err = safeClient.DNSFile(client.DNSFileInfo{
		Name:     newName,
		Service:  newServiceName,
		FilePath: "/" + fileName,
		Offset:   2,
		Length:   6,
	})
	require.NoError(t, err)
	requireReadCloserEqualsString(t, "me Con", file.Body)

	// Do an explicit delete (even though we have some deferred) and make sure deleted
	require.NoError(t, safeClient.DNSDeleteName(newName))
	names, err = safeClient.DNSNames()
	require.NoError(t, err)
	require.NotContains(t, names, newName)
}
