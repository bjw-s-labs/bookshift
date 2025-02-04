package imap

import (
	"encoding/base64"
	"fmt"
	"log/slog"
	"os"
	"path"
	"slices"
	"strings"

	"github.com/bjw-s-labs/bookshift/pkg/util"
	"github.com/emersion/go-imap/v2"
	"github.com/emersion/go-imap/v2/imapclient"
)

type ImapMessage struct {
	uid        imap.UID
	imapClient *ImapClient
}

type messageAttachmentPart struct {
	part           []int
	filename       string
	attachmentSize uint32
}

func NewImapMessage(ic *ImapClient) *ImapMessage {
	return &ImapMessage{
		// ImapMessage: msg,
		imapClient: ic,
	}
}

func (im *ImapMessage) DownloadAttachments(dstFolder string, validExtensions []string, overwriteExistingFile bool, removeMessageAfterDownload bool) error {
	// Fetch basic message information from the server
	message, err := im.fetchByUID(im.uid, &imap.FetchOptions{
		Envelope:      true,
		BodyStructure: &imap.FetchItemBodyStructure{Extended: true},
	})
	if err != nil {
		return err
	}

	var messageSender string
	for _, addr := range message.Envelope.From {
		messageSender = strings.Join([]string{fmt.Sprintf("%s (%s@%s)", addr.Name, addr.Mailbox, addr.Host)}, "")
	}
	messageSubject := message.Envelope.Subject

	// Find message attachment parts
	msgAttachmentParts, err := im.determineAttachmentParts(message, validExtensions)
	if err != nil {
		return err
	}

	// Create target folder if required
	if _, err := os.Stat(dstFolder); os.IsNotExist(err) {
		slog.Info("Creating local folder", "folder", dstFolder)
		if err := os.MkdirAll(dstFolder, os.ModeDir|0755); err != nil {
			return err
		}
	}

	// Download the attachments
loopMsgAttachmentParts:
	for _, msgAttachmentPart := range msgAttachmentParts {
		safeFileName := util.SafeFileName(msgAttachmentPart.filename)
		dstPath := path.Join(dstFolder, safeFileName)

		slog.Info("Downloading email attachment", "host", im.imapClient.Host, "sender", messageSender, "subject", messageSubject, "filename", msgAttachmentPart.filename)

		message, err := im.fetchByUID(im.uid, &imap.FetchOptions{
			BodySection: []*imap.FetchItemBodySection{{Part: msgAttachmentPart.part}},
		})
		if err != nil {
			return err
		}

		for _, section := range message.BodySection {
			if section != nil {
				// Check if the file already exists
				_, err := os.Stat(dstPath)
				if !os.IsNotExist(err) {
					if !overwriteExistingFile {
						slog.Warn("File already exists, skipping download", "file", dstPath)
						break loopMsgAttachmentParts
					}

					slog.Info("Overwriting existing file", "file", dstPath)
				}

				// Download the file
				tmpFile, err := os.CreateTemp("", "bookshift-")
				if err != nil {
					os.Remove(tmpFile.Name())
					return err
				}
				defer tmpFile.Close()

				decodedContent, err := base64.StdEncoding.DecodeString(string(section))
				if err != nil {
					return err
				}

				writer := util.NewFileWriter(tmpFile, int64(msgAttachmentPart.attachmentSize), true)
				if _, err := writer.Write(decodedContent); err != nil {
					return err
				}
				os.Rename(tmpFile.Name(), dstPath)

				slog.Info("Succesfully downloaded attachment", "filename", safeFileName)
			}
		}
	}

	if removeMessageAfterDownload {
		im.DeleteFromServer()
	}

	return nil
}

func (im *ImapMessage) fetchByUID(uid imap.UID, options *imap.FetchOptions) (*imapclient.FetchMessageBuffer, error) {
	messages, err := im.imapClient.Client.Fetch(imap.UIDSetNum(uid), options).Collect()
	if err != nil {
		return nil, err
	}

	if len(messages) != 1 {
		return nil, fmt.Errorf("len(messages) = %v, want 1", len(messages))
	}

	return messages[0], nil
}

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
				attachmentFileExtension := strings.Trim(path.Ext(attachmentFilename), ".")

				// Only continue if the file extension is wanted
				if len(validExtensions) > 0 && !slices.Contains(validExtensions, attachmentFileExtension) {
					return false
				}

				// Find out the attachment size
				var attachmentSize uint32
				switch p := partObj.(type) {
				case *imap.BodyStructureSinglePart:
					attachmentSize = p.Size
				}

				messageAttachmentParts = append(messageAttachmentParts, &messageAttachmentPart{
					part:           attachmentPart,
					filename:       attachmentFilename,
					attachmentSize: attachmentSize,
				})
			}
			return true
		}),
	)

	return messageAttachmentParts, nil
}

func (im *ImapMessage) DeleteFromServer() error {
	if _, err := im.imapClient.Client.Store(imap.UIDSetNum(im.uid), &imap.StoreFlags{
		Op:    imap.StoreFlagsAdd,
		Flags: []imap.Flag{imap.FlagDeleted},
	}, nil).Collect(); err != nil {
		return err
	}

	if _, err := im.imapClient.Client.Expunge().Collect(); err != nil {
		return err
	}

	return nil
}
