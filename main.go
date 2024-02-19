package main

import (
	"encoding/binary"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"time"

	"k8s.io/klog/v2"
)

type FileEntry struct {
	Offset string `json:"offset"`
	Size   int    `json:"size"`
	// Executable bool   `json:"executable"`
	// TODO: integrity
}

type DirectoryEntry struct {
	Files map[string]interface{} `json:"files"`
}

func init() {
	klog.InitFlags(nil)
	flag.Parse()
}

var outputFlag = flag.String("output", "output.asar", "output asar filename")
var sourceFlag = flag.String("source", "source", "source path to pack")

func main() {
	klog.Infof("preparing index...")
	offset := 0

	index := DirectoryEntry{
		Files: make(map[string]interface{}),
	}

	walkRoot, err := filepath.EvalSymlinks(*sourceFlag)
	if err != nil {
		klog.Fatalf("unable to eval source: %v", err)
	}

	var filesToDump []string
	ts := time.Now()

	err = filepath.Walk(walkRoot, func(fullPath string, info fs.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if fullPath == walkRoot {
			return nil
		}

		if info.IsDir() {

		} else if info.Mode().IsRegular() {
			p := strings.TrimPrefix(fullPath, walkRoot+"/")

			parts := strings.Split(p, "/")
			ptr := &index
			for _, part := range parts[:len(parts)-1] {
				if ptr.Files[part] == nil {
					ptr.Files[part] = DirectoryEntry{
						Files: make(map[string]interface{}),
					}
				}

				v, ok := ptr.Files[part].(DirectoryEntry)
				if ok {
					ptr = &v
				} else {
					return errors.New("unknonwn file type")
				}
			}

			ptr.Files[filepath.Base(p)] = FileEntry{
				Offset: fmt.Sprintf("%d", offset),
				Size:   int(info.Size()),
			}

			filesToDump = append(filesToDump, fullPath)
			offset += int(info.Size())
		}

		return nil
	})

	if err != nil {
		klog.Fatalf("error occured during walk: %v", err)
	}

	data, err := json.Marshal(index)
	if err != nil {
		klog.Fatalf("unable to generate index: %v", err)
	}

	fd, err := os.Create(*outputFlag)
	if err != nil {
		panic(err)
	}
	defer fd.Close()

	paddingLength := (4 - len(data)%4) % 4

	klog.Infof("index built in %s (%d bytes, %d padding)", time.Since(ts), len(data), paddingLength)

	preheader := make([]byte, 16+paddingLength)
	binary.LittleEndian.PutUint32(preheader[0:4], 4)
	binary.LittleEndian.PutUint32(preheader[4:8], uint32(len(data)+paddingLength+8))
	binary.LittleEndian.PutUint32(preheader[8:12], uint32(len(data)+paddingLength+4))
	binary.LittleEndian.PutUint32(preheader[12:16], uint32(len(data)))

	fd.Write(preheader)
	fd.Write(data)

	for idx, file := range filesToDump {
		infd, err := os.Open(file)
		if err != nil {
			panic(err)
		}

		io.Copy(fd, infd)
		infd.Close()
		if idx%10000 == 0 {
			klog.Infof("%d/%d files dumped...", idx+1, len(filesToDump))
		}
	}

	klog.Infof("finished in %s", time.Since(ts))
}
