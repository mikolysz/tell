package main

import (
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path"
	"strings"

	flag "github.com/spf13/pflag"
)

const (
	photoSizeLimit = 10 * 1024 * 1024
	fileSizeLimit  = 50 * 1024 * 1024
)

// environment contains the flags and arguments passed to the program
type environment struct {
	token            string
	authorizeNewUser bool
	noUpload         bool // If the file is too big, error out instead of uploading to transfer.sh

	msg Message
}

func initEnvironment(args []string) (*environment, error) {
	// set up and parse flags
	//
	// We continue on error here so that tests can run.
	// main catches the error, and we exit there.
	fs := flag.NewFlagSet("tell", flag.ContinueOnError)
	env := &environment{}

	fs.StringVarP(&env.token, "token", "t", "", "Save the provided Telegram bot token in the config file")
	fs.BoolVarP(&env.authorizeNewUser, "authorize-user", "a", false, "Authorize a new user to use the bot")
	fs.StringVarP(&env.msg.filePath, "file", "f", "", "Send the provided file")
	fileType := fs.String("file-type", "", "The type of file to send. One of: animation, audio, document, photo, sticker, video, video_note, voice or upload. Will be detected automatically if omitted")
	fs.BoolVarP(&env.noUpload, "no-upload", "n", false, "Do not upload files to transfer.sh if they are too big")

	if err := fs.Parse(args); err != nil {
		return nil, err
	}

	// Validate the provided flags and arguments.
	//
	// Get the message from the command line
	// This way, we can use tell like echo, without having to quote the message
	env.msg.text = strings.Join(fs.Args(), " ")

	if env.token != "" || env.authorizeNewUser {
		if env.msg.filePath != "" || env.msg.text != "" {
			return nil, fmt.Errorf("Cannot modify configuration and send a message at the same time")
		}
	}

	if env.msg.filePath == "" && *fileType != "" {
		return nil, fmt.Errorf("Filetype is present, but no file was specified.")
	}

	if env.msg.filePath == "" && env.noUpload {
		return nil, fmt.Errorf("Cannot use --no-upload without a file")
	}

	if env.msg.filePath == "" {
		env.msg.messageType = textMessage
		if env.msg.text == "" {
			// If we have no message, read from stdin.

			msg, err := io.ReadAll(os.Stdin)
			if err != nil {
				return nil, fmt.Errorf("failed to read from stdin: %w", err)
			}
			env.msg.text = string(msg)
		}

		return env, nil
	}

	detected, err := detectFileType(env.msg.filePath)
	if err != nil {
		return nil, err
	}

	if *fileType == "" {
		if detected == fileUploadMessage && env.noUpload {
			return nil, fmt.Errorf("file is too big to send via Telegram and --no-upload was specified")
		}
		env.msg.detected = true
		env.msg.messageType = detected
	} else {
		typ, err := messageTypeFromString(*fileType)
		if err != nil {
			return nil, err
		}

		if typ == fileUploadMessage && env.noUpload {
			return nil, fmt.Errorf("Cannot use --no-upload with file type 'upload'")
		}

		// detected is only equal to fileUploadMessage if the file is too large.
		if detected == fileUploadMessage && env.noUpload {
			return nil, fmt.Errorf("File too big to send via Telegram and --no-upload was specified")
		}

		if detected == fileUploadMessage {
			env.msg.messageType = fileUploadMessage // Too large to send with the chosen filetype, forced override
		} else {
			env.msg.messageType = typ
		}
	}

	if info, ok := typeInfo[env.msg.messageType]; ok {
		if info.text == "" && env.msg.text != "" {
			return nil, fmt.Errorf("sending text is not supported with message type %s", env.msg.messageType)
		}
		if info.file == "" && env.msg.filePath != "" {
			return nil, fmt.Errorf("sending files is not supported with message type %s", env.msg.messageType)
		}
	}

	return env, nil
}

func detectFileType(filePath string) (messageType, error) {
	stat, err := os.Stat(filePath)
	if errors.Is(err, fs.ErrNotExist) {
		return 0, fmt.Errorf("file %s does not exist", filePath)
	}

	if err != nil {
		return 0, fmt.Errorf("error when statting file: %w", err)
	}

	if stat.IsDir() {
		return directoryMessage, nil
	}

	extension := path.Ext(filePath)

	typ, ok := extensions[extension]
	if !ok {
		typ = documentMessage
	}

	if typ == photoMessage && stat.Size() > photoSizeLimit {
		typ = documentMessage
	}

	if stat.Size() > fileSizeLimit {
		typ = fileUploadMessage
	}

	return typ, nil
}

func messageTypeFromString(typ string) (messageType, error) {
	switch typ {
	case "animation":
		return animationMessage, nil
	case "audio":
		return audioMessage, nil
	case "document":
		return documentMessage, nil
	case "photo":
		return photoMessage, nil
	case "sticker":
		return stickerMessage, nil
	case "video":
		return videoMessage, nil
	case "video_note":
		return videoNoteMessage, nil
	case "voice":
		return voiceMessage, nil
	case "upload":
		return fileUploadMessage, nil
	case "folder":
		return directoryMessage, nil
	default:
		return textMessage, fmt.Errorf("Invalid file type: %s", typ)
	}
}
