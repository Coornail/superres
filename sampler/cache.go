package sampler

import (
	"image"
)

type ImageSamplerCache struct {
	Sampler ImageSampler

	serveCache bool
	cache      []image.Point
	i          int
}

func (isc *ImageSamplerCache) HasMore() bool {
	var hasMore bool
	if !isc.serveCache {
		hasMore = isc.Sampler.HasMore()
		if !hasMore {
			isc.serveCache = true
		}
	} else {
		hasMore = isc.i < len(isc.cache)
	}

	return hasMore
}

func (isc *ImageSamplerCache) Next() (x, y int) {
	if isc.serveCache {
		x, y := isc.cache[isc.i].X, isc.cache[isc.i].Y
		isc.i++

		return x, y
	} else {
		x, y := isc.Sampler.Next()
		isc.cache = append(isc.cache, image.Point{X: x, Y: y})
		return x, y
	}
}

func (isc *ImageSamplerCache) Reset() {
	isc.i = 0
}

func NewSamplerCache(sampler ImageSampler) *ImageSamplerCache {
	cache := make([]image.Point, 0)
	return &ImageSamplerCache{
		Sampler: sampler,

		cache: cache,
	}
}
