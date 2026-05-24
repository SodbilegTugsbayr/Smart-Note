package noteman

import "testing"

func TestPrepareResponseSetsHasFileFromFilePath(t *testing.T) {
	note := &Note{FilePath: " /tmp/material.pdf "}

	note.PrepareResponse()

	if !note.HasFile {
		t.Fatal("PrepareResponse() should set HasFile when FilePath is present")
	}
}

func TestPrepareResponseClearsHasFileWhenFilePathBlank(t *testing.T) {
	note := &Note{FilePath: " \n\t "}

	note.PrepareResponse()

	if note.HasFile {
		t.Fatal("PrepareResponse() should clear HasFile when FilePath is blank")
	}
}
