[![GoDoc](https://godoc.org/github.com/bitsgofer/regen?status.svg)](https://godoc.org/github.com/bitsgofer/regen)

This is a fork of <https://github.com/nilium/regen>.

There are some significant changes from the original code, however:

- Converted from binary release to a library, exposing the `GenString` function for reuse.
- Switched from `crypto/rand` to just using `math/rand` by default.
- Added an option to control of the randomness (using a function whose behavior is similar to rand.Int63n).
