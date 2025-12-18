# ptsrv

*point to - service* - minimalist URL shortener and permalink database.

## Run

```
PT_DIR=./pt PT_AUTH=secret go run .
```

**Pushover notifications**

If `PUSHOVER_CREDENTIALS` is defined and the notify directory exists (defined by `$PT_NOTIFY_DIR` - default: `$PT_DIR/.notify`),
notifications to the given Pushover account can be activated on link-by-link basis by touching files in the notify directory.
For instance, to notify on requests to `h5.pt/test`: `touch ${PT_NOTIFY_DIR}/test` in the notify dir.



### Docker

```
docker run -it \
    -p 4600:4600 \
    -v ./pt:/pt \
    -e PT_DIR=/pt \
    -e PT_AUTH=secret \
    haflan/ptsrv
```

`ptsrv` does not deal with TLS.
Use a reverse proxy like [Caddy](https://caddyserver.com/) for production.


## API
`ptsrv` simply stores (long) URLs in files, so that `https://$HOST/<code>` redirects to the URL found in the file `$PT_DIR/<code>`.
Files can be maintained manually, but `ptsrv` also has an API for convenience:

    GET /.list?auth=<auth>[&json]       Lists all links and targets
    POST /[code]?auth=<auth>            Creates new link

**Parameters:**

* `auth`: Query - authentication code (alt: `auth: <auth>` header)
* `json`: Query - return JSON instead of plaintext (alt: `accept: application/json` header)
* `code`: Path - requested short code (default: random code)

Updating links is not possible with the API. If the requested code already exists, the server will respond with an error.
Change the files manually on server if you need to update or delete anything.

### Examples
```
read -s SECRET
$ curl -H 'auth: $SECRET' -d 'https://example.org' localhost:4600/
Jv20
$ curl -d 'https://example.org' localhost:4600/?auth=$SECRET
akG7
$ curl -d 'https://example.org' localhost:4600/short?auth=$SECRET
short
$ curl localhost:4600/.list?auth=$SECRET
Jv20
https://example.org

akG7
https://example.org

short
https://example.org
$ curl 'localhost:4600/.list?auth=$SECRET&json'
[{"code":"Jv20","target":"https://example.org"},{"code":"akG7","target":"https://example.org"},{"code":"short","target":"https://example.org"}]
```

## Motivation
Motivated by the desire to avoid long links and dead links.
By using `ptsrv` as a central database of long-lived links it's easy to spot and fix dead ones.
