package webseed

import (
	"fmt"
	"net/http"
	"net/url"
	"path"
	"strings"

	"github.com/anacrolix/torrent/metainfo"
)

type PathEscaper func(pathComps []string) string

// Escapes path name components suitable for appending to a webseed URL. This works for converting
// S3 object keys to URLs too.
//
// Contrary to the name, this actually does a QueryEscape, rather than a PathEscape. This works
// better with most S3 providers.
func EscapePath(pathComps []string) string {
	return defaultPathEscaper(pathComps)
}

func defaultPathEscaper(pathComps []string) string {
	var ret []string
	for _, comp := range pathComps {
		ret = append(ret, url.QueryEscape(comp))
	}
	return path.Join(ret...)
}

func trailingPath(
	infoName string,
	fileComps []string,
	pathEscaper PathEscaper,
) string {
	if pathEscaper == nil {
		pathEscaper = defaultPathEscaper
	}
	return pathEscaper(append([]string{infoName}, fileComps...))
}

// Creates a request per BEP 19.
func newRequest(
	url_ string, fileIndex int,
	info *metainfo.Info,
	offset, length int64,
	pathEscaper PathEscaper,
) (*http.Request, error) {
	fileInfo := info.UpvertedFiles()[fileIndex]
	if strings.HasSuffix(url_, "/") {
		// BEP specifies that we append the file path. We need to escape each component of the path
		// for things like spaces and '#'.
		url_ += trailingPath(info.Name, fileInfo.Path, pathEscaper)
	}
	req, err := http.NewRequest(http.MethodGet, url_, nil)
	if err != nil {
		return nil, err
	}
	if offset != 0 || length != fileInfo.Length {
		req.Header.Set("Range", fmt.Sprintf("bytes=%d-%d", offset, offset+length-1))
	}
	return req, nil
}
