package text2suffix

import (
	"bufio"
	"bytes"
	"context"
	"index/suffixarray"
	"io"
	"iter"
	"log"
	"os"
	"path/filepath"
	"strings"
)

func DataToIndexToWriter(
	data []byte,
	wtr io.Writer,
) error {
	var ix *suffixarray.Index = suffixarray.New(data)
	return ix.Write(wtr)
}

type FileSync func(*os.File) error

type SuffixArrayConfigFs struct {
	InputDirname  string
	OutputDirname string
	MaxFileSize   int64
	OutputExt     string
	FileSync
}

func (c SuffixArrayConfigFs) CreateSuffixArrayIndex(
	ctx context.Context,
	basenames iter.Seq[string],
) error {
	irt, ie := os.OpenRoot(c.InputDirname)
	if nil != ie {
		return ie
	}
	defer irt.Close()

	ort, oe := os.OpenRoot(c.OutputDirname)
	if nil != oe {
		return oe
	}
	defer ort.Close()

	var ibuf bytes.Buffer

	for basename := range basenames {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		var ext string = filepath.Ext(basename)
		var noext string = strings.TrimSuffix(basename, ext)
		var oname string = noext + c.OutputExt

		ifile, ie := irt.Open(basename)
		ofile, oe := ort.Create(oname)

		err := func() error {
			if nil == ie {
				defer ifile.Close()
			}

			if nil == oe {
				defer ofile.Close()
			}

			stat, e := ifile.Stat()
			if nil != e {
				return e
			}

			var isize int64 = stat.Size()
			if c.MaxFileSize < isize {
				log.Printf("skipping too large file(%s)\n", basename)
				log.Printf("set ENV_MAX_FILE_SIZE larger to process it\n")
				return nil
			}

			limited := &io.LimitedReader{
				R: ifile,
				N: c.MaxFileSize,
			}

			ibuf.Reset()
			_, e = io.Copy(&ibuf, limited)
			if nil != e {
				return e
			}

			e = DataToIndexToWriter(ibuf.Bytes(), ofile)
			if nil != e {
				return e
			}

			return c.FileSync(ofile)
		}()

		if nil != err {
			return err
		}
	}

	return nil
}

func StdinToBaseNames() iter.Seq[string] {
	return func(yield func(string) bool) {
		var s *bufio.Scanner = bufio.NewScanner(os.Stdin)
		for s.Scan() {
			var basename string = s.Text()
			if !yield(basename) {
				return
			}
		}

		e := s.Err()
		if nil != e {
			log.Printf("error while getting basenames: %v\n", e)
		}
	}
}

func (c SuffixArrayConfigFs) StdinToBasenamesToIndexFiles(
	ctx context.Context,
) error {
	return c.CreateSuffixArrayIndex(ctx, StdinToBaseNames())
}

func FileSyncNop(_ *os.File) error { return nil }
func FileSyncAll(f *os.File) error { return f.Sync() }

const OutputExtDefault string = ".dat"

const MaxFileSizeDefault int64 = 16777216

var SuffixArrayConfigFsDefault SuffixArrayConfigFs = SuffixArrayConfigFs{
	InputDirname:  "",
	OutputDirname: "",
	MaxFileSize:   MaxFileSizeDefault,
	OutputExt:     OutputExtDefault,
	FileSync:      FileSyncAll,
}

func (c SuffixArrayConfigFs) WithInputDirname(d string) SuffixArrayConfigFs {
	c.InputDirname = d
	return c
}

func (c SuffixArrayConfigFs) WithOutputDirname(d string) SuffixArrayConfigFs {
	c.OutputDirname = d
	return c
}

func (c SuffixArrayConfigFs) WithMaxFileSize(i int64) SuffixArrayConfigFs {
	c.MaxFileSize = i
	return c
}

func (c SuffixArrayConfigFs) WithOutputExt(s string) SuffixArrayConfigFs {
	c.OutputExt = s
	return c
}

func (c SuffixArrayConfigFs) WithFileSync(fsync FileSync) SuffixArrayConfigFs {
	c.FileSync = fsync
	return c
}
