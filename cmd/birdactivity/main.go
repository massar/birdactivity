package main

import (
	"bytes"
	"context"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/mattn/go-mastodon"
	"github.com/microcosm-cc/bluemonday"
	"github.com/mmcdole/gofeed"
)

// maxActivityPostLength determines maximum length of a ActivityPub Post
const maxActivityPostLength = 450

// dodebug determines if we output Debug info
var dodebug bool

// noop determines if changes (posts) are made
var noop bool

// Debug is a quick function to output Debug messages based on the `dodebug` flag
// If an account is provided it is prefixed before the messages for easier identification
// which account the debug message pertains to
func Debug(acct *Account, format string, args ...interface{}) {
	if !dodebug {
		return
	}

	if acct != nil {
		args = append([]interface{}{acct.Name}, args...)
		fmt.Printf("[%s] "+format, args...)
	} else {
		fmt.Printf(format, args...)
	}
	return
}

// fetchURL retrieves a URL with the provided UserAgent
func fetchURL(cfg *Cfg, url string) (body []byte, err error) {
	client := &http.Client{}

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return
	}

	// Set the User-Agent so that we are identifiable
	// Should generally be in the format "BirdActivity (contact@example.org)"
	req.Header.Set("User-Agent", cfg.UserAgent)

	// Retrieve the URL
	resp, err := client.Do(req)
	if err != nil {
		return
	}

	defer resp.Body.Close()

	// Return the complete body and error if there
	return ioutil.ReadAll(resp.Body)
}

// doAuth authenticates against a Mastodon server
func doAuth(cfg *Cfg, acct *Account) (mc *mastodon.Client, err error) {
	mc = mastodon.NewClient(&mastodon.Config{
		Server:       cfg.MstdnServer,
		ClientID:     cfg.MstdnClientID,
		ClientSecret: cfg.MstdnClientSecret,
	})

	err = mc.Authenticate(context.Background(), acct.MstdnUser, acct.MstdnPass)
	return
}

// textCleanup strips HTML from a text
// this to avoid posting HTML which is not accepted in ActivityPub
func textCleanup(in string) (out string) {
	p := bluemonday.NewPolicy()
	p.AllowAttrs("href").OnElements("a")
	p.AllowURLSchemes("mailto", "http", "https")
	p.RequireParseableURLs(true)
	out = p.Sanitize(in)

	// Convert links from <a href="LINK">LINKTO</a> => LINK - LINKTO
	var re = regexp.MustCompile(`<a href="(https?://[^\s]+)">([^<]*)</a>`)
	out = re.ReplaceAllString(out, `$1 - $2`)

	return
}

// doAccount handles the core function of BirdActivity
// for a given account:
// - read the cache from cache/<account>.cache
// - fetch the RSS/Atom from the URL
// - store that RSS as cache/<account>.feed
// - parse the feed
// - for each item, check if in cache, and post if not
// - store the cache
func doAccount(cfg *Cfg, acct *Account) {
	var mc *mastodon.Client
	authed := false

	// The filename of the ID cache file
	cf := filepath.Join(cfg.CacheDir, acct.Name+".cache")
	Debug(acct, "Loading Cache %s\n", cf)
	cache, err := loadCache(cf)
	if err != nil {
		log.Fatal(err)
		return
	}

	// Fetch the feed
	Debug(acct, "Fetching from %s\n", acct.FeedURL)
	feedbody, err := fetchURL(cfg, acct.FeedURL)
	if err != nil {
		log.Fatal(err)
		return
	}

	// Store feed into Feed File
	rf := filepath.Join(cfg.CacheDir, acct.Name+".feed")
	Debug(acct, "Writing Feed to %s\n", rf)
	err = ioutil.WriteFile(rf, feedbody, 0644)
	if err != nil {
		log.Fatal(err)
		return
	}

	// Parse the feed
	gf := gofeed.NewParser()
	feed, err := gf.Parse(bytes.NewReader(feedbody))
	if err != nil {
		log.Fatal(err)
		return
	}

	// For each item, check cache and post
	for i := len(feed.Items) - 1; i >= 0; i-- {
		itm := feed.Items[i]

		// Is it already procesed?
		_, ok := cache.GUIDs[itm.GUID]
		if ok {
			// Already listed
			continue
		}

		// What field to use for posting?
		txt := itm.Title
		switch acct.PostField {
		case "title", "":
			// The default
			break

		case "content":
			txt = itm.Content
			break

		case "titlecontent":
			txt = itm.Title + " " + itm.Content
			break

		default:
			err = fmt.Errorf("Invalid postfield for %s: \"%s\" (expected: title|content|titlecontent)", acct.Name, acct.PostField)
			return
		}

		// Ensure there is no HTML
		txt = textCleanup(txt)

		// Trim Spaces and internal duplicate spaces
		txt = strings.TrimSpace(txt)
		pattern := regexp.MustCompile(`\s+`)
		txt = pattern.ReplaceAllString(txt, " ")

		// Limit length if post (ActivityPub is 500 chars)
		if len(txt) > maxActivityPostLength {
			txt = txt[0:maxActivityPostLength-4] + "..."
		}

		// Debug details, useful especially when doing 'noop' runs
		Debug(acct, "New entry ID \"%s\":\n8<-------------\n%s\n------------>8\n", itm.GUID, txt)

		// Do not actually post?
		if noop {
			continue
		}

		// If not authenticated yet, authenticate
		// this causes us only to auth once per account
		// and not auth if there are no new posts to post
		if !authed {
			mc, err = doAuth(cfg, acct)
			if err != nil {
				Debug(acct, "Authentication failed: %s\n", err)
				continue
			}
		}

		// Post the item from the feed
		_, err = mc.PostStatus(context.Background(), &mastodon.Toot{
			Status: txt,
		})
		if err != nil {
			Debug(acct, "Posting %s failed: %s\n", itm.GUID, err)
			continue
		}

		// If all went well, store the ID in the cache
		// so that we do not post it again
		cache.GUIDs[itm.GUID] = true
	}

	// Save the cache item, unless we are doing noop runs
	// (which actually do not change the cache but this way we do not update the file either)
	if !noop {
		saveCache(cf, cache)
	}

	return
}

// main is the program main
// it determines flags and does the actual work where needed
func main() {
	var onlyacct string
	var cfgfn string
	var totrash string

	// Figure out the flags the user called us with
	flag.StringVar(&onlyacct, "a", "", "Only run against this account")
	flag.StringVar(&cfgfn, "c", "config.json", "Configuration Filename")
	flag.StringVar(&totrash, "trash", "", "Trash all previous posts for a given account")
	flag.BoolVar(&dodebug, "debug", false, "Enable Debugging")
	flag.BoolVar(&noop, "noop", false, "No-Operation: fetch, parse, but do not post")
	flag.Parse()

	// Load the configuration from the config file (defaults to `config.json` above)
	cfg, err := loadConfig(cfgfn)
	if err != nil {
		log.Fatal(err)
		return
	}

	// Do we need to trash an account?
	// trashing is useful when we posted something wrong to an account
	// this will remove all posts; one can then also remove the cache
	// to fully re-import a feed in the next run
	if totrash != "" {
		err = trash(cfg, totrash)
		if err != nil {
			log.Fatal(err)
			return
		}

		return
	}

	// Normal posting run
	// For each account...
	for a := 0; a < len(cfg.Accounts); a++ {
		acct := &cfg.Accounts[a]

		// If we do not only want a given account
		if onlyacct != "" && onlyacct != acct.Name {
			Debug(nil, "a: %s vs o: %s - not it\n", acct.Name, onlyacct)
			continue
		}

		// Skip in-active accounts
		if !acct.Active && onlyacct == "" {
			Debug(nil, "a: %s vs o: %s - inactive\n", acct.Name, onlyacct)
			continue
		}

		// Do the account: fetch feed and post...
		doAccount(cfg, acct)
	}

	return
}
