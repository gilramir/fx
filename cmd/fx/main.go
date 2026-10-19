// Command fx dumps a file as hex, like xxd, filling the terminal width.
package main

import (
	"fmt"
	"io"
	"os"

	"github.com/gilramir/argparse/v2"
	"golang.org/x/term"

	"github.com/gilramir/fx/internal/dump"
)

type options struct {
	File  string
	Cols  int
	Group int
	Seek  int64
	Len   int64
	Upper bool
	Width int
	NoFit bool
	Pager bool
}

func main() {
	opts := &options{Group: 2}
	ap := argparse.New(&argparse.Command{
		Name:        "fx",
		Description: "Dump a file as hex, like xxd, using the full width of the terminal.",
		Values:      opts,
		Epilog: "When output is not a terminal, lines have the standard width " +
			"(16 bytes) unless --width or --cols is given.",
	})
	ap.Add(&argparse.Argument{Name: "file", NumArgsGlob: "?",
		Help: "File to read; standard input if omitted or \"-\""})
	ap.Add(&argparse.Argument{Switches: []string{"-c", "--cols"}, MetaVar: "N",
		Help: "Bytes per line, instead of fitting the terminal"})
	ap.Add(&argparse.Argument{Switches: []string{"-g", "--group"}, MetaVar: "N",
		Help: "Bytes per group"})
	ap.Add(&argparse.Argument{Switches: []string{"-s", "--seek"}, MetaVar: "OFFSET",
		Help: "Start at this byte offset (0x for hex)"})
	ap.Add(&argparse.Argument{Switches: []string{"-l", "--len"}, MetaVar: "N",
		Help: "Stop after this many bytes"})
	ap.Add(&argparse.Argument{Switches: []string{"-u", "--upper"},
		Help: "Upper-case hex digits"})
	ap.Add(&argparse.Argument{Switches: []string{"-w", "--width"}, MetaVar: "COLUMNS",
		Help: "Fit lines to this many columns instead of the terminal's width"})
	ap.Add(&argparse.Argument{Switches: []string{"--no-fit"},
		Help: "Use the standard line width even on a terminal"})
	ap.Add(&argparse.Argument{Switches: []string{"-p", "--pager"},
		Help: "Show the output in a pager ($PAGER, else less, else more), with lines " +
			"fitted to the terminal. Piping into a pager yourself loses the terminal width"})
	ap.Parse()

	if err := run(opts, ap.Root.Seen); err != nil {
		fmt.Fprintln(os.Stderr, "fx:", err)
		os.Exit(1)
	}
}

func run(opts *options, seen map[string]bool) error {
	if opts.Seek < 0 || opts.Len < 0 {
		return fmt.Errorf("--seek and --len must not be negative")
	}
	in, size, err := openInput(opts.File, opts.Seek)
	if err != nil {
		return err
	}
	defer in.Close()
	var r io.Reader = in
	if seen["Len"] {
		r = io.LimitReader(r, opts.Len)
		if size < 0 || opts.Len < size-opts.Seek {
			size = opts.Seek + opts.Len
		}
	}

	// The offset column must hold the last offset, if we can know it.
	last := int64(-1)
	if size > 0 {
		last = size - 1
	}
	offW := dump.OffsetWidth(last)
	width := fitWidth(opts, seen)

	// Measure the terminal before starting the pager: once it's running,
	// our output goes into its pipe.
	var out io.Writer = os.Stdout
	if opts.Pager && term.IsTerminal(int(os.Stdout.Fd())) {
		pg, err := startPager()
		if err != nil {
			return err
		}
		out = pg.in
		return pg.finish(dumpTo(out, r, opts, seen, width, offW))
	}
	return dumpTo(out, r, opts, seen, width, offW)
}

func dumpTo(out io.Writer, r io.Reader, opts *options, seen map[string]bool, width, offW int) error {
	if opts.Group < 1 {
		return fmt.Errorf("--group must be at least 1")
	}
	cols := dump.HexStandardCols
	switch {
	case seen["Cols"]:
		cols = opts.Cols
	case width > 0:
		cols = dump.FitHexCols(width, opts.Group, offW)
	}
	return dump.Hex(out, r, dump.HexOptions{
		Cols: cols, Group: opts.Group, Upper: opts.Upper, Start: opts.Seek, OffsetWidth: offW,
	})
}

// fitWidth returns the number of columns to fit lines to, or 0 to use the
// standard line width.
func fitWidth(opts *options, seen map[string]bool) int {
	switch {
	case seen["Width"]:
		return opts.Width
	case opts.NoFit:
		return 0
	}
	fd := int(os.Stdout.Fd())
	if !term.IsTerminal(fd) {
		return 0
	}
	w, _, err := term.GetSize(fd)
	if err != nil {
		return 0
	}
	return w
}

// openInput opens the file (or stdin) positioned at seek. size is the
// file's total size, or -1 if unknown.
func openInput(name string, seek int64) (io.ReadCloser, int64, error) {
	if name == "" || name == "-" {
		if seek > 0 {
			if _, err := io.CopyN(io.Discard, os.Stdin, seek); err != nil && err != io.EOF {
				return nil, 0, err
			}
		}
		return io.NopCloser(os.Stdin), -1, nil
	}
	f, err := os.Open(name)
	if err != nil {
		return nil, 0, err
	}
	size := int64(-1)
	if st, err := f.Stat(); err == nil && st.Mode().IsRegular() {
		size = st.Size()
	}
	if seek > 0 {
		if _, err := f.Seek(seek, io.SeekStart); err != nil {
			f.Close()
			return nil, 0, err
		}
	}
	return f, size, nil
}
