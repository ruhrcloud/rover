package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"

	"github.com/emersion/go-imap"
	"github.com/emersion/go-imap/client"
	"github.com/emersion/go-message/mail"
)

type Imap struct {
	Host string `json:"host"`
	User string `json:"user"`
	Pass string `json:"pass"`
}

type Webdav struct {
	Baseurl string `json:"baseurl"`
	User    string `json:"user"`
	Pass    string `json:"pass"`
}

type Config struct {
	Imap   Imap   `json:"imap"`
	Webdav Webdav `json:"webdav"`
}

func main() {
	var file string
	flag.StringVar(&file, "config", "config.json", "path to the config file")
	flag.Parse()

	f, err := os.Open(file)
	if err != nil {
		log.Fatal("Error opening config:", err)
	}
	defer f.Close()

	byt, err := io.ReadAll(f)
	if err != nil {
		log.Fatal("Error reading config:", err)
	}

	var conf Config
	json.Unmarshal(byt, &conf)

	client, err := client.DialTLS(conf.Imap.Host, nil)
	if err != nil {
		log.Fatal("Error connecting to IMAP server:", err)
	}
	defer client.Logout()

	if err := client.Login(conf.Imap.User, conf.Imap.Pass); err != nil {
		log.Fatal("IMAP login failed:", err)
	}
	log.Println("Logged in to IMAP server")

	mbox, err := client.Select("INBOX", false)
	if err != nil {
		log.Fatal("Unable to select INBOX:", err)
	}
	log.Printf("Mailbox %s selected. Total messages: %d\n", mbox.Name, mbox.Messages)

	criteria := imap.NewSearchCriteria()
	criteria.WithoutFlags = []string{"\\Seen"}
	uids, err := client.Search(criteria)
	if err != nil {
		log.Fatal("Search error:", err)
	}
	if len(uids) == 0 {
		log.Println("No unread messages found.")
		return
	}
	log.Println("Found message UIDs:", uids)

	seqset := new(imap.SeqSet)
	seqset.AddNum(uids...)
	section := &imap.BodySectionName{}
	messages := make(chan *imap.Message, 10)
	done := make(chan error, 1)

	go func() {
		done <- client.Fetch(seqset, []imap.FetchItem{imap.FetchEnvelope, section.FetchItem()}, messages)
	}()

	for msg := range messages {
		if msg.Envelope == nil {
			continue
		}
		log.Printf("Processing message UID %d: %s\n", msg.Uid, msg.Envelope.Subject)

		r := msg.GetBody(section)
		if r == nil {
			log.Println("No message body returned for UID", msg.Uid)
			continue
		}

		mr, err := mail.CreateReader(r)
		if err != nil {
			log.Println("Failed to create mail reader for UID", msg.Uid, ":", err)
			continue
		}

		for {
			part, err := mr.NextPart()
			if err == io.EOF {
				break
			}
			if err != nil {
				log.Println("Error reading part:", err)
				break
			}

			if h, ok := part.Header.(*mail.AttachmentHeader); ok {
				filename, _ := h.Filename()
				log.Println("Found attachment:", filename)

				buf := new(bytes.Buffer)
				if _, err := io.Copy(buf, part.Body); err != nil {
					log.Println("Error reading attachment:", err)
					continue
				}

				uploadURL := fmt.Sprintf("%s%s", conf.Webdav.Baseurl, filename)
				log.Println("Uploading to:", uploadURL)

				req, err := http.NewRequest("PUT", uploadURL, bytes.NewReader(buf.Bytes()))
				if err != nil {
					log.Println("Error creating HTTP request:", err)
					continue
				}
				req.SetBasicAuth(conf.Webdav.User, conf.Webdav.Pass)

				resp, err := http.DefaultClient.Do(req)
				if err != nil {
					log.Println("Error uploading attachment:", err)
					continue
				}
				resp.Body.Close()

				if resp.StatusCode >= 200 && resp.StatusCode < 300 {
					log.Printf("Attachment %s uploaded successfully.\n", filename)
				} else {
					log.Printf("Failed to upload %s: %s\n", filename, resp.Status)
				}
			}
		}
	}

	if err := <-done; err != nil {
		log.Fatal("Fetch error:", err)
	}

	log.Println("Processing complete.")
}
