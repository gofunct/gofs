package gofs

import (
	"bytes"
	"github.com/gofunct/gofs/assetfs"
	"github.com/gofunct/gofs/print"
	"io/ioutil"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

// Verbose indicates whether to log verbosely
var Verbose = false

// Write writes all assets to the file system.
func Write() func(assets []*assetfs.Asset) error {
	return func(assets []*assetfs.Asset) error {
		// cache directories to avoid Mkdirall
		madeDirs := map[string]bool{}
		for _, asset := range assets {
			writePath := asset.WritePath
			dir := filepath.Dir(writePath)
			if !madeDirs[dir] {
				err := os.MkdirAll(dir, 0755)
				if err != nil {
					return err
				}
				madeDirs[dir] = true
			}
			f, err := os.Create(writePath)
			if err != nil {
				return err
			}
			defer f.Close()
			asset.Buffer.WriteTo(f)
		}
		return nil
	}
}

// AddHeader prepends header to each asset's buffer unless it is already
// prefixed with the header.
func AddHeader(header string) func(*assetfs.Asset) error {
	return func(asset *assetfs.Asset) error {
		if asset.IsText() {
			if bytes.HasPrefix(asset.Bytes(), []byte(header)) {
				return nil
			}
			buffer := bytes.NewBufferString(header)
			buffer.Write(asset.Bytes())
			asset.Buffer = *buffer
		}
		return nil
	}
}

// Trace traces an asset, printing key properties of asset to the console.
func Trace() func(*assetfs.Asset) error {
	return func(asset *assetfs.Asset) error {
		print.Debug("filter", asset.Dump())
		return nil
	}
}

// ReplacePattern replaces the leading part of a path in all assets.
//
//      ReplacePath("views/", "dist/views")
//
// This should be used before the Write() filter.
func ReplacePattern(pattern, repl string) func(*assetfs.Asset) error {
	re := regexp.MustCompile(pattern)
	return func(asset *assetfs.Asset) error {
		if asset.IsText() {
			s := asset.String()
			if s != "" {
				asset.RewriteString(re.ReplaceAllString(s, repl))
			}
		}
		return nil
	}
}

// Cat concatenates all assets with a join string. Cat clears all assets
// from the pipeline replacing it with a single asset of the concatenated value.
func Cat(join string, dest string) func(*assetfs.Pipeline) error {
	return func(pipeline *assetfs.Pipeline) error {
		var buffer bytes.Buffer
		for i, asset := range pipeline.Assets {
			if i > 0 {
				buffer.WriteString(join)
			}
			buffer.Write(asset.Bytes())
		}

		// removes existing assets
		pipeline.Truncate()

		// add new asset for the concatenated buffer
		asset := &assetfs.Asset{WritePath: dest}
		asset.Write(buffer.Bytes())
		pipeline.AddAsset(asset)
		return nil
	}
}

// Load loads all the files from glob patterns and creates the initial
// asset array for a pipeline. This loads the entire contents of the file, binary
// or text, into a buffer. Consider creating your own loader if dealing
// with large files.
func Load(patterns ...string) func(*assetfs.Pipeline) error {
	return func(pipeline *assetfs.Pipeline) error {
		fileAssets, _, err := fs.Glob(patterns)
		if err != nil {
			return err
		}

		for _, info := range fileAssets {
			if !info.IsDir() {
				data, err := ioutil.ReadFile(info.Path)
				if err != nil {
					return err
				}
				asset := &assetfs.Asset{Info: info}
				asset.Write(data)
				asset.WritePath = info.Path
				pipeline.AddAsset(asset)
			}
		}
		return nil
	}
}

// ReplacePath replaces the leading part of a path in all assets.
//
//      ReplacePath("src/", "dist/")
//
// This should be used before the Write() filter.
func ReplacePath(from string, to string) func(*assetfs.Asset) error {
	return func(asset *assetfs.Asset) error {
		oldPath := asset.WritePath
		if !strings.HasPrefix(oldPath, from) {
			return nil
		}
		asset.WritePath = to + oldPath[len(from):]
		if Verbose {
			print.Debug("pipeline", "ReplacePath %s => %s\n", oldPath, asset.WritePath)
		}
		return nil
	}
}

// Str passes asset.Buffer string through any `str` filter for processing.
// asset.Buffer is assigned then asigned the result value from filter.
func Str(handler func(string) string) func(*assetfs.Asset) error {
	return func(asst *assetfs.Asset) error {
		if asst.IsText() {
			s := asst.String()
			asst.RewriteString(handler(s))
		}
		return nil
	}
}
