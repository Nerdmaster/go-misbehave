# Go Misbehave!

This is a lame little server that poops out fake content for a given duration
or number of bytes (roughly) in order to fill up bad actors' spidering
databases with useless garbage.

## Basics

Compile this. You can use the Makefile, but it's just a stupid-simple shortcut
to building the binary in `bin/` (`go build -o ./bin/misbehave`).

The server can be configured to run for a given duration or max bytes,
whichever comes first - use `bin/misbehave -h` for details on the options you
can configure.

The byte limit is a soft limit. It really just says "once you exceed this, wrap
things up."

Running the server requires a wordlist in "common.txt". One word per line,
preferably common english words. The more words, the more randomness in your
output.

This is not a natural language anything. Its purpose is to be a low-CPU garbage
generator, not to build realistic-looking output.

## Fake HTML

The server poops out semi-valid HTML markup using a small list of tags and the
external wordlist. I suggest grabbing a wordlist like SCOWL or even just the
standard Linux spelling dictionary.

When the server is hit, it generates its page until it reaches a limit: either
it hits the maximum bytes or a time linit. Once that happens, the generator
stops and any still-open tags are closed. This allows the page to seem (to a
very stupid spider) to be a real page.

The HTML will occasionally have links to other "pages" using "/misbehave" as
the prefix, and a random word from the word list. e.g.:

    <a href="/misbehave/friendly/the/tire/lead">...</a>`

With a word list of even a mere 100 items, you can theoretically generate a
hundred million URLs. This helps keep spiders trapped longer, which lets you
waste more of their time while keeping your server happy.

Will it cost big spider companies anything to do this? No, realistically it
won't. Unless a lot of people do this....

## Server Load

The neat thing is that it takes a lot to do anything to your CPU. Directing
misbehaved spiders to "Go Misbehave" will very likely save you CPU compared to
just about anything other than static pages. It can really help protect
sluggish apps amazingly well.

The real issue you'll have, if you have any issues, will be bandwidth.
Depending on spider activity and your server specs, you may find you can't
afford to actually send this much data to misbehaving spiders.

## Apache configuration example

To make our misbehaving server intercept traffic from certain bots, we use
Apache like so:

```
# Redirect 'bytespider' user agents
RewriteEngine On
RewriteCond %{HTTP_USER_AGENT} bytespider [NC]
RewriteCond %{REQUEST_URI} !^/misbehave
RewriteRule ^/?(.*)$ https://my.server.com/misbehave/$1 [L,R=302]

# Misbehaving server
ProxyPass "/misbehave" "http://localhost:6000/misbehave" nocanon
ProxyPassReverse "/misbehave" "http://localhost:6000/misbehave"
```

This makes all "ByteSpider" user agents redirect to /misbehave, where the Go
application listens (via `mod_proxy` directives) and does its thing. The fake
URLs that "Go Misbehave" generates are handled properly by the proxy and
rewrite rules such that a spider will find itself following link after link of
nonsense.

Obviously you won't necessarily want to match only on user agent, but other
options that Apache can handle are better left to their docs, and many people
will use something better anyway, like nginx or HAProxy.

## Credits

- Señor Hoodoo came up with the concept
- ChatGPT did most of the coding
- Doogles built out some of the logic where it was easier to add/fix code than
  ask ChatGPT to do it
