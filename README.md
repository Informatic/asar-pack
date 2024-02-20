# asar-pack

Efficient [.asar](https://github.com/electron/asar) packer tool.

This was created as a quick replacement due to very high memory usage, low
performance and limit of 2^21 *entries* (files+directories) in globbed
directories (due to use of
[Promise.all](https://github.com/electron/asar/blob/94cb8bd/lib/crawlfs.js#L22))
in official `asar` tool.

For a sample dataset of ~2.6 milion 2-4 byte files in 2 directories deep tree
on a consumer NVMe drive `asar-pack` is able to create a bundle in about a
minute. In contrast, at that point, official asar tool would need increase in
NodeJS stack size, patching for `Promise.all` argument length, and would still
fail after ~1.5h on header `JSON.stringify` string size limit.

## Limitations
Due to its simplicity this tool has several limitations:

* [ ] Missing glob patterns support
* [ ] Missing `integrity` (seems to work with latest Electron just fine due to
  backwards compatiblity)
* [ ] Empty directories are not emitted in the bundle
* [ ] Missing `executable` flag
* [ ] Missing ordering file support
* [ ] Missing `--unpack` functionality
* [ ] Missing `--exclude-hidden` functionality
* [ ] Missing listing/extraction functionality

## Building
```sh
CGO_ENABLED=0 go build .
```

## Usage
```sh
./asar-pack -source /source/path -output /artifact/target.asar
```

### Testing dataset
```sh
for a in {0..40}; do
    echo $a;
    for b in {0..255}; do
        mkdir -p $a/$b;
        for c in {0..255}; do
            echo $c > $a/$b/$c;
        done;
    done;
done
```
