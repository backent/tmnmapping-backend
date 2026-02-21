package buildingproposal

import "context"

type ServiceBuildingProposalInterface interface {
	SyncFromERP(ctx context.Context) error
}
