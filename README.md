# Forwarding HTTP/S Proxy

# Background

This server is useful, if one wants to have originating requests to a
destination service from a set of well-known IPs.


## Moving closer

If you are hosting your code as apps on Heroku, one option to move closer to a
fixed IP address could be to host your code on an EC2 instance and connect it an
EIP (Elastic IP).

## Proxy

To proxy HTTPS requests, one seems to have two options in software: Use the
[HTTP Tunnel](https://en.wikipedia.org/wiki/HTTP_tunnel) feature via the
[CONNECT](https://www.ietf.org/rfc/rfc2817.txt) method, also called a
[Forwarding Proxy](https://en.wikipedia.org/wiki/Proxy_server), or a [Reverse
Proxy](https://en.wikipedia.org/wiki/Reverse_proxy). There are hardware
solutions on OSI layer 3 instead of layer 7, namely a NAT proxy, but this is not
discussed here as we don't have access to physical hardware or want to invest
into a NAT proxy e.g. on AWS.


### Forwarding Proxy

A forwarding proxy can come in two flavours:

It either first terminates an incoming client request, evaluates it, and
forwards the request to a destination. This works for HTTP as well as for HTTPS.
A subtle but important side-effect of using a fowarding proxy for HTTPS is that
it would terminate the request, thus being able to inspect the request's content
(and modify it).

The alternative is to use tunneling via the `CONNECT` method. By this, a proxy
accepts an initial CONNECT request entailing the entire URL as `HOST` value,
rather than just the host address. The proxy then opens up a TCP connection to
the destination and transparently forwards the raw communication from the client
to the destination. This comes with the subtle difference that only the initial
`CONNECT` request from the client to the proxy is terminated and can be
analyzed, however any further communication is not terminated nor intercepted
thus SSL communication can't be read by the proxy.

One additional subtle thing to mention is that forwarding proxies using
tunneling rely on clients to understand and comply to the HTTP tunneling RFC
2817 and thus have to be explicitly configured to use HTTP/S proxying, usually
picking up the proxy url from environment varialbles such as `HTTP_PROXY` and
`HTTPS_PROXY` respectively. For Go, see
[net/http/Transport](https://golang.org/pkg/net/http/#Transport) and


### Reverse Proxy

A reverse proxy accepts incoming requests from clients and routes them to a
specific destination based on the request. Discrimators for destination could be
the host, path, query parameters, any header even the body. A reverse proxy in
any case intercepts and terminates the HTTP and HTTPS connections and creates
new requests to the destinations.


# Alternatives

One can consider alternatives in protocol and product, as well as _Make vs Buy_.

Alternatives to a forwarding proxy would be a reverse proxy or NAT proxy. Revese
proxies, as outlined above would require more routing logic. A NAT proxy would
require more network and hardware configuration even if it's abstracted by IaaS
providers.

Alternatives products would be [Squid](http://squid-cache.org),
[NGinx](https://www.nginx.com), [HAProxy](http://haproxy.org),
[Varnish](http://varnish-cache.org), [TinyProxy](https://tinyproxy.github.io),
etc. Nginx is primarily meant to be used as a reverse proxy and I had difficulty
massaging it into acting as a forwarding proxy. HAProxy is similar to Nginx
meant to be used as reverse proxy as well as HTTP cache. Varnish is primarily
meant to be used as HTTP cache. Squid and TinyProxy are closest to be working as
forwarding proxies, however I had difficulties setting them up on a new AMI
Linux 2 EC2 instance, and TinyProxy is not maintained anymore since several
years.

Alternatives in _Make vs Buy_ would be:

Heroku's [Private Spaces](https://www.heroku.com/private-spaces) feature stable
outbound IPs:

>Stable outbound IPs
>
>Securely connect apps to third party cloud services and corporate networks.

This would weight in with a price tag of either $1000 or $3000 respectively for
GDPR compliance.

Another _Make vs Buy_ alternative would be hosted Heroku's Add-ons that would
offer stable IP support such as
[Fixie](https://elements.heroku.com/addons/fixie),
[Guru301](https://elements.heroku.com/addons/guru301),
[Proximo](https://elements.heroku.com/addons/proximo), [QuotaGuard
Static](https://elements.heroku.com/addons/quotaguardstatic),

_Fixie_ and _Guru301_ only support Heroku's _United States_ region, not Europe.
_QuotaGuard Static_ and _Proximo_ both support Heroku's _Europe_ region and
would offer 20,000 respectively 50,000 requests per month for $19 respectively
$25 per month.


# What's missing

Compared to especially hosted solutions, insight into the proxies operations
such as logging, monitoring, usage statistics needs to be added if desired.
Additionally one has to setup the binary as a reliable server and automate
deployments.


# Implementation details

This is NIH (_Not Invented Here_ syndrome), thus quality and feature set is not
en-par with hosted or off-the-shelf solutions.

It is a simple HTTPS tunneling proxy that starts a Go HTTPS server at a given
port awaiting `CONNECT` requests, basically dropping everything else. To start
the HTTPS server one has to provide a server certificate and private key for the
TLS handshake phase.

Once a client requests a `CONNECT` it will create a TCP connection to the
provided destination host, and on successfully establishing this connection,
hijack the original client connection, and transparently and bidirectionally
copying incoming and outgoing TCP byte streams.

It has minimal logging using Uber's Zap logger.
