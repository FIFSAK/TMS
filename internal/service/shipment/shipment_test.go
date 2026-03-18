package shipment

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	domain "github.com/FIFSAK/TMS/internal/domain/shipment"
	"github.com/FIFSAK/TMS/pkg/store"
)

type MockRepository struct {
	mock.Mock
}

func (m *MockRepository) Create(ctx context.Context, s domain.Shipment) (string, error) {
	args := m.Called(ctx, s)
	return args.String(0), args.Error(1)
}

func (m *MockRepository) GetByID(ctx context.Context, id string) (domain.Shipment, error) {
	args := m.Called(ctx, id)
	return args.Get(0).(domain.Shipment), args.Error(1)
}

func (m *MockRepository) List(ctx context.Context) ([]domain.Shipment, error) {
	args := m.Called(ctx)
	return args.Get(0).([]domain.Shipment), args.Error(1)
}

func (m *MockRepository) UpdateStatus(ctx context.Context, id string, status domain.Status) error {
	args := m.Called(ctx, id, status)
	return args.Error(0)
}

func (m *MockRepository) AddEvent(ctx context.Context, event domain.Event) (string, error) {
	args := m.Called(ctx, event)
	return args.String(0), args.Error(1)
}

func (m *MockRepository) ListEvents(ctx context.Context, shipmentID string) ([]domain.Event, error) {
	args := m.Called(ctx, shipmentID)
	return args.Get(0).([]domain.Event), args.Error(1)
}

func TestCreateShipment(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		repo := new(MockRepository)
		svc := NewShipmentService(repo)
		ctx := context.Background()

		repo.On("Create", ctx, mock.AnythingOfType("shipment.Shipment")).Return("id-1", nil)
		repo.On("AddEvent", ctx, mock.AnythingOfType("shipment.Event")).Return("event-1", nil)

		result, err := svc.CreateShipment(ctx, CreateRequest{
			ReferenceNumber: "REF-001",
			Origin:          "NYC",
			Destination:     "LA",
			DriverName:      "John",
			UnitNumber:      "UNIT-1",
			ShipmentAmount:  1000,
			DriverRevenue:   500,
		})

		assert.NoError(t, err)
		assert.Equal(t, "id-1", result.ID)
		assert.Equal(t, domain.StatusPending, result.Status)
		repo.AssertExpectations(t)
	})

	t.Run("validation error", func(t *testing.T) {
		repo := new(MockRepository)
		svc := NewShipmentService(repo)

		_, err := svc.CreateShipment(context.Background(), CreateRequest{
			ReferenceNumber: "",
			Origin:          "NYC",
			Destination:     "LA",
		})

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "validation")
	})
}

func TestGetShipment(t *testing.T) {
	t.Run("found", func(t *testing.T) {
		repo := new(MockRepository)
		svc := NewShipmentService(repo)
		ctx := context.Background()

		expected := domain.Shipment{ID: "id-1", ReferenceNumber: "REF-001"}
		repo.On("GetByID", ctx, "id-1").Return(expected, nil)

		result, err := svc.GetShipment(ctx, "id-1")
		assert.NoError(t, err)
		assert.Equal(t, expected, result)
	})

	t.Run("not found", func(t *testing.T) {
		repo := new(MockRepository)
		svc := NewShipmentService(repo)
		ctx := context.Background()

		repo.On("GetByID", ctx, "missing").Return(domain.Shipment{}, store.ErrorNotFound)

		_, err := svc.GetShipment(ctx, "missing")
		assert.ErrorIs(t, err, store.ErrorNotFound)
	})
}

func TestListShipments(t *testing.T) {
	repo := new(MockRepository)
	svc := NewShipmentService(repo)
	ctx := context.Background()

	expected := []domain.Shipment{
		{ID: "1", ReferenceNumber: "REF-001"},
		{ID: "2", ReferenceNumber: "REF-002"},
	}
	repo.On("List", ctx).Return(expected, nil)

	result, err := svc.ListShipments(ctx)
	assert.NoError(t, err)
	assert.Len(t, result, 2)
}

func TestAddEvent(t *testing.T) {
	t.Run("valid transition", func(t *testing.T) {
		repo := new(MockRepository)
		svc := NewShipmentService(repo)
		ctx := context.Background()

		repo.On("GetByID", ctx, "id-1").Return(domain.Shipment{
			ID:     "id-1",
			Status: domain.StatusPending,
		}, nil)
		repo.On("AddEvent", ctx, mock.AnythingOfType("shipment.Event")).Return("event-1", nil)
		repo.On("UpdateStatus", ctx, "id-1", domain.StatusPickedUp).Return(nil)

		event, err := svc.AddEvent(ctx, "id-1", domain.StatusPickedUp, "driver picked up")
		assert.NoError(t, err)
		assert.Equal(t, "event-1", event.ID)
		assert.Equal(t, domain.StatusPickedUp, event.Status)
		repo.AssertExpectations(t)
	})

	t.Run("invalid transition", func(t *testing.T) {
		repo := new(MockRepository)
		svc := NewShipmentService(repo)
		ctx := context.Background()

		repo.On("GetByID", ctx, "id-1").Return(domain.Shipment{
			ID:     "id-1",
			Status: domain.StatusPending,
		}, nil)

		_, err := svc.AddEvent(ctx, "id-1", domain.StatusDelivered, "skip ahead")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "invalid transition")
	})

	t.Run("shipment not found", func(t *testing.T) {
		repo := new(MockRepository)
		svc := NewShipmentService(repo)
		ctx := context.Background()

		repo.On("GetByID", ctx, "missing").Return(domain.Shipment{}, store.ErrorNotFound)

		_, err := svc.AddEvent(ctx, "missing", domain.StatusPickedUp, "")
		assert.Error(t, err)
	})
}

func TestGetEvents(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		repo := new(MockRepository)
		svc := NewShipmentService(repo)
		ctx := context.Background()

		repo.On("GetByID", ctx, "id-1").Return(domain.Shipment{ID: "id-1"}, nil)
		repo.On("ListEvents", ctx, "id-1").Return([]domain.Event{
			{ID: "e1", Status: domain.StatusPending},
			{ID: "e2", Status: domain.StatusPickedUp},
		}, nil)

		events, err := svc.GetEvents(ctx, "id-1")
		assert.NoError(t, err)
		assert.Len(t, events, 2)
	})

	t.Run("shipment not found", func(t *testing.T) {
		repo := new(MockRepository)
		svc := NewShipmentService(repo)
		ctx := context.Background()

		repo.On("GetByID", ctx, "missing").Return(domain.Shipment{}, store.ErrorNotFound)

		_, err := svc.GetEvents(ctx, "missing")
		assert.Error(t, err)
	})
}
