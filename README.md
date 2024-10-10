# ptsrv

*point to - service* - minimalist URL shortener.

## Run with Docker

```
docker run -it \
    -p 4600:4600 \
    -v ./pt:/pt \
    -e PT_DIR=/pt \
    -e PT_AUTH=secret \
    haflan/ptsrv
```

## Usage

POST your link to the path you want, or POST to the server root to autogenerate an ID / path. If successful, the server responds with the ID. The auth key can be set in a header or query param.

```
$ curl -H 'auth: secret' -d 'https://example.org' localhost:4600/
Jv20
$ curl -d 'https://example.org' localhost:4600/?auth=secret
akG7
$ curl -d 'https://example.org' localhost:4600/short?auth=secret
short
```
