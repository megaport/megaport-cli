package mve

import (
	"context"
	"strings"

	megaport "github.com/megaport/megaportgo"
)

func filterMVEImages(images []*megaport.MVEImage, vendor, productCode string, id int, version string, releaseImage bool) []*megaport.MVEImage {
	var filtered []*megaport.MVEImage
	for _, image := range images {
		if vendor != "" && image.Vendor != vendor {
			continue
		}
		if productCode != "" && image.ProductCode != productCode {
			continue
		}
		if id != 0 && image.ID != id {
			continue
		}
		if version != "" && image.Version != version {
			continue
		}
		if releaseImage && !image.ReleaseImage {
			continue
		}
		filtered = append(filtered, image)
	}
	return filtered
}

func filterMVEs(mves []*megaport.MVE, locationID int, vendor, name string) []*megaport.MVE {
	var filtered []*megaport.MVE
	if mves == nil {
		return filtered
	}
	for _, mve := range mves {
		if mve == nil {
			continue
		}
		if locationID > 0 && mve.LocationID != locationID {
			continue
		}
		mveVendor := mve.Vendor
		if vendor != "" && !strings.EqualFold(mveVendor, vendor) {
			continue
		}
		if name != "" && !strings.Contains(strings.ToLower(mve.Name), strings.ToLower(name)) {
			continue
		}
		filtered = append(filtered, mve)
	}
	return filtered
}

var listMVEResourceTagsFunc = func(ctx context.Context, client *megaport.Client, mveID string) (map[string]string, error) {
	return client.MVEService.ListMVEResourceTags(ctx, mveID)
}
