package main

// Account defines an account as it will be configured in config.json
type Account struct {
	Name      string // A local definition for a name of this account (used for .rss + .cache files in the 'cache' directory)
	Active    bool   // If this account is active and should be fetched
	FeedURL   string // The RSS/Atom feed to be fetched from
	PostField string // The field from the feed to post (options: title, content, titlecontent)
	MstdnUser string // The username to login on the Mastodon server
	MstdnPass string // The password to use for logging in to the Mastodon server
}
