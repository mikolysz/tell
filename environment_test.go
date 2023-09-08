package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"strings"
	"testing"
)

func TestValidArguments(t *testing.T) {
	valid := []string{
		"hello",
		"-t token",
		"-a",
		"-f testdata/foo",
		"-f testdata/foo Message caption",
		"-f testdata/foo --file-type audio",
		"-f testdata/foo --file-type photo My caption",
	}

	for _, v := range valid {
		t.Run("Args: "+v, func(t *testing.T) {
			args := strings.Split(v, " ")
			_, err := initEnvironment(args)

			if err != nil {
				t.Errorf("validation failed for valid arguments: '%s': %s", v, err)
			}
		})
	}
}

func TestInvalidArguments(t *testing.T) {
	invalid := []string{
		"-t token hello", // both a token and a message
		"-a Hello",       // both a message and the authorize flag
		"-f testdata/foo --file-type no_such_type",
		"--file-type photo caption", // There's a file type, but not a file.
		"--no-upload",               // There's no file, so we can't upload it.
		"-f testdata/foo --no-upload --file-type upload",
		"-f testdata/foo --file-type voice This is a caption, but voice messages don't support captions.",
	}

	for _, iv := range invalid {
		t.Run("Args: "+iv, func(t *testing.T) {
			args := strings.Split(iv, " ")
			_, err := initEnvironment(args)
			if err == nil {
				t.Errorf("validation succeeded for invalid arguments: '%s'", iv)
			}
		})
	}
}

func TestFiletypeDetection(t *testing.T) {
	_, err := initEnvironment([]string{"-f", "this_file_does_not_exist"})
	if err == nil {
		t.Errorf("validation succeeded for non-existent file")
	}

	env, err := initEnvironment([]string{"-f", "testdata/foo"})
	if err != nil {
		t.Errorf("validation failed for existing file: %s", err)
	}

	if env.msg.messageType != documentMessage {
		t.Errorf("wrong type detected for file without extension, expected document, got %s", env.msg.messageType)
	}

	env, err = initEnvironment([]string{"-f", "testdata/foo", "--file-type", "audio"})
	if err != nil {
		t.Errorf("validation failed for existing file %s when file type was explicitly specified", err)
	}

	if env.msg.messageType != audioMessage {
		t.Errorf("wrong type detected for file with explicit audio type, expected audio, got %s", env.msg.messageType)
	}

	env, err = initEnvironment([]string{"-f", "testdata"})
	if err != nil {
		t.Errorf("validation failed for directory: %s", err)
	}

	if env.msg.messageType != directoryMessage {
		t.Errorf("wrong type detected for directory, expected folder, got %s", env.msg.messageType)
	}

	file, cleanup, err := tempFile("jpg", 42)
	if err != nil {
		t.Errorf("failed to create temporary photo file: %s", err)
	}
	defer cleanup()

	env, err = initEnvironment([]string{"-f", file})
	if err != nil {
		t.Errorf("validation failed for photo: %s", err)
	}

	if env.msg.messageType != photoMessage {
		t.Errorf("wrong type detected for photo, expected photo, got %s", env.msg.messageType)
	}

	file, cleanup, err = tempFile("jpg", photoSizeLimit+1)
	if err != nil {
		t.Errorf("failed to create temporary photo file: %s", err)
	}
	defer cleanup()

	env, err = initEnvironment([]string{"-f", file})
	if err != nil {
		t.Errorf("validation failed for large photo: %s", err)
	}
	if env.msg.messageType != documentMessage {
		t.Errorf("wrong type detected for photo exceeding file size limit, expected document, got %s", env.msg.messageType)
	}

	file, cleanupISO, err := tempFile("iso", fileSizeLimit+1)
	if err != nil {
		t.Errorf("failed to create large file: %s", err)
	}
	defer cleanupISO()

	env, err = initEnvironment([]string{"-f", file})
	if err != nil {
		t.Errorf("validation failed for large file: %s", err)
	}
	if env.msg.messageType != fileUploadMessage {
		t.Errorf("wrong type detected for file exceeding file size limit, expected file upload, got %s", env.msg.messageType)
	}

	env, err = initEnvironment([]string{"-f", file, "--file-type", "audio"})
	if err != nil {
		t.Errorf("validation failed for large file with explicit file type: %s", err)
	}
	if env.msg.messageType != fileUploadMessage {
		t.Errorf("Type not set to upload for large file with manual type override, expected file upload, got %s", env.msg.messageType)
	}

	env, err = initEnvironment([]string{"-f", file, "--no-upload"})
	if err == nil {
		t.Errorf("validation succeeded for large file with no-upload flag")
	}
}

// tempFile creates a temporary file with the given size and extension.
// The file is filled with garbage contents.
// It returns the path to the file and a function that can be used to remove it.
func tempFile(extension string, sizeBytes int) (string, func(), error) {
	f, err := ioutil.TempFile("", "tell_test_*."+extension)
	if err != nil {
		return "", nil, fmt.Errorf("failed to create temporary file: %w", err)
	}

	remove := func() {
		_ = os.Remove(f.Name())
	}

	// Fill the file with garbage.
	_, err = f.Write(make([]byte, sizeBytes))
	if err != nil {
		remove()
		return "", nil, fmt.Errorf("failed to write to temporary file: %w", err)
	}

	if err := f.Close(); err != nil {
		remove()
		return "", nil, fmt.Errorf("failed to close temporary file: %w", err)
	}

	return f.Name(), remove, nil
}
