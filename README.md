# fx

A hex dumper that fills the terminal width.

## Building

Needs Go 1.27.1 or newer; dependencies download automatically.

```sh
go build ./cmd/fx
```

or install into `$(go env GOPATH)/bin`:

```sh
go install ./cmd/fx
```

## Examples

Given this file:

```sh
printf 'The quick brown fox jumps over the lazy dog.\nPack my box with five dozen liquor jugs.\n' > sample.txt
```

Standard output, 16 bytes per row (the default when output isn't a terminal,
or with `--no-fit`):

```console
$ fx --no-fit sample.txt
00000000: 5468 6520 7175 6963 6b20 6272 6f77 6e20  The quick brown 
00000010: 666f 7820 6a75 6d70 7320 6f76 6572 2074  fox jumps over t
00000020: 6865 206c 617a 7920 646f 672e 0a50 6163  he lazy dog..Pac
00000030: 6b20 6d79 2062 6f78 2077 6974 6820 6669  k my box with fi
00000040: 7665 2064 6f7a 656e 206c 6971 756f 7220  ve dozen liquor 
00000050: 6a75 6773 2e0a                           jugs..
```

32 bytes per row:

```console
$ fx -c 32 sample.txt
00000000: 5468 6520 7175 6963 6b20 6272 6f77 6e20 666f 7820 6a75 6d70 7320 6f76 6572 2074  The quick brown fox jumps over t
00000020: 6865 206c 617a 7920 646f 672e 0a50 6163 6b20 6d79 2062 6f78 2077 6974 6820 6669  he lazy dog..Pack my box with fi
00000040: 7665 2064 6f7a 656e 206c 6971 756f 7220 6a75 6773 2e0a                           ve dozen liquor jugs..
```

On a terminal, fx picks the bytes per row to fit its width.

## Options

Run `fx --help` for options.
