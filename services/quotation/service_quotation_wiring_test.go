package quotation_test

import (
	"reflect"
	"testing"

	service "github.com/malikabdulaziz/tmn-backend/services/quotation"
	"github.com/malikabdulaziz/tmn-backend/testutil"
	"github.com/malikabdulaziz/tmn-backend/testutil/mocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Every dependency the constructor accepts must actually be assigned.
//
// RepositorySalesPackage was taken as a parameter and silently dropped, so it was
// nil for the life of the process. Nothing caught it because the only code path
// that touches it is package mode, and package mode was unreachable -- no package
// had a price, so nobody could quote one. The first quotation to try it got a nil
// pointer dereference and a 500.
//
// Reflecting over the struct catches the whole class rather than this one instance:
// add a dependency, forget to wire it, and this fails.
func TestQuotationServiceWiresEveryDependency(t *testing.T) {
	db, _ := testutil.NewMockDB(t)

	svc := service.NewServiceQuotationImpl(
		db,
		&mocks.MockRepositoryQuotation{},
		&mocks.MockRepositoryBuildingPrice{},
		&mocks.MockRepositoryBuilding{},
		&mocks.MockRepositoryCustomer{},
		&mocks.MockRepositoryBrand{},
		&mocks.MockRepositoryUser{},
		&mocks.MockRepositorySalesAssignment{},
		&mocks.MockRepositorySalesPackage{},
	)

	impl, ok := svc.(*service.ServiceQuotationImpl)
	require.True(t, ok, "constructor should return the concrete implementation")

	value := reflect.ValueOf(impl).Elem()
	for i := 0; i < value.NumField(); i++ {
		field := value.Field(i)
		name := value.Type().Field(i).Name

		switch field.Kind() {
		case reflect.Interface, reflect.Ptr, reflect.Map, reflect.Slice, reflect.Func:
			assert.False(t, field.IsNil(), "%s was accepted by the constructor but never assigned", name)
		}
	}
}
