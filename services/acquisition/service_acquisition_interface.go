package acquisition

import "context"

type ServiceAcquisitionInterface interface {
	SyncFromERP(ctx context.Context) error
}
