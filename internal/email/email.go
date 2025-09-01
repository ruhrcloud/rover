package email

import (
	"bytes"
	"context"
	"crypto/tls"
	"errors"
	"fmt"
	"io"
	"log"
	"mime"
	"path"
	"strings"
	"text/template"
	"time"

	"github.com/emersion/go-imap"
	"github.com/emersion/go-imap/client"
	message "github.com/emersion/go-message"
	"github.com/emersion/go-message/mail"
	"github.com/gosimple/slug"
	"golang.org/x/net/html/charset"

	"github.com/ruhrcloud/rover/internal/config"
	webdav "github.com/ruhrcloud/rover/internal/webdav"
)

type Result struct {
	Processed           int
	SkippedRecipient    int
	MsgsNoAttachments   int
	MsgsWithAttachments int
	TotalParts          int
	Uploaded            int
	SeenToMark          []uint32
}

func init() {
	message.CharsetReader = charset.NewReaderLabel
}

func Once(ctx context.Context, t config.Task, dav *webdav.Client) (Result, error) {
	var res Result
	
	opts := t.From
	conn, err := client.Dial(opts.Host)
	if err != nil {
		return res, err
	}
	defer conn.Logout()

	tlsConfig := &tls.Config{ServerName: "mail.your-server.de"}
	err = conn.StartTLS(tlsConfig)
	if err != nil {
		return res, err
	}

	err = conn.Login(opts.User, opts.Pass)
	if err != nil {
		return res, err
	}

	_, err = conn.Select(opts.Mailbox, false)
	if err != nil {
		return res, err
	}

	crit := imap.NewSearchCriteria()
	if t.Filter.Seen != nil {
		if *t.Filter.Seen {
			crit.WithFlags = []string{imap.SeenFlag}
		} else {
			crit.WithoutFlags = []string{imap.SeenFlag}
		}
	}

	uids, err := conn.UidSearch(crit)
	if err != nil {
		return res, err
	}
	if len(uids) == 0 {
		log.Printf("[%s] no messages matched criteria", t.Name)
		return res, nil
	}

	seq := new(imap.SeqSet)
	seq.AddNum(uids...)
	section := &imap.BodySectionName{}
	items := []imap.FetchItem{
		imap.FetchEnvelope,
		imap.FetchUid,
		section.FetchItem(),
	}
	msgCh := make(chan *imap.Message, 16)
	done := make(chan error, 1)
	go func() {
		done <- conn.UidFetch(seq, items, msgCh)
	}()

	tmpl := template.New("name").Funcs(
		template.FuncMap{
			"slug": slug.Make,
		})
	_, err = tmpl.Parse(t.Format)
	if err != nil {
		return res, fmt.Errorf("failed to parse format %w", err)
	}

	parts := []string{}
	if v := strings.Trim(t.Path, "/"); v != "" {
		parts = append(parts, v)
	}
	for _, tg := range t.Tags {
		if v := strings.Trim(strings.TrimSpace(tg), "/"); v != "" {
			parts = append(parts, v)
		}
	}
	relDir := strings.Trim(strings.Join(parts, "/"), "/")
	if relDir != "" {
		if err := dav.Mkdir(ctx, relDir); err != nil {
			return res, fmt.Errorf("mkdir %q: %w", "/"+relDir, err)
		}
	}

	allowed := map[string]struct{}{}
	for _, e := range t.Filter.Extensions {
		e = strings.ToLower(strings.TrimPrefix(strings.TrimSpace(e), "."))
		if e != "" {
			allowed[e] = struct{}{}
		}
	}
	byExt := len(allowed) > 0
	dec := new(mime.WordDecoder)

	for msg := range msgCh {
		if msg == nil || msg.Envelope == nil {
			continue
		}
		en := msg.Envelope
		if len(t.Filter.Recipients) > 0 && !recipientsMatch(en, t.Filter.Recipients) {
			// log.Printf("[%s] UID %d: skip (recipient filter)", t.Name, msg.Uid)
			res.SkippedRecipient++
			continue
		}
		r := msg.GetBody(section)
		if r == nil {
			res.Processed++
			continue
		}
		mr, err := mail.CreateReader(r)
		if err != nil {
			res.Processed++
			continue
		}
		subj := en.Subject
		if s, err := dec.DecodeHeader(subj); err == nil && s != "" {
			subj = s
		}
		when := en.Date
		if when.IsZero() {
			when = time.Now()
		}

		att := 0
		up := 0

		for {
			p, err := mr.NextPart()
			if errors.Is(err, io.EOF) {
				break
			}
			if err != nil {
				break
			}
			ah, ok := p.Header.(*mail.AttachmentHeader)
			if !ok {
				continue
			}
			name, _ := ah.Filename()
			if name == "" {
				name = "attachment.bin"
			}
			ext := strings.TrimPrefix(strings.ToLower(path.Ext(name)), ".")
			if byExt {
				if _, ok := allowed[ext]; !ok {
				  // log.Printf("[%s] UID %d: skip attachment %q by extension", t.Name, msg.Uid, name)
					continue
				}
			}
			att++
			base := strings.TrimSuffix(name, "."+ext)

			buf := new(bytes.Buffer)
			if _, err := io.Copy(buf, p.Body); err != nil {
				continue
			}

			data := struct {
				OrigBase string
				OrigExt  string
				Subject  string
				UID      uint32
				Date     string
				DateTime string
			}{
				OrigBase: base,
				OrigExt:  "." + ext,
				Subject:  subj,
				UID:      msg.Uid,
				Date:     when.Format("2006-01-02"),
				DateTime: when.Format("20060102-150405"),
			}
			var b bytes.Buffer
			if err := tmpl.Execute(&b, data); err != nil || b.Len() == 0 {
				return res, fmt.Errorf("template execution failed")
			}

			out := strings.TrimSpace(b.String())
			out = strings.ReplaceAll(out, "/", "-")
			out = strings.ReplaceAll(out, "\\", "-")
			if path.Ext(out) == "" && ext != "" {
				out += "." + ext
			}
			target := out
			if relDir != "" {
				target = webdav.Join(relDir, out)
			}
			if exists, _ := dav.Exists(ctx, target); exists {
				dir := path.Dir(target)
				bn := path.Base(target)
				be := strings.TrimSuffix(bn, path.Ext(bn))
				target = webdav.Join(dir, be+"-dup."+ext)
			}
			if err := dav.Create(ctx, target, buf.Bytes()); err != nil {
				log.Printf("[%s] UID %d: upload /%s failed: %v", t.Name, msg.Uid, target, err)
				continue
			}
			log.Printf("[%s] uploaded -> /%s", t.Name, target)
			up++
			res.Uploaded++
		}

		if att == 0 {
			res.MsgsNoAttachments++
		} else {
			res.MsgsWithAttachments++
			res.TotalParts += att
		}
		if up > 0 && t.MarkSeen {
			res.SeenToMark = append(res.SeenToMark, msg.Uid)
		}
		res.Processed++
	}
	if err := <-done; err != nil {
		return res, err
	}
	return res, nil
}

func MarkSeen(ctx context.Context, t config.Task, uids []uint32) error {
	if len(uids) == 0 {
		return nil
	}
	cli, err := client.DialTLS(t.From.Host, &tls.Config{ServerName: strings.Split(t.From.Host, ":")[0]})
	if err != nil {
		return err
	}
	defer cli.Logout()
	if err := cli.Login(t.From.User, t.From.Pass); err != nil {
		return err
	}
	if _, err := cli.Select(t.From.Mailbox, false); err != nil {
		return err
	}
	seq := new(imap.SeqSet)
	seq.AddNum(uids...)
	item := imap.FormatFlagsOp(imap.AddFlags, true)
	flags := []interface{}{imap.SeenFlag}
	return cli.UidStore(seq, item, flags, nil)
}

func recipientsMatch(env *imap.Envelope, want []string) bool {
	m := map[string]struct{}{}
	for _, w := range want {
		if s := strings.ToLower(strings.TrimSpace(w)); s != "" {
			m[s] = struct{}{}
		}
	}
	check := func(l []*imap.Address) bool {
		for _, a := range l {
			if a == nil {
				continue
			}
			if _, ok := m[strings.ToLower(strings.TrimSpace(a.Address()))]; ok {
				return true
			}
		}
		return false
	}
	return check(env.To) || check(env.Cc) || check(env.Bcc)
}

