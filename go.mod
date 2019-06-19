module github.com/rliebz/tusk

require (
	github.com/fatih/color v1.7.0
	github.com/google/go-cmp v0.3.0
	github.com/kr/pretty v0.1.0 // indirect
	github.com/mattn/go-colorable v0.1.2 // indirect
	github.com/pkg/errors v0.8.1
	github.com/urfave/cli v1.20.0
	golang.org/x/sys v0.0.0-20190618155005-516e3c20635f // indirect
	gopkg.in/check.v1 v1.0.0-20180628173108-788fd7840127 // indirect
	gopkg.in/yaml.v2 v2.2.2
	gotest.tools v2.2.0+incompatible
)

replace github.com/urfave/cli => github.com/rliebz/cli v0.0.1-tusk
