package main

import (
	"fmt"
	"path/filepath"
	"sort"
	"strings"

	"github.com/triole/logseal"
)

type tJoinerEntry struct {
	Path        string                 `json:"path"`
	SplitPath   []string               `json:"split_path,omitempty"`
	Size        uint64                 `json:"size,omitempty"`
	LastMod     int64                  `json:"lastmod,omitempty"`
	Created     int64                  `json:"created,omitempty"`
	FrontMatter map[string]interface{} `json:"front_matter,omitempty"`
	Content     map[string]interface{} `json:"content,omitempty"`
}

type tJoinerIndex []tJoinerEntry

func (arr tJoinerIndex) Len() int {
	return len(arr)
}

func getDepth(pth string) int {
	return len(strings.Split(pth, string(filepath.Separator))) - 1
}

func (arr tJoinerIndex) Less(i, j int) bool {
	si1 := fmt.Sprintf("%05d_%s", getDepth(arr[i].Path), arr[i].Path)
	si2 := fmt.Sprintf("%05d_%s", getDepth(arr[j].Path), arr[j].Path)
	return si1 < si2
}

func (arr tJoinerIndex) Swap(i, j int) {
	arr[i], arr[j] = arr[j], arr[i]
}

func makeJoinerIndex(params tIDXParams) (joinerIndex tJoinerIndex) {
	lg.Debug(
		"make joiner index",
		logseal.F{"index_params": fmt.Sprintf("%+v", params)},
	)
	dataFiles := find(params.Endpoint.Folder, params.Endpoint.RxFilter)
	ln := len(dataFiles)

	if ln < 1 {
		lg.Warn("no data files found", logseal.F{"path": params.Endpoint.Folder})
	} else {
		chin := make(chan string, params.Threads)
		chout := make(chan tJoinerEntry, params.Threads)

		lg.Debug(
			"start to process files",
			logseal.F{"no": ln, "threads": params.Threads},
		)

		for _, fil := range dataFiles {
			go readDataFile(fil, params.Endpoint, chin, chout)
		}

		c := 0
		for li := range chout {
			joinerIndex = append(joinerIndex, li)
			c++
			if c >= ln {
				close(chin)
				close(chout)
				break
			}
		}
		joinerIndex = sortJoinerIndex(joinerIndex, params)
		lg.Debug(
			"index created",
			logseal.F{"path": params.Endpoint.Folder},
		)
	}
	return
}

func sortJoinerIndex(arr tJoinerIndex, params tIDXParams) tJoinerIndex {
	switch params.SortBy {
	case "created":
		sort.Slice(arr, func(i, j int) bool {
			if !params.Ascending {
				return arr[i].Created > arr[j].Created
			}
			return arr[i].Created < arr[j].Created
		})
	case "lastmod":
		sort.Slice(arr, func(i, j int) bool {
			if !params.Ascending {
				return arr[i].LastMod > arr[j].LastMod
			}
			return arr[i].LastMod < arr[j].LastMod
		})
	case "size":
		sort.Slice(arr, func(i, j int) bool {
			if !params.Ascending {
				return arr[i].Size > arr[j].Size
			}
			return arr[i].Size < arr[j].Size
		})
	default:
		if params.Ascending {
			sort.Sort(tJoinerIndex(arr))
		} else {
			sort.Sort(sort.Reverse(tJoinerIndex(arr)))
		}
	}
	return arr
}