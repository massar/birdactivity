# BirdActivity

BirdActivity is a simple combination of [gofeed](https://github.com/mmcdole/gofeed) for reading/parsing RSS/Atom/JSON feeds and [go-mastodon](https://github.com/mattn/go-mastodon/tree/master) for posting them.

This allows one to fetch a RSS feed from a URL and feed the items contained to a Mastodon feed.

This can be useful for converting news pages into Mastodon posts, or, if one has a RSS feed of a Birdsite account, to convert those into Mastodon posts.

## Example configuration

For the configuration file a simple JSON file is used (typically named `config.json`).

The format with all the fields are shown below.

One specifies a HTTP User-Agent with the `useragent` option. Please always include a contact address in the User-Agent, so that one can be contacted in case the bot does something odd, it might happen ;)

One can specify a single Mastodon Server per config file (but could have multiple config files in different runs).
The `mstdnclientid` and `mstdnclientsecret` options define the Client ID and Secret as defined on the given Mastodon server.

Then one defines one or more accounts on the Mastodon server.

For each account one defines a local `name` for easy indentification in the logs or for the `only` command line option.
One defines if the account is `active` allowing one to temporarily disable it.
The `feedurl` defines the URL where to fetch the Atom, RSS or JSON feeds from.
The `postfield` argument is used to determine which content from the retrieved field should be used for posting.
The `mstdnuser` (typically the user's email address) and `mstdnpass` fields are used to authenticate the user to the Mastodon server so that it can post to the given account.

```
{
	"useragent": "birdactivity (operator@example.net)",
	"mstdnserver": "https://mastadon.example.net",
	"mstdnclientid": "<Mastadon Client ID>",
	"mstdnclientsecret": "<Masadon Client Secret>",
	"cachedir": "/var/lib/birdactivity/cache",
	"accounts": [
		{
			"name": "firstaccount",
			"active": true,
			"feedurl": "https://www.example.com/feeds.rss",
			"postfield": "title",
			"mstdnuser": "first@example.net",
			"mstdnpass": "XNQ41CS0-A-LONG-PASSWORD-HOcuQI32ER3gAJT9"
		},
		{
			"name": "secondaccount",
			"active": false,
			"feedurl": "https://other.example.org/feeds.rss",
			"postfield": "titlecontent",
			"mstdnuser": "second@example.net",
			"mstdnpass": "p6aoB58V-A-DIFFERENT-LONG-PASSWORD-IagzFS4h"
		}
	]
}
```

Do configure the Mastodon account with the "Bot" flag to indicate that it is a bot / auto-generated content account.

## Crontab

I personally run this from Crontab every 10 minutes; maybe at a later point I'll simply add a simple timer which can delay randomly and run as a daemon.

## Example usages

This tool is used by the author for amongst others:

 - Unofficial [Mastodon Releases](https://secluded.ch/@mastodonreleases/), so that I know when a new version is out, this sources the [Mastodon Releases Page on GitHub](https://github.com/mastodon/mastodon/releases) which also has a `.atom` mode if one appends to the URL.
 - Unofficial [Init7 Mastodon Mirror](https://secluded.ch/@init7) = [Init7](https://init7.net) Twitter feed, so that I do not have to look at hthe birdsite anymore

