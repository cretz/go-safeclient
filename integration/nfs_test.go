// +build integration

package integration

import (
	"github.com/cretz/go-safeclient/client"
	"github.com/stretchr/testify/require"
	"io/ioutil"
	"path"
	"strings"
	"testing"
	"time"
)

func TestSimpleNFS(t *testing.T) {
	assertSimpleNFS(t, false, false)
}

func TestSimpleNFSShared(t *testing.T) {
	assertSimpleNFS(t, true, false)
}

func TestSimpleNFSPrivate(t *testing.T) {
	assertSimpleNFS(t, false, true)
}

func TestSimpleNFSSharedPrivate(t *testing.T) {
	assertSimpleNFS(t, true, true)
}

func assertSimpleNFS(t *testing.T, shared bool, private bool) {
	// Create a new directory to work with
	dirInfo := client.CreateDirInfo{
		DirPath: "/" + randomName(),
		Private: private,
		// TODO: https://maidsafe.atlassian.net/browse/CS-57 (we're ignoring all metadata for now)
		Metadata: "",
		Shared:   shared,
	}
	require.NoError(t, safeClient.CreateDir(dirInfo))
	// Go ahead and defer a deletion of it
	defer safeClient.DeleteDir(client.DeleteDirInfo{DirPath: dirInfo.DirPath, Shared: shared})

	checkDirFunctions(t, dirInfo, shared, private)
	checkFileFunctions(t, dirInfo, shared, private)
	// Not even testing moves right now until documented: https://maidsafe.atlassian.net/browse/CS-60
	// checkMoveFunctions(t, dirInfo, shared, private)
}

func checkDirFunctions(t *testing.T, baseDir client.CreateDirInfo, shared bool, private bool) {
	// Grab all directories from the base and make sure this new one is there (created by calling function)
	getDir, err := safeClient.GetDir(client.GetDirInfo{DirPath: "/", Shared: shared})
	require.NoError(t, err)
	var foundDir *client.DirInfo
	for _, dir := range getDir.SubDirs {
		if dir.Name == strings.TrimPrefix(baseDir.DirPath, "/") {
			foundDir = &dir
			break
		}
	}
	require.NotNil(t, foundDir)
	require.Equal(t, strings.TrimPrefix(baseDir.DirPath, "/"), foundDir.Name)
	require.Equal(t, baseDir.Private, foundDir.Private)
	require.Equal(t, baseDir.Versioned, foundDir.Versioned)
	require.Equal(t, baseDir.Metadata, foundDir.Metadata)
	require.Equal(t, foundDir.CreatedOn, foundDir.ModifiedOn)
	require.WithinDuration(t, time.Now(), foundDir.CreatedOn.Time(), 10*time.Second)

	// Grab just this directory and make sure it has no files and it's the exact same as what we had above
	getDir, err = safeClient.GetDir(client.GetDirInfo{DirPath: baseDir.DirPath, Shared: shared})
	require.NoError(t, err)
	require.Empty(t, getDir.Files)
	require.Empty(t, getDir.SubDirs)
	require.Exactly(t, *foundDir, getDir.Info)

	// Add a child directory, and make sure it looks right
	childName := randomName()
	childDirInfo := client.CreateDirInfo{
		DirPath:  path.Join(baseDir.DirPath, childName),
		Private:  private,
		Metadata: "",
		Shared:   shared,
	}
	require.NoError(t, safeClient.CreateDir(childDirInfo))
	getDir, err = safeClient.GetDir(client.GetDirInfo{DirPath: baseDir.DirPath, Shared: shared})
	require.NoError(t, err)
	require.Empty(t, getDir.Files)
	require.Len(t, getDir.SubDirs, 1)
	foundDir = &getDir.SubDirs[0]
	require.Equal(t, childName, foundDir.Name)
	require.Equal(t, baseDir.Private, foundDir.Private)
	require.Equal(t, baseDir.Versioned, foundDir.Versioned)
	require.Equal(t, baseDir.Metadata, foundDir.Metadata)
	require.Equal(t, foundDir.CreatedOn, foundDir.ModifiedOn)
	require.WithinDuration(t, time.Now(), foundDir.CreatedOn.Time(), 10*time.Second)

	// Change the child dir name and make sure it took
	newChildName := randomName()
	newChildDirPath := path.Join(baseDir.DirPath, newChildName)
	err = safeClient.ChangeDir(client.ChangeDirInfo{
		DirPath: childDirInfo.DirPath,
		Shared:  shared,
		NewName: newChildName,
	})
	getDir, err = safeClient.GetDir(client.GetDirInfo{DirPath: baseDir.DirPath, Shared: shared})
	require.NoError(t, err)
	require.Len(t, getDir.SubDirs, 1)
	require.Equal(t, newChildName, getDir.SubDirs[0].Name)
	require.Equal(t, foundDir.CreatedOn, getDir.SubDirs[0].CreatedOn)
	// TODO: https://maidsafe.atlassian.net/browse/CS-58
	//require.True(t, foundDir.ModifiedOn.Time().Before(getDir.SubDirs[0].ModifiedOn.Time()))

	// Delete the child directory and make sure it's gone
	require.NoError(t, safeClient.DeleteDir(client.DeleteDirInfo{DirPath: newChildDirPath, Shared: shared}))
	getDir, err = safeClient.GetDir(client.GetDirInfo{DirPath: baseDir.DirPath, Shared: shared})
	require.NoError(t, err)
	require.Empty(t, getDir.SubDirs)
}

func checkFileFunctions(t *testing.T, baseDir client.CreateDirInfo, shared bool, private bool) {
	// Create a file in base directory and make sure it's there
	fileName := randomName()
	fileInfo := client.CreateFileInfo{
		FilePath: path.Join(baseDir.DirPath, fileName),
		Shared:   shared,
		Metadata: "",
	}
	require.NoError(t, safeClient.CreateFile(fileInfo))
	getDir, err := safeClient.GetDir(client.GetDirInfo{DirPath: baseDir.DirPath, Shared: shared})
	require.NoError(t, err)
	require.Len(t, getDir.Files, 1)
	require.Empty(t, getDir.SubDirs)
	require.Equal(t, fileName, getDir.Files[0].Name)
	require.Equal(t, int64(0), getDir.Files[0].Size)
	require.Equal(t, fileInfo.Metadata, getDir.Files[0].Metadata)
	// TODO: https://maidsafe.atlassian.net/browse/CS-59
	//require.Equal(t, getDir.Files[0].CreatedOn, getDir.Files[0].ModifiedOn)
	require.WithinDuration(t, time.Now(), getDir.Files[0].CreatedOn.Time(), 10*time.Second)

	// Change the name and confirm changed
	newFileName := randomName()
	newFilePath := path.Join(baseDir.DirPath, newFileName)
	err = safeClient.ChangeFile(client.ChangeFileInfo{
		FilePath: fileInfo.FilePath,
		Shared:   shared,
		NewName:  newFileName,
	})
	getDir, err = safeClient.GetDir(client.GetDirInfo{DirPath: baseDir.DirPath, Shared: shared})
	require.NoError(t, err)
	require.Len(t, getDir.Files, 1)
	require.Equal(t, newFileName, getDir.Files[0].Name)

	// Write "FOO BAR BAZ"
	err = safeClient.WriteFile(client.WriteFileInfo{
		FilePath: newFilePath,
		Shared:   shared,
		Contents: ioutil.NopCloser(strings.NewReader("FOO BAR BAZ")),
	})
	require.NoError(t, err)

	// Make sure the size and dates are right
	getDir, err = safeClient.GetDir(client.GetDirInfo{DirPath: baseDir.DirPath, Shared: shared})
	require.NoError(t, err)
	require.Len(t, getDir.Files, 1)
	require.Equal(t, int64(11), getDir.Files[0].Size)
	require.True(t, getDir.Files[0].CreatedOn.Time().Before(getDir.Files[0].ModifiedOn.Time()))

	// Full content has to be right
	rc, err := safeClient.GetFile(client.GetFileInfo{FilePath: newFilePath, Shared: shared})
	require.NoError(t, err)
	requireReadCloserEqualsString(t, "FOO BAR BAZ", rc)

	// Pull out just "O BAR B"
	rc, err = safeClient.GetFile(client.GetFileInfo{
		FilePath: newFilePath,
		Shared:   shared,
		Offset:   2,
		Length:   7,
	})
	require.NoError(t, err)
	requireReadCloserEqualsString(t, "O BAR B", rc)

	// Change BAR to QUX
	err = safeClient.WriteFile(client.WriteFileInfo{
		FilePath: newFilePath,
		Shared:   shared,
		Contents: ioutil.NopCloser(strings.NewReader("QUX")),
		Offset:   4,
	})
	require.NoError(t, err)

	// Re-check the content
	rc, err = safeClient.GetFile(client.GetFileInfo{FilePath: newFilePath, Shared: shared})
	require.NoError(t, err)
	requireReadCloserEqualsString(t, "FOO QUX BAZ", rc)

	// Delete the file and make sure it's gone
	err = safeClient.DeleteFile(client.DeleteFileInfo{FilePath: newFilePath, Shared: shared})
	require.NoError(t, err)
	getDir, err = safeClient.GetDir(client.GetDirInfo{DirPath: baseDir.DirPath, Shared: shared})
	require.NoError(t, err)
	require.Empty(t, getDir.Files)
}

func checkMoveFunctions(t *testing.T, baseDir client.CreateDirInfo, shared bool, private bool) {
	// We're gonna create a directory and a file in it with the contents "Hello world" then we're gonna move the
	// whole thing under another directory... (we test file moves later on)

	// Create the dir and file
	firstDirName := randomName()
	firstDirInfo := client.CreateDirInfo{
		DirPath:  path.Join(baseDir.DirPath, firstDirName),
		Private:  private,
		Metadata: "",
		Shared:   shared,
	}
	require.NoError(t, safeClient.CreateDir(firstDirInfo))
	firstFileName := randomName()
	firstFileInfo := client.CreateFileInfo{
		FilePath: path.Join(firstDirInfo.DirPath, firstFileName),
		Shared:   shared,
		Metadata: "",
	}
	require.NoError(t, safeClient.CreateFile(firstFileInfo))
	err := safeClient.WriteFile(client.WriteFileInfo{
		FilePath: firstFileInfo.FilePath,
		Shared:   shared,
		Contents: ioutil.NopCloser(strings.NewReader("Test file!")),
	})
	require.NoError(t, err)

	// Make sure it's there as expected with nothing else
	// First check base dir
	getDir, err := safeClient.GetDir(client.GetDirInfo{DirPath: baseDir.DirPath, Shared: shared})
	require.NoError(t, err)
	require.Empty(t, getDir.Files)
	require.Len(t, getDir.SubDirs, 1)
	require.Equal(t, firstDirName, getDir.SubDirs[0].Name)
	// Now check the subdir
	getDir, err = safeClient.GetDir(client.GetDirInfo{DirPath: firstDirInfo.DirPath, Shared: shared})
	require.NoError(t, err)
	require.Empty(t, getDir.SubDirs)
	require.Len(t, getDir.Files, 1)
	require.Equal(t, firstFileName, getDir.Files[0].Name)

	// Now let's make a directory to move that entire directory under
	secondDirName := randomName()
	secondDirInfo := client.CreateDirInfo{
		DirPath:  path.Join(baseDir.DirPath, secondDirName),
		Private:  private,
		Metadata: "",
		Shared:   shared,
	}
	require.NoError(t, safeClient.CreateDir(secondDirInfo))

	// Move it and do not retain source
	// TODO: test moving from shared to non-shared and vice-versa
	err = safeClient.MoveDir(client.MoveDirInfo{
		SrcPath:      firstDirInfo.DirPath,
		SrcShared:    shared,
		DestPath:     secondDirInfo.DirPath,
		DestShared:   shared,
		RetainSource: false,
	})
	require.NoError(t, err)

	// Now re-obtain the base dir and make sure only our new dir is there (the other is moved under it)
	getDir, err = safeClient.GetDir(client.GetDirInfo{DirPath: baseDir.DirPath, Shared: shared})
	require.NoError(t, err)
	require.Empty(t, getDir.Files)
	require.Len(t, getDir.SubDirs, 1)
	require.Equal(t, secondDirName, getDir.SubDirs[0].Name)

	// Obtain the new dir and make sure it only has the dir we moved under it
	getDir, err = safeClient.GetDir(client.GetDirInfo{DirPath: secondDirInfo.DirPath, Shared: shared})
	require.NoError(t, err)
	require.Empty(t, getDir.Files)
	require.Len(t, getDir.SubDirs, 1)
	require.Equal(t, firstDirName, getDir.SubDirs[0].Name)

	// Obtain the directory created under the new dir to make sure our file is still in there and
	// everything is the same
	getDir, err = safeClient.GetDir(client.GetDirInfo{
		DirPath: path.Join(secondDirInfo.DirPath, firstDirName),
		Shared:  shared,
	})
	require.NoError(t, err)
	require.Empty(t, getDir.SubDirs)
	require.Len(t, getDir.Files, 1)
	require.Equal(t, firstFileName, getDir.Files[0].Name)
	// Check the content too
	rc, err := safeClient.GetFile(client.GetFileInfo{
		FilePath: path.Join(secondDirInfo.DirPath, firstDirName, firstFileName),
		Shared:   shared,
	})
	require.NoError(t, err)
	requireReadCloserEqualsString(t, "Test file!", rc)

	// Cool, now let's test file movement...let's move that file deep under there to the very root dir leaving an
	// empty dir where it was
	err = safeClient.MoveFile(client.MoveFileInfo{
		SrcPath:      path.Join(secondDirInfo.DirPath, firstDirName, firstFileName),
		SrcShared:    shared,
		DestPath:     path.Join(baseDir.DirPath, firstFileName),
		DestShared:   shared,
		RetainSource: false,
	})
	require.NoError(t, err)

	// Now make sure it is there in the root
	getDir, err = safeClient.GetDir(client.GetDirInfo{DirPath: baseDir.DirPath, Shared: shared})
	require.NoError(t, err)
	require.Len(t, getDir.Files, 1)
	require.Equal(t, firstFileName, getDir.Files[0].Name)
	// Check the content too
	rc, err = safeClient.GetFile(client.GetFileInfo{
		FilePath: path.Join("/", firstFileName),
		Shared:   shared,
	})
	require.NoError(t, err)
	requireReadCloserEqualsString(t, "Test file!", rc)

	// Confirm that deep subdirectory is empty
	getDir, err = safeClient.GetDir(client.GetDirInfo{
		DirPath: path.Join(secondDirInfo.DirPath, firstDirName),
		Shared:  shared,
	})
	require.NoError(t, err)
	require.Empty(t, getDir.SubDirs)
	require.Empty(t, getDir.Files)
}
