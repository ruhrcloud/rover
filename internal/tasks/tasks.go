package tasks

import (
	"context"
	"log"
	"time"

	"github.com/ruhrcloud/rover/internal/config"
	"github.com/ruhrcloud/rover/internal/email"
	webdav "github.com/ruhrcloud/rover/internal/webdav"
)

func Run(ctx context.Context, cfg *config.Config) error {
	for i := range cfg.Tasks {
		t := cfg.Tasks[i]
		go loop(ctx, t)
	}

	<-ctx.Done()
	return nil
}

func loop(ctx context.Context, t config.Task) {
	duration, err := time.ParseDuration(t.Interval)
	if err != nil {
		duration = time.Duration(5 * float64(time.Minute))
	}
	log.Printf("[%s] starting task to run every %s", t.Name, duration)

	ticker := time.NewTicker(duration)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			opts := t.To
			client, err := webdav.New(webdav.Opts{
				BaseURL: opts.BaseURL,
				Auth:    opts.Auth,
				User:    opts.User,
				Pass:    opts.Pass,
				Token:   opts.Token,
			})
			if err != nil {
				log.Printf("[%s] webdav: %v", t.Name, err)
				continue
			}

			res, err := email.Once(ctx, t, client)
			if err != nil {
				log.Printf("[%s] %v", t.Name, err)
				continue
			}
			
			o := len(res.SeenToMark) > 0
			if o && t.MarkSeen {
				err := email.MarkSeen(ctx, t, res.SeenToMark)
				if err != nil {
					log.Printf("[%s] failed to mark seen: %v", t.Name, err)
				}
			}

			log.Printf("[%s] processed %d and uploaded %d attachments", 
				t.Name, res.Processed, res.Uploaded)
		}
	}
}

