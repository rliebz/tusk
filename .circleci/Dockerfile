FROM golang:1.10

WORKDIR /go/src/github.com/rliebz/tusk

RUN go get github.com/alecthomas/gometalinter && gometalinter --install
RUN go get github.com/jstemmer/go-junit-report
RUN apt-get update && \
      apt-get -y install rpm && \
      curl -sL https://git.io/goreleaser > /tmp/goreleaser && \
      chmod +x /tmp/goreleaser && \
      mv /tmp/goreleaser /usr/local/bin
RUN apt-get update && \
      apt-get install -y python-pip python-dev build-essential && \
      pip install --no-cache-dir mkdocs==0.17.3 mkdocs-cinder==0.11.0

CMD [ "bash" ]
