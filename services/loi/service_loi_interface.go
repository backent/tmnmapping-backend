package loi

import "context"

type ServiceLOIInterface interface {
	SyncFromERP(ctx context.Context) error
}
