// Highly experimental plugin for caching rendered pages for guests
package main

import (
	"sync/atomic"

	"github.com/Azareal/Gosora/common"
)

var hyperPageCache *HyperPageCache

func init() {
	common.Plugins.Add(&common.Plugin{UName: "hyperdrive", Name: "Hyperdrive", Author: "Azareal", Init: initHyperdrive, Deactivate: deactivateHyperdrive})
}

func initHyperdrive(plugin *common.Plugin) error {
	hyperPageCache = newHyperPageCache()
	plugin.AddHook("somewhere", deactivateHyperdrive)
	return nil
}

func deactivateHyperdrive(plugin *common.Plugin) {
	hyperPageCache = nil
}

type HyperPageCache struct {
	topicList atomic.Value
}

func newHyperPageCache() *HyperPageCache {
	pageCache := new(HyperPageCache)
	pageCache.topicList.Store([]byte(""))
	return pageCache
}
