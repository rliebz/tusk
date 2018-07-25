FROM golang:1.10

WORKDIR /go/src/github.com/rliebz/tusk

# Use install script v2.0.6 for binary v2.0.5
RUN curl -sL https://raw.githubusercontent.com/alecthomas/gometalinter/v2.0.6/scripts/install.sh | bash -s -- -b /usr/local/bin v2.0.5
RUN go get github.com/jstemmer/go-junit-report
RUN apt-get update && \
      apt-get -y install rpm && \
      curl -sL https://git.io/goreleaser > /tmp/goreleaser && \
      chmod +x /tmp/goreleaser && \
      mv /tmp/goreleaser /usr/local/bin
RUN apt-get update && \
      apt-get install -y python-pip && \
      pip install --no-cache-dir \
      mkdocs==0.17.3 \
      mkdocs-cinder==0.11.0 \
      Pygments==2.2.0

CMD [ "bash" ]
