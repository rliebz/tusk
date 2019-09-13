FROM golang:1.13

WORKDIR /tusk

RUN set -e; \
  apt-get update; \
  apt-get install -y \
    python-pip \
    rpm; \
  wget -qO- 'https://install.goreleaser.com/github.com/golangci/golangci-lint.sh' \
    | sh -s -- -b $GOPATH/bin v1.18.0; \
  go get github.com/jstemmer/go-junit-report; \
  wget -qO- 'https://github.com/goreleaser/goreleaser/releases/download/v0.117.2/goreleaser_Linux_x86_64.tar.gz' \
    | tar xvzf - -C /usr/local/bin goreleaser; \
    chmod +x /usr/local/bin/goreleaser; \
  pip install --no-cache-dir \
    mkdocs==1.0.4 \
    mkdocs-rtd-dropdown==1.0.2 \
    Pygments==2.3.0;

CMD [ "bash" ]
