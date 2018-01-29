package license

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"reflect"
	"testing"
)

func TestInit(t *testing.T) {
	lenFiles := len(fileNames) * len(fileExtensions)
	if len(fileTable) != lenFiles {
		t.Fatalf("fileTable not initialized: %#v", fileTable)
	}
	if len(DefaultLicenseFiles) != len(fileTable) {
		t.Fatalf("DefaultLicenseFiles not initialized: %#v", DefaultLicenseFiles)
	}

	if len(licenseTable) == 0 {
		t.Fatalf("licenseTable not initialized: %#v", licenseTable)
	}
	if len(KnownLicenses) != len(licenseTable) {
		t.Fatalf("KnownLicenses not initialized: %#v", KnownLicenses)
	}
}

func TestNewLicense(t *testing.T) {
	l := New("MyLicense", "Some license text.")
	if l.Type != "MyLicense" {
		t.Fatalf("bad license type: %s", l.Type)
	}
	if l.Text != "Some license text." {
		t.Fatalf("bad license text: %s", l.Text)
	}
}

func TestNewFromFile(t *testing.T) {
	lf := filepath.Join("fixtures", "licenses", "MIT")

	lh, err := os.Open(lf)
	if err != nil {
		t.Fatalf("err: %s", err)
	}
	defer lh.Close()

	licenseText, err := ioutil.ReadAll(lh)
	if err != nil {
		t.Fatalf("err: %s", err)
	}

	l, err := NewFromFile(lf)
	if err != nil {
		t.Fatalf("err: %s", err)
	}

	if l.Type != "MIT" {
		t.Fatalf("unexpected license type: %s", l.Type)
	}

	if l.Text != string(licenseText) {
		t.Fatalf("unexpected license text: %s", l.Text)
	}

	if l.File != lf {
		t.Fatalf("unexpected file path: %s", l.File)
	}
}

func TestNewFromFile_fails(t *testing.T) {
	// Fails if the file doesn't exist
	if _, err := NewFromFile("/unicorns/leprechauns"); err == nil {
		t.Fatalf("expected error loading non-existent file")
	}

	// Fails if license type from file is not guessable
	f, err := ioutil.TempFile("", "go-license")
	if err != nil {
		t.Fatalf("err: %s", err)
	}
	defer os.Remove(f.Name())
	defer f.Close()

	if _, err := f.WriteString("No license data"); err != nil {
		t.Fatalf("err: %v", err)
	}

	if _, err := NewFromFile(f.Name()); err == nil {
		t.Fatalf("expected error guessing license type from non-license file")
	}
}

func TestNewFromDir(t *testing.T) {
	d, err := ioutil.TempDir("", "go-license")
	if err != nil {
		t.Fatalf("err: %s", err)
	}
	defer os.RemoveAll(d)

	fPath := filepath.Join(d, "LICENSE")
	f, err := os.Create(fPath)
	if err != nil {
		t.Fatalf("err: %s", err)
	}

	lh, err := os.Open(filepath.Join("fixtures", "licenses", "MIT"))
	if err != nil {
		t.Fatalf("err: %s", err)
	}
	defer lh.Close()

	licenseText, err := ioutil.ReadAll(lh)
	if err != nil {
		t.Fatalf("err: %s", err)
	}

	if _, err := f.Write(licenseText); err != nil {
		t.Fatalf("err: %s", err)
	}

	l, err := NewFromDir(d)
	if err != nil {
		t.Fatalf("err: %s", err)
	}

	if l.Type != "MIT" {
		t.Fatalf("unexpected license type: %s", l.Type)
	}

	if l.Text != string(licenseText) {
		t.Fatalf("unexpected license text: %s", l.Text)
	}

	if l.File != fPath {
		t.Fatalf("unexpected file path: %s", l.File)
	}
}

func TestNewFromDir_fails(t *testing.T) {
	d, err := ioutil.TempDir("", "go-license")
	if err != nil {
		t.Fatalf("err: %v", err)
	}
	defer os.RemoveAll(d)

	// This file should be ignored
	if _, err := os.Create(filepath.Join(d, "nope")); err != nil {
		t.Fatalf("err: %v", err)
	}

	// Fails if the directory contains no license files
	if _, err := NewFromDir(d); err == nil {
		t.Fatalf("expected error loading empty directory")
	}

	// Fails if the directory does not exist
	if _, err := NewFromDir("go-license-nonexistent"); err == nil {
		t.Fatalf("expected error loading non-existent directory")
	}

	// Fails if multiple licenses are found. Also checks that casing is
	// ignored in the license file name.
	if _, err := os.Create(filepath.Join(d, "LICENSE.txt")); err != nil {
		t.Fatalf("err: %v", err)
	}
	if _, err := os.Create(filepath.Join(d, "copying.RST")); err != nil {
		t.Fatalf("err: %v", err)
	}
	_, err = NewFromDir(d)
	if err == nil || err != ErrMultipleLicenses {
		t.Fatalf("expect %q, got: %v", ErrMultipleLicenses, err)
	}

	// Fails if the directory specified is actually a file
	if _, err := NewFromDir(filepath.Join(d, "LICENCSE.txt")); err == nil {
		t.Fatalf("expected error loading file as directory")
	}
}

func TestLicenseRecognized(t *testing.T) {
	// Known licenses are recognized
	l := New("MIT", "The MIT License (MIT)")
	if !l.Recognized() {
		t.Fatalf("license was not recognized")
	}

	// Unknown licenses are not recognized
	l = New("None", "No license text")
	if l.Recognized() {
		t.Fatalf("fake license was recognized")
	}
}

func TestLicenseTypes(t *testing.T) {
	for ltype, _ := range licenseTable {
		file := filepath.Join("fixtures", "licenses", ltype)
		fh, err := os.Open(file)
		if err != nil {
			t.Fatalf("err: %s", err)
		}
		lbytes, err := ioutil.ReadAll(fh)
		if err != nil {
			t.Fatalf("err :%s", err)
		}
		ltext := string(lbytes)

		l := New("", ltext)
		if err := l.GuessType(); err != nil {
			t.Fatalf("err: %s", err)
		}
		if l.Type != ltype {
			t.Fatalf("\nexpected: %s\ngot: %s", ltype, l.Type)
		}
	}
}

func TestLicenseTypes_Abbreviated(t *testing.T) {
	// Abbreviated Apache 2.0 license is recognized
	l := New("", "http://www.apache.org/licenses/LICENSE-2.0")
	if err := l.GuessType(); err != nil {
		t.Fatalf("err: %s", err)
	}
	if l.Type != LicenseApache20 {
		t.Fatalf("\nexpected: %s\ngot: %s", LicenseApache20, l.Type)
	}
}

func TestNormalize(t *testing.T) {
	for in, expect := range map[string]string{
		"HeLlO wOrLd":    "hello world",
		"hello  world":   "hello world",
		"hello\r\nworld": "hello world",
		"hello\nworld":   "hello world",
		"hello, world":   "hello world",
	} {
		if out := normalize(in); out != expect {
			t.Fatalf("expect: %q, got: %q", expect, out)
		}
	}
}

func TestSearchDir(t *testing.T) {
	// Make a temp dir
	dir, err := ioutil.TempDir("", "go-license")
	if err != nil {
		t.Fatalf("err: %v", err)
	}
	defer os.RemoveAll(dir)

	// Create all the recognized license file names
	for _, name := range DefaultLicenseFiles {
		path := filepath.Join(dir, name)
		if err := ioutil.WriteFile(path, []byte{}, 0600); err != nil {
			t.Fatalf("err: %v", err)
		}
	}

	// Create a bogus file name (shouldn't be returned)
	err = ioutil.WriteFile(filepath.Join(dir, "nope"), []byte{}, 0600)
	if err != nil {
		t.Fatalf("err: %v", err)
	}

	// Create a directory. Should be skipped over.
	if err := os.Mkdir(filepath.Join(dir, "dir"), 0700); err != nil {
		t.Fatalf("err: %v", err)
	}

	// Find all the recognized files
	result, err := SearchDir(dir)
	if err != nil {
		t.Fatalf("err: %v", err)
	}

	// Check that we got the expected result
	if !reflect.DeepEqual(result, DefaultLicenseFiles) {
		t.Fatalf("expect: %v\nactual: %v", DefaultLicenseFiles, result)
	}
}
