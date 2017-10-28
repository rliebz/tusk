package appcli

import (
	"unicode"

	"github.com/urfave/cli"
)

// flagsByName sorts flags alphabetically.
type flagsByName []cli.Flag

func (f flagsByName) Len() int {
	return len(f)
}

func (f flagsByName) Swap(i, j int) {
	f[i], f[j] = f[j], f[i]
}

func (f flagsByName) Less(i, j int) bool {
	return lexicographicLess(f[i].GetName(), f[j].GetName())
}

// commandsByName sorts commands alphabetically.
type commandsByName []cli.Command

func (c commandsByName) Len() int {
	return len(c)
}

func (c commandsByName) Swap(i, j int) {
	c[i], c[j] = c[j], c[i]
}

func (c commandsByName) Less(i, j int) bool {
	return lexicographicLess(c[i].Name, c[j].Name)
}

// lexicographicLess compares strings alphabetically considering case.
func lexicographicLess(i, j string) bool {
	iRunes := []rune(i)
	jRunes := []rune(j)

	lenShared := len(iRunes)
	if lenShared > len(jRunes) {
		lenShared = len(jRunes)
	}

	for index := 0; index < lenShared; index++ {
		ir := iRunes[index]
		jr := jRunes[index]

		if lir, ljr := unicode.ToLower(ir), unicode.ToLower(jr); lir != ljr {
			return lir < ljr
		}

		if ir != jr {
			return ir < jr
		}
	}

	return i < j
}
