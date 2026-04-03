package mve

import (
	"context"
	"strings"

	"github.com/megaport/megaport-cli/internal/utils"
	megaport "github.com/megaport/megaportgo"
)

func filterMVEImages(images []*megaport.MVEImage, vendor, productCode string, id int, version string, releaseImage bool) []*megaport.MVEImage {
	return utils.Filter(images, func(image *megaport.MVEImage) bool {
		if vendor != "" && image.Vendor != vendor {
			return false
		}
		if productCode != "" && image.ProductCode != productCode {
			return false
		}
		if id != 0 && image.ID != id {
			return false
		}
		if version != "" && image.Version != version {
			return false
		}
		if releaseImage && !image.ReleaseImage {
			return false
		}
		return true
	})
}

func filterMVEs(mves []*megaport.MVE, locationID int, vendor, name string) []*megaport.MVE {
	return utils.Filter(mves, func(mve *megaport.MVE) bool {
		if mve == nil {
			return false
		}
		if locationID > 0 && mve.LocationID != locationID {
			return false
		}
		if vendor != "" && !strings.EqualFold(mve.Vendor, vendor) {
			return false
		}
		if name != "" && !strings.Contains(strings.ToLower(mve.Name), strings.ToLower(name)) {
			return false
		}
		return true
	})
}

var listMVEResourceTagsFunc = func(ctx context.Context, client *megaport.Client, mveID string) (map[string]string, error) {
	return client.MVEService.ListMVEResourceTags(ctx, mveID)
}

var lockMVEFunc = func(ctx context.Context, client *megaport.Client, mveUID string) (*megaport.ManageProductLockResponse, error) {
	return client.ProductService.ManageProductLock(ctx, &megaport.ManageProductLockRequest{ProductID: mveUID, ShouldLock: true})
}

var unlockMVEFunc = func(ctx context.Context, client *megaport.Client, mveUID string) (*megaport.ManageProductLockResponse, error) {
	return client.ProductService.ManageProductLock(ctx, &megaport.ManageProductLockRequest{ProductID: mveUID, ShouldLock: false})
}

var restoreMVEFunc = func(ctx context.Context, client *megaport.Client, mveUID string) (*megaport.RestoreProductResponse, error) {
	return client.ProductService.RestoreProduct(ctx, mveUID)
}
