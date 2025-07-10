package main

func main() {
	domains := []string{
		"localhost:3002", "localhost:3000", "crashloop.sh:443", "biznesbees.com:443",
		"google.com:443", "github.com:443", "stackoverflow.com:443", "microsoft.com:443",
		"apple.com:443", "amazon.com:443", "facebook.com:443", "twitter.com:443",
		"linkedin.com:443", "reddit.com:443", "youtube.com:443", "cloudflare.com:443",
		"mozilla.org:443", "wikipedia.org:443", "bing.com:443", "yahoo.com:443",
		"duckduckgo.com:443", "baidu.com:443", "yandex.com:443", "zoom.us:443",
		"slack.com:443", "dropbox.com:443", "adobe.com:443", "oracle.com:443",
		"ibm.com:443", "intel.com:443", "netflix.com:443", "paypal.com:443",
		"salesforce.com:443", "shopify.com:443", "spotify.com:443", "twitch.tv:443",
		"wordpress.com:443", "tumblr.com:443", "quora.com:443", "pinterest.com:443",
		"imgur.com:443", "vimeo.com:443", "bitbucket.org:443", "gitlab.com:443",
		"heroku.com:443", "digitalocean.com:443", "linode.com:443", "ovh.com:443",
		"godaddy.com:443", "namecheap.com:443", "gandi.net:443", "fastly.com:443",
		"akamai.com:443", "verisign.com:443", "letsencrypt.org:443", "sectigo.com:443",
		"globalsign.com:443", "digicert.com:443", "thawte.com:443", "symantec.com:443",
		"comodo.com:443", "rapidssl.com:443", "geotrust.com:443", "startssl.com:443",
		"trustwave.com:443", "mail.ru:443", "protonmail.com:443", "zoho.com:443",
		"icloud.com:443", "outlook.com:443", "office.com:443", "skype.com:443",
		"weibo.com:443", "tencent.com:443", "alibaba.com:443", "jd.com:443",
		"booking.com:443", "airbnb.com:443", "expedia.com:443", "tripadvisor.com:443",
		"uber.com:443", "lyft.com:443", "doordash.com:443", "grubhub.com:443",
		"instacart.com:443", "coursera.org:443", "edx.org:443", "udemy.com:443",
		"pluralsight.com:443", "khanacademy.org:443", "bbc.com:443", "cnn.com:443",
		"nytimes.com:443", "theguardian.com:443", "forbes.com:443", "bloomberg.com:443",
		"reuters.com:443", "wsj.com:443", "ft.com:443", "economist.com:443",
	}
	CheckCertsforAllPeers(domains, 4, true)
}
