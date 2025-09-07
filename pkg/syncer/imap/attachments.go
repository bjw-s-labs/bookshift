package imap

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"io"
	"log/slog"
	"mime/quotedprintable"
	"os"
	"path"
	"path/filepath"
	"slices"
	"strings"

	"github.com/bjw-s-labs/bookshift/pkg/util"
	"github.com/emersion/go-imap/v2"
	"github.com/emersion/go-imap/v2/imapclient"
)

// messageAttachmentPart captures details of a single attachment part within a message.
type messageAttachmentPart struct {
	part           []int
	filename       string
	attachmentSize uint32
	encoding       string
}

// DownloadAttachments downloads all valid attachments to dstFolder and optionally deletes the message.
func (im *ImapMessage) DownloadAttachments(dstFolder string, validExtensions []string, overwriteExistingFile bool, removeMessageAfterDownload bool) error {
	// Fetch basic message information from the server
	message, err := im.imapClient.fetchByUID(im.uid, &imap.FetchOptions{
		Envelope:      true,
		BodyStructure: &imap.FetchItemBodyStructure{Extended: true},
	})
	if err != nil {
		return err
	}

	var messageSender string
	for _, addr := range message.Envelope.From {
		messageSender = fmt.Sprintf("%s (%s@%s)", addr.Name, addr.Mailbox, addr.Host)
	}
	messageSubject := message.Envelope.Subject

	// Find message attachment parts
	msgAttachmentParts, err := im.determineAttachmentParts(message, validExtensions)
	if err != nil {
		return err
	}

	// Create target folder if required
	if _, err := os.Stat(dstFolder); os.IsNotExist(err) {
		if util.DryRun {
			slog.Info("[dry-run] Would create local folder", "folder", dstFolder)
			return nil
		}
		slog.Info("Creating local folder", "folder", dstFolder)
		if err := os.MkdirAll(dstFolder, 0755); err != nil {
			return err
		}
	}

	// Download the attachments
	for _, msgAttachmentPart := range msgAttachmentParts {
		safeFileName := util.SafeFileName(msgAttachmentPart.filename)
		dstPath := filepath.Join(dstFolder, safeFileName)

		slog.Info("Downloading email attachment", "host", im.imapClient.Host, "sender", messageSender, "subject", messageSubject, "filename", msgAttachmentPart.filename)

		message, err := im.imapClient.fetchByUID(im.uid, &imap.FetchOptions{
			BodySection: []*imap.FetchItemBodySection{{Part: msgAttachmentPart.part}},
		})
		if err != nil {
			return err
		}

		for _, section := range message.BodySection {
			// Check if the file already exists
			_, err := os.Stat(dstPath)
			if !os.IsNotExist(err) {
				if !overwriteExistingFile {
					slog.Warn("File already exists, skipping download", "file", dstPath)
					continue // skip this attachment only
				}

				slog.Info("Overwriting existing file", "file", dstPath)
			}

			// Download the file
			if util.DryRun {
				slog.Info("[dry-run] Would download email attachment", "uid", im.uid, "filename", safeFileName, "destination", dstPath)
				continue
			}
			tmpFile, err := os.CreateTemp(dstFolder, "bookshift-")
			if err != nil {
				return err
			}
			defer tmpFile.Close()

			slog.Debug("Downloading to temporary file", "file", tmpFile.Name())
			writer := util.NewFileWriter(tmpFile, int64(msgAttachmentPart.attachmentSize), true)
			// Decode according to transfer-encoding
			var src io.Reader
			switch strings.ToLower(msgAttachmentPart.encoding) {
			case "base64":
				src = base64.NewDecoder(base64.StdEncoding, bytes.NewReader(section.Bytes))
			case "quoted-printable":
				src = quotedprintable.NewReader(bytes.NewReader(section.Bytes))
			case "7bit", "8bit", "binary", "":
				// pass through as-is
				src = bytes.NewReader(section.Bytes)
			default:
				// unknown: pass-through but warn
				slog.Warn("Unknown transfer-encoding, writing raw bytes", "encoding", msgAttachmentPart.encoding)
				src = bytes.NewReader(section.Bytes)
			}

			if _, err := io.Copy(writer, src); err != nil {
				return err
			}

			if err := tmpFile.Sync(); err != nil { // ensure data flushed before rename
				os.Remove(tmpFile.Name())
				return err
			}
			if err := os.Rename(tmpFile.Name(), dstPath); err != nil {
				os.Remove(tmpFile.Name())
				return err
			}

			slog.Info("Successfully downloaded attachment", "uid", im.uid, "filename", safeFileName, "path", dstPath)
		}
	}

	if removeMessageAfterDownload {
		if util.DryRun {
			slog.Info("[dry-run] Would delete message from server", "uid", im.uid)
		} else {
			if err := im.DeleteFromServer(); err != nil {
				return err
			}
		}
	}

	return nil
}

// determineAttachmentParts walks the body structure to find attachment parts matching validExtensions.
func (im *ImapMessage) determineAttachmentParts(msg *imapclient.FetchMessageBuffer, validExtensions []string) ([]*messageAttachmentPart, error) {
	if msg == nil {
		return nil, fmt.Errorf("could not determine message attachment parts")
	}

	var messageAttachmentParts []*messageAttachmentPart

	msg.BodyStructure.Walk(
		imap.BodyStructureWalkFunc(func(part []int, partObj imap.BodyStructure) (walkChildren bool) {
			// Detect attachment parts
			if partObj.Disposition() != nil && partObj.Disposition().Params["filename"] != "" {
				attachmentPart := part
				attachmentFilename := partObj.Disposition().Params["filename"]
				attachmentFileExtension := path.Ext(attachmentFilename)

				// Only continue if the file extension is wanted
				if len(validExtensions) > 0 && !slices.Contains(validExtensions, attachmentFileExtension) {
					return false
				}

				// Find out the attachment size and encoding
				var attachmentSize uint32
				var encoding string
				switch p := partObj.(type) {
				case *imap.BodyStructureSinglePart:
					attachmentSize = p.Size
					// Encoding is a string-like type; stringify safely
					encoding = strings.ToLower(fmt.Sprintf("%s", p.Encoding))
					if encoding == "" {
						// Default to base64 for attachments when not specified; aligns with previous behavior and tests
						encoding = "base64"
					}
				}

				messageAttachmentParts = append(messageAttachmentParts, &messageAttachmentPart{
					part:           attachmentPart,
					filename:       attachmentFilename,
					attachmentSize: attachmentSize,
					encoding:       encoding,
				})
			}
			return true
		}),
	)

	return messageAttachmentParts, nil
}
