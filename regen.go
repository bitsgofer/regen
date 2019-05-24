// Copyright 2016 Noel Cower. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be found in the LICENSE.txt file.

// Package regen is a tool to parse and generate random strings from regular expressions.
//
// This is a fork of github.com/nilium/regen.
//
// It provides a function to generate string from regular expressions.
//
// regen works by parsing a regular expression and walking its op tree.
// It is currently not guaranteed to produce entirely accurate results, but will at least try.
//
// Limitations from upstream:
//
//   - Currently, word boundaries are not supported (until I decide how best to randomly insert
//     a word boundary character).
//   - Using a word boundary op (\b or \B) will currently cause regen to panic.
//   - In addition, line endings are also poorly supported right now and EOT markers are treated
//     as the end of string generation.
package regen

import (
	"bytes"
	"io"
	"math/rand"
	"regexp/syntax"
)

const unboundMax = 32

type Options struct {
	randInt63n func(n int64) int64
}

func newOptions() Options {
	return Options{
		randInt63n: rand.Int63n, // uses the RNG in math/rand by default
	}
}

// Option allow us to modify the behavior of GenString.
type Option func(*Options)

// RandSrc modifies GenString, such that we use a different function to generate random number.
// This function will affect number of repetitions as well as the exact character used.
// Given the input n, randInt63n must return a number in [0, n).
func RandSrc(randInt63n func(n int64) int64) func(*Options) {
	return func(opts *Options) {
		opts.randInt63n = randInt63n
	}
}

// GenString writes a response that should, ideally, be a match for rx to w, and proceeds to do the same for its
// sub-expressions where applicable. Returns io.EOF if it encounters OpEndText. This may not be entirely correct
// behavior for OpEndText handling. Otherwise, returns nil.
func GenString(w *bytes.Buffer, rx *syntax.Regexp, Options ...Option) error {
	opts := newOptions()
	for _, applyOpt := range Options {
		applyOpt(&opts)
	}
	randInt63n := opts.randInt63n

	switch rx.Op {
	case syntax.OpNoMatch:
		return nil
	case syntax.OpEmptyMatch:
		return nil
	case syntax.OpLiteral:
		w.WriteString(string(rx.Rune))
	case syntax.OpCharClass:
		sum := 0
		for i := 0; i < len(rx.Rune); i += 2 {
			sum += 1 + int(rx.Rune[i+1]-rx.Rune[i])
		}

		for i, nth := 0, rune(randInt63n(int64(sum))); i < len(rx.Rune); i += 2 {
			min, max := rx.Rune[i], rx.Rune[i+1]
			delta := max - min
			if nth <= delta {
				w.WriteRune(min + nth)
				return nil
			}
			nth -= 1 + delta
		}
		panic("unreachable")
	case syntax.OpAnyCharNotNL:
		w.WriteRune(rune(' ' + randInt63n(95)))
	case syntax.OpAnyChar:
		i := int(randInt63n(96))
		ch := rune(' ' + i)
		if i == 95 {
			ch = '\n'
		}
		w.WriteRune(ch)
	case syntax.OpBeginLine:
		if w.Len() != 0 {
			w.WriteByte('\n')
		}
	case syntax.OpEndLine:
		if w.Len() != 0 {
			w.WriteByte('\n')
		} else {
			return io.EOF
		}
	case syntax.OpBeginText:
	case syntax.OpEndText:
		return io.EOF
	case syntax.OpWordBoundary:
		fallthrough
	case syntax.OpNoWordBoundary:
		panic("regen: word boundaries not supported yet")
	case syntax.OpStar, syntax.OpPlus:
		min := 0
		if rx.Op == syntax.OpPlus {
			min = 1
		}
		max := min + unboundMax

		for sz := min + int(randInt63n(int64(max)-int64(min)+1)); sz > 0; sz-- {
			for _, rx := range rx.Sub {
				GenString(w, rx, options...)
			}
		}
	case syntax.OpQuest:
		if randInt63n(0xFFFFFFFF) > 0x7FFFFFFF {
			for _, rx := range rx.Sub {
				if err := GenString(w, rx, options...); err != nil {
					return err
				}
			}
		}
	case syntax.OpRepeat:
		min := rx.Min
		max := rx.Max
		if max == -1 {
			max = min + unboundMax
		}
		for sz := min + int(randInt63n(int64(max)-int64(min)+1)); sz > 0; sz-- {
			for _, rx := range rx.Sub {
				if err := GenString(w, rx, options...); err != nil {
					return err
				}
			}
		}

	case syntax.OpConcat, syntax.OpCapture:
		for _, rx := range rx.Sub {
			if err := GenString(w, rx, options...); err != nil {
				return err
			}
		}
	case syntax.OpAlternate:
		nth := randInt63n(int64(len(rx.Sub)))
		return GenString(w, rx.Sub[nth], options...)
	}

	return nil
}
