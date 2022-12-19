package main

import (
	"context"
	"fmt"

	"github.com/mattn/go-mastodon"
)

// trash 'trashes' an account by removing all the posts it has made
func trash(cfg *Cfg, totrash string) (err error) {
	var acct *Account = nil

	// Find the named account (can only trash one at a time, which is a 'safety' feature :) )
	for m := 0; m < len(cfg.Accounts); m++ {
		acct = &cfg.Accounts[m]

		if acct.Name != totrash {
			acct = nil
			continue
		}

		break
	}

	if acct == nil {
		err = fmt.Errorf("No such account: %s", totrash)
		return
	}

	fmt.Printf("Trashing account %s...\n", totrash)

	// Authenticate to Mastadon...
	mc, err := doAuth(cfg, acct)
	if err != nil {
		err = fmt.Errorf("Authentication Failed: %s", err)
		return
	}

	// Handle Pagination, there can be multiple pages of posts
	var pg mastodon.Pagination

	for {
		fmt.Printf("Fetching Timeline... maxID='%s'\n", pg.MaxID)

		// Get the Timeline for the user at the given page
		timeline, err2 := mc.GetTimelineHome(context.Background(), &pg)
		if err2 != nil {
			err = err2
			break
		}

		// Cleanup each item we get
		for i := len(timeline) - 1; i >= 0; i-- {
			fmt.Printf("Cleaning: %s\n", timeline[i].ID)

			err = mc.DeleteStatus(context.Background(), timeline[i].ID)
			if err != nil {
				fmt.Printf("== err: %s\n", err)
			} else {
				fmt.Printf("== ok\n")
			}
		}

		// Keep track of what we are deleting...
		// this is still a bit funny, as deletes make the pagination go odd ;)
		// might have to CTRL-C if needed...
		fmt.Printf("Checking maxID='%s' (len=%d)\n", pg.MaxID, len(timeline))
		if pg.MaxID == "" {
			break
		}

		pg.SinceID = ""
		pg.MinID = ""
	}

	fmt.Printf("Trashing complete...\n")
	return
}
