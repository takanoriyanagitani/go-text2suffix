package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"strconv"

	ts "github.com/takanoriyanagitani/go-text2suffix"
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

var inputDir IO[string] = envValByKey("ENV_INPUT_DIR_NAME")

var outputDir IO[string] = envValByKey("ENV_OUTPUT_DIR_NAME")

var maxFileSizeInt IO[int] = Bind(
	envValByKey("ENV_MAX_FILE_SIZE"),
	Lift(strconv.Atoi),
).Or(Of(int(ts.MaxFileSizeDefault)))

var disableFsync IO[bool] = Bind(
	envValByKey("ENV_DISABLE_FSYNC"),
	Lift(strconv.ParseBool),
).Or(Of(false))

var fsync IO[ts.FileSync] = Bind(
	disableFsync,
	Lift(func(disable bool) (ts.FileSync, error) {
		switch disable {
		case true:
			return ts.FileSyncNop, nil
		default:
			return ts.FileSyncAll, nil
		}
	}),
)

var config IO[ts.SuffixArrayConfigFs] = Bind(
	All(
		inputDir.ToAny(),
		outputDir.ToAny(),
		maxFileSizeInt.ToAny(),
		fsync.ToAny(),
	),
	Lift(func(a []any) (ts.SuffixArrayConfigFs, error) {
		return ts.SuffixArrayConfigFsDefault.
			WithInputDirname(a[0].(string)).
			WithOutputDirname(a[1].(string)).
			WithMaxFileSize(int64(a[2].(int))).
			WithFileSync(a[3].(ts.FileSync)), nil
	}),
)

func cfg2stdin2names2ixfiles(cfg ts.SuffixArrayConfigFs) IO[Void] {
	return func(ctx context.Context) (Void, error) {
		return Empty, cfg.StdinToBasenamesToIndexFiles(ctx)
	}
}

var env2cfg2stdin2names2ixfiles IO[Void] = Bind(
	config,
	cfg2stdin2names2ixfiles,
)

var sub IO[Void] = func(ctx context.Context) (Void, error) {
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	return env2cfg2stdin2names2ixfiles(ctx)
}

func main() {
	_, e := sub(context.Background())
	if nil != e {
		log.Printf("%v\n", e)
	}
}
