FROM golang:1.15

WORKDIR /tusk

RUN set -e; \
  apt-get update; \
  apt-get install -y \
    python3-pip \
    rpm; \
  wget -qO- 'https://install.goreleaser.com/github.com/golangci/golangci-lint.sh' \
    | sh -s -- -b $GOPATH/bin v1.30.0; \
  go get github.com/jstemmer/go-junit-report; \
  wget -qO- 'https://github.com/goreleaser/goreleaser/releases/download/v0.117.2/goreleaser_Linux_x86_64.tar.gz' \
    | tar xvzf - -C /usr/local/bin goreleaser; \
    chmod +x /usr/local/bin/goreleaser; \
  pip3 install --no-cache-dir \
    mkdocs==1.1;

CMD [ "bash" ]
