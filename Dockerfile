
# GitHub:       https://github.com/gohugoio
# Twitter:      https://twitter.com/gohugoio
# Website:      https://gohugo.io/

FROM golang:1.11-stretch AS build

WORKDIR /go/src/github.com/sigmonsays/webchat
RUN apt-get install \
    git gcc g++ binutils
COPY . /go/src/github.com/sigmonsays/webchat/
RUN go get -d .
ENV GOPATH=/go
RUN go install github.com/sigmonsays/webchat/...

# ---

FROM alpine:3.9
RUN apk add --no-cache ca-certificates
COPY --from=build /go/bin/chat /webchat
EXPOSE  8080
ENTRYPOINT [ "/webchat" ]
CMD [ "-h" ]
