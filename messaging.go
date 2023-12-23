package main

import (
	"archive/zip"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"

	"github.com/PaulSonOfLars/gotgbot/v2"
)

type messageType int

//go:generate stringer -type=messageType
const (
	textMessage messageType = iota
	animationMessage
	audioMessage
	documentMessage
	photoMessage
	stickerMessage
	videoMessage
	videoNoteMessage
	voiceMessage
	fileUploadMessage
	directoryMessage
)

var typeInfo = map[messageType]struct {
	method     string   // The Telegram API method to use for sending this type of message
	extensions []string // File extensions that map to this type
	text       string   // the field in the message request to use for the message text, if any
	file       string   // the field in the message request to use for the file, if any
}{
	textMessage: {
		method: "sendMessage",
		text:   "text",
	},
	animationMessage: {
		method:     "sendAnimation",
		extensions: []string{".gif"},
		text:       "caption",
		file:       "animation",
	},
	audioMessage: {
		method:     "sendAudio",
		extensions: []string{".mp3", ".m4a"},
		text:       "caption",
		file:       "audio",
	},
	documentMessage: {
		method: "sendDocument",
		text:   "caption",
		file:   "document",
	},
	photoMessage: {
		method:     "sendPhoto",
		extensions: []string{".jpg", ".jpeg", ".png"},
		text:       "caption",
		file:       "photo",
	},
	stickerMessage: {
		method:     "sendSticker",
		extensions: []string{".webp"},
		file:       "sticker",
	},
	videoMessage: {
		method:     "sendVideo",
		extensions: []string{".mp4"},
		text:       "caption",
		file:       "video",
	},
	videoNoteMessage: {
		method: "sendVideoNote",
		file:   "video_note",
	},
	voiceMessage: {
		method:     "sendVoice",
		extensions: []string{".ogg", ".oga"},
		file:       "voice",
	},
}

var extensions map[string]messageType

func init() {
	extensions = make(map[string]messageType)
	for typ, info := range typeInfo {
		for _, ext := range info.extensions {
			extensions[ext] = typ
		}
	}
}

type Message struct {
	messageType messageType
	detected    bool // True if messageType was automatically detected
	text        string
	filePath    string
	noUpload    bool // True if the file should not be uploaded to external services, even if too large for Telegram
}

func (msg *Message) Send(bot *gotgbot.Bot, chatID int64) error {
	if msg.messageType == directoryMessage {
		zipPath, remove, err := createArchive(msg.filePath)
		if remove != nil {
			defer remove()
		}

		if err != nil {
			return fmt.Errorf("failed to create archive: %w", err)
		}
		msg.filePath = zipPath

		stat, err := os.Lstat(zipPath)
		if err != nil {
			return fmt.Errorf("failed to stat zip file: %w", err)
		}
		size := stat.Size()
		if size < fileSizeLimit || msg.noUpload {
			msg.messageType = documentMessage
		} else {
			msg.messageType = fileUploadMessage
		}
	}

	if msg.messageType == fileUploadMessage {
		url, err := uploadFile(msg.filePath)
		if err != nil {
			return fmt.Errorf("failed to upload file: %w", err)
		}

		if msg.text != "" {
			msg.text += "\n"
		}
		msg.text += url
		msg.filePath = ""
		msg.messageType = textMessage
	}

	typ := typeInfo[msg.messageType]

	// Constructing the *Opts structs for each message type is a bit of a pain, it's easier to just use the lower-level Request API here.
	params := map[string]string{
		"chat_id": fmt.Sprint(chatID),
	}
	var data map[string]gotgbot.NamedReader

	if msg.filePath != "" {
		f, err := os.Open(msg.filePath)
		if err != nil {
			return fmt.Errorf("failed to open file: %w", err)
		}
		defer f.Close()
		params[typ.file] = "attach://" + typ.file
		data = map[string]gotgbot.NamedReader{
			typ.file: f,
		}
	}

	if typ.text != "" {
		params[typ.text] = msg.text
	}

	_, err := bot.Request(typ.method, params, data, nil)
	if err != nil {
		return fmt.Errorf("failed to send message: %w", err)
	}
	return nil
}

// uploadFile uploads a (potentially large) file to transfer.sh
func uploadFile(filePath string) (string, error) {
	f, err := os.Open(filePath)
	if err != nil {
		return "", fmt.Errorf("failed to open file: %w", err)
	}
	defer f.Close()

	name := url.PathEscape(f.Name())
	req, err := http.NewRequest("PUT", "https://transfer.sh/"+name, f)
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("User-Agent", "Tell (https://github.com/mikolysz/tell)")
	req.Header.Set("Content-Type", "application/octet-stream")

	resp, err := http.DefaultClient.Do(req)
	if err != nil || resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("failed to upload file to transfer.sh: %w", err)
	}
	defer resp.Body.Close()

	url, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read transfer.sh response: %w", err)
	}

	return string(url), nil
}

// createArchive creates a zip archive of a directory
func createArchive(srcDirPath string) (zipPath string, remove func(), err error) {
	tmpdir, err := os.MkdirTemp("", "tell")
	if err != nil {
		return "", nil, fmt.Errorf("failed to create temporary file: %w", err)
	}

	remove = func() {
		os.RemoveAll(tmpdir)
	}

	srcDirPath, err = filepath.Abs(srcDirPath)
	if err != nil {
		return "", remove, fmt.Errorf("failed to get absolute path: %w", err)
	}

	zipPath = filepath.Join(tmpdir, filepath.Base(srcDirPath)+".zip")
	zipFile, err := os.Create(zipPath)
	if err != nil {
		return "", remove, fmt.Errorf("failed to create zip file: %w", err)
	}
	defer zipFile.Close()

	zipWriter := zip.NewWriter(zipFile)
	defer zipWriter.Close()

	if err := filepath.Walk(srcDirPath, func(srcFilePath string, info os.FileInfo, err error) error {
		if err != nil {
			return fmt.Errorf("failed to walk directory: %w", err)
		}

		destPath, err := filepath.Rel(srcDirPath, srcFilePath)
		if err != nil {
			return fmt.Errorf("failed to get relative path: %w", err)
		}

		// Package zip distinguishes files and directories by whether the path ends in a '/'.
		if info.IsDir() && !strings.HasSuffix(destPath, "/") {
			destPath += "/"
		}

		fh, err := zip.FileInfoHeader(info)
		if err != nil {
			return fmt.Errorf("Failed to create file header: %w", err)
		}

		fh.Name = destPath
		fh.Method = zip.Deflate
		fileWriter, err := zipWriter.CreateHeader(fh)
		if err != nil {
			return fmt.Errorf("failed to create zip file: %w", err)
		}

		if info.IsDir() {
			return nil
		}

		srcFile, err := os.Open(srcFilePath)
		if err != nil {
			return fmt.Errorf("failed to open file: %w", err)
		}
		defer srcFile.Close()

		_, err = io.Copy(fileWriter, srcFile)
		if err != nil {
			return fmt.Errorf("failed to copy file to zip: %w", err)
		}

		return nil
	}); err != nil {
		return "", remove, fmt.Errorf("failed to walk directory: %w", err)
	}

	return zipPath, remove, nil
}
