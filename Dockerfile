FROM golang:1.19-alpine as go-builder
RUN apk add --no-cache git
WORKDIR /go/nginx-smtp-auth
COPY . .
RUN go generate -x -v
RUN go build -o /bin/nginx-smtp-auth 
RUN chmod +x /bin/nginx-smtp-auth

MAINTAINER Zikani Nyirenda Mwase <zikani.nmwase@ymail.com>
FROM golang:1.19-alpine
COPY --from=go-builder /bin/nginx-smtp-auth /nginx-smtp-auth
CMD ["/nginx-smtp-auth"]
