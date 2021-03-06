package vcs

import (
	"io/ioutil"
	"time"
	//"log"
	"os"
	"testing"
)

// Canary test to ensure BzrRepo implements the Repo interface.
var _ Repo = &BzrRepo{}

// To verify bzr is working we perform integration testing
// with a known bzr service. Due to the long time of repeatedly checking out
// repos these tests are structured to work together.

func TestBzr(t *testing.T) {

	tempDir, err := ioutil.TempDir("", "go-vcs-bzr-tests")
	if err != nil {
		t.Error(err)
	}
	defer func() {
		err = os.RemoveAll(tempDir)
		if err != nil {
			t.Error(err)
		}
	}()

	repo, err := NewBzrRepo("https://launchpad.net/govcstestbzrrepo", tempDir+"/govcstestbzrrepo")
	if err != nil {
		t.Error(err)
	}

	if repo.Vcs() != Bzr {
		t.Error("Bzr is detecting the wrong type")
	}

	// Check the basic getters.
	if repo.Remote() != "https://launchpad.net/govcstestbzrrepo" {
		t.Error("Remote not set properly")
	}
	if repo.LocalPath() != tempDir+"/govcstestbzrrepo" {
		t.Error("Local disk location not set properly")
	}

	//Logger = log.New(os.Stdout, "", log.LstdFlags)

	// Do an initial clone.
	err = repo.Get()
	if err != nil {
		t.Errorf("Unable to clone Bzr repo. Err was %s", err)
	}

	// Verify Bzr repo is a Bzr repo
	if repo.CheckLocal() == false {
		t.Error("Problem checking out repo or Bzr CheckLocal is not working")
	}

	// Test internal lookup mechanism used outside of Bzr specific functionality.
	ltype, err := DetectVcsFromFS(tempDir + "/govcstestbzrrepo")
	if err != nil {
		t.Error("detectVcsFromFS unable to Bzr repo")
	}
	if ltype != Bzr {
		t.Errorf("detectVcsFromFS detected %s instead of Bzr type", ltype)
	}

	// Test NewRepo on existing checkout. This should simply provide a working
	// instance without error based on looking at the local directory.
	nrepo, nrerr := NewRepo("https://launchpad.net/govcstestbzrrepo", tempDir+"/govcstestbzrrepo")
	if nrerr != nil {
		t.Error(nrerr)
	}
	// Verify the right oject is returned. It will check the local repo type.
	if nrepo.CheckLocal() == false {
		t.Error("Wrong version returned from NewRepo")
	}

	err = repo.UpdateVersion("2")
	if err != nil {
		t.Errorf("Unable to update Bzr repo version. Err was %s", err)
	}

	// Use Version to verify we are on the right version.
	v, err := repo.Version()
	if v != "2" {
		t.Error("Error checking checked out Bzr version")
	}
	if err != nil {
		t.Error(err)
	}

	// Use Date to verify we are on the right commit.
	d, err := repo.Date()
	if d.Format(longForm) != "2015-07-31 09:50:42 -0400" {
		t.Error("Error checking checked out Bzr commit date")
	}
	if err != nil {
		t.Error(err)
	}

	// Perform an update.
	err = repo.Update()
	if err != nil {
		t.Error(err)
	}

	v, err = repo.Version()
	if v != "3" {
		t.Error("Error checking checked out Bzr version")
	}
	if err != nil {
		t.Error(err)
	}

	tags, err := repo.Tags()
	if err != nil {
		t.Error(err)
	}
	if tags[0] != "1.0.0" {
		t.Error("Bzr tags is not reporting the correct version")
	}

	branches, err := repo.Branches()
	if err != nil {
		t.Error(err)
	}
	if len(branches) != 0 {
		t.Error("Bzr is incorrectly returning branches")
	}

	if repo.IsReference("1.0.0") != true {
		t.Error("Bzr is reporting a reference is not one")
	}

	if repo.IsReference("foo") == true {
		t.Error("Bzr is reporting a non-existant reference is one")
	}

	if repo.IsDirty() == true {
		t.Error("Bzr incorrectly reporting dirty")
	}

	ci, err := repo.CommitInfo("3")
	if err != nil {
		t.Error(err)
	}
	if ci.Commit != "3" {
		t.Error("Bzr.CommitInfo wrong commit id")
	}
	if ci.Author != "Matt Farina <matt@mattfarina.com>" {
		t.Error("Bzr.CommitInfo wrong author")
	}
	if ci.Message != "Updated Readme with pointer." {
		t.Error("Bzr.CommitInfo wrong message")
	}
	ti, err := time.Parse(time.RFC1123Z, "Fri, 31 Jul 2015 09:51:37 -0400")
	if err != nil {
		t.Error(err)
	}
	if !ti.Equal(ci.Date) {
		t.Error("Bzr.CommitInfo wrong date")
	}

	_, err = repo.CommitInfo("asdfasdfasdf")
	if err != ErrRevisionUnavailable {
		t.Error("Bzr didn't return expected ErrRevisionUnavailable")
	}
}

func TestBzrCheckLocal(t *testing.T) {
	// Verify repo.CheckLocal fails for non-Bzr directories.
	// TestBzr is already checking on a valid repo
	tempDir, err := ioutil.TempDir("", "go-vcs-bzr-tests")
	if err != nil {
		t.Error(err)
	}
	defer func() {
		err = os.RemoveAll(tempDir)
		if err != nil {
			t.Error(err)
		}
	}()

	repo, _ := NewBzrRepo("", tempDir)
	if repo.CheckLocal() == true {
		t.Error("Bzr CheckLocal does not identify non-Bzr location")
	}

	// Test NewRepo when there's no local. This should simply provide a working
	// instance without error based on looking at the remote localtion.
	_, nrerr := NewRepo("https://launchpad.net/govcstestbzrrepo", tempDir+"/govcstestbzrrepo")
	if nrerr != nil {
		t.Error(nrerr)
	}
}
