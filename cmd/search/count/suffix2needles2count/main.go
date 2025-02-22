package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"strconv"

	sc "github.com/takanoriyanagitani/go-text2suffix/search/count"
	. "github.com/takanoriyanagitani/go-text2suffix/util"
)

func envValByKey(key string) IO[string] {
	return func(_ context.Context) (string, error) {
		val, found := os.LookupEnv(key)
		switch found {
		case true:
			return val, nil
		default:
			return "", fmt.Errorf("env var %s missing", key)
		}
	}
}

var idxFileSizeMaxInt IO[int] = Bind(
	envValByKey("ENV_IDX_FILE_SIZE_MAX"),
	Lift(strconv.Atoi),
).Or(Of(int(sc.IndexFileSizeMaxDefault)))

var ixname IO[string] = envValByKey("ENV_INDEX_FILE_NAME")

var ixname2stdin2needles2counts2stdout IO[Void] = Bind(
	All(
		idxFileSizeMaxInt.ToAny(),
		ixname.ToAny(),
	),
	func(a []any) IO[Void] {
		return func(ctx context.Context) (Void, error) {
			return Empty, sc.IndexFilenameToStdinToNeedlesToCountsToStdout(
				ctx,
				a[1].(string),
				int64(a[0].(int)),
			)
		}
	},
)

var sub IO[Void] = func(ctx context.Context) (Void, error) {
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	return ixname2stdin2needles2counts2stdout(ctx)
}

func main() {
	_, e := sub(context.Background())
	if nil != e {
		log.Printf("%v\n", e)
	}
}
