package count

import (
	"bufio"
	"bytes"
	"context"
	"errors"
	"fmt"
	"index/suffixarray"
	"io"
	"io/fs"
	"iter"
	"log"
	"os"
)

var ErrTooBigIndexFile error = errors.New("too big index file")

type Idx struct {
	*suffixarray.Index
}

func (i Idx) CountAll(needle []byte) int {
	return len(i.Index.Lookup(needle, -1))
}

func (i Idx) NeedlesToCountToWriter(
	ctx context.Context,
	needles iter.Seq[string],
	wtr io.Writer,
) error {
	var buf bytes.Buffer

	for needle := range needles {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		buf.Reset()
		_, _ = buf.WriteString(needle) // error is nil or panic

		var nb int = i.CountAll(buf.Bytes())
		fmt.Fprintf(wtr, "%s: %v\n", needle, nb)
	}

	return nil
}

func IndexFileToIndexToNeedlesToCountsToWriter(
	ctx context.Context,
	ixfile fs.File,
	ixfileMax int64,
	needles iter.Seq[string],
	wtr io.Writer,
) error {
	defer ixfile.Close()

	stat, e := ixfile.Stat()
	if nil != e {
		return e
	}

	var ixsz int64 = stat.Size()
	if ixfileMax < ixsz {
		log.Printf("the index file is too big.\n")
		log.Printf("set ENV_IDX_FILE_SIZE_MAX larger to accept it.\n")
		return ErrTooBigIndexFile
	}

	rdr := &io.LimitedReader{
		R: ixfile,
		N: ixfileMax,
	}

	var ix suffixarray.Index
	e = ix.Read(rdr)
	if nil != e {
		return e
	}

	return Idx{&ix}.NeedlesToCountToWriter(
		ctx,
		needles,
		wtr,
	)
}

const IndexFileSizeMaxDefault int64 = 16777216

func IndexFilenameToStdinToNeedlesToCountsToStdout(
	ctx context.Context,
	ixname string,
	ixfileMax int64,
) error {
	var bw *bufio.Writer = bufio.NewWriter(os.Stdout)
	defer bw.Flush()

	f, e := os.Open(ixname)
	if nil != e {
		return e
	}

	var needles iter.Seq[string] = func(
		yield func(string) bool,
	) {
		var s *bufio.Scanner = bufio.NewScanner(os.Stdin)
		for s.Scan() {
			if !yield(s.Text()) {
				return
			}
		}

		e := s.Err()
		if nil != e {
			log.Printf("error while getting needle lines: %v\n", e)
		}
	}

	return IndexFileToIndexToNeedlesToCountsToWriter(
		ctx,
		f,
		ixfileMax,
		needles,
		bw,
	)
}
