package util

import (
	ts "github.com/takanoriyanagitani/go-text2suffix"
)

func ComposeErr[T, U, V any](
	f func(T) (U, error),
	g func(U) (V, error),
) func(T) (V, error) {
	return ts.ComposeErr(f, g)
}
