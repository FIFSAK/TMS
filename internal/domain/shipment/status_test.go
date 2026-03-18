package shipment

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestStatus_IsValid(t *testing.T) {
	tests := []struct {
		status Status
		valid  bool
	}{
		{StatusPending, true},
		{StatusPickedUp, true},
		{StatusInTransit, true},
		{StatusDelivered, true},
		{StatusCancelled, true},
		{Status("unknown"), false},
		{Status(""), false},
	}

	for _, tt := range tests {
		t.Run(string(tt.status), func(t *testing.T) {
			assert.Equal(t, tt.valid, tt.status.IsValid())
		})
	}
}

func TestStatus_CanTransitionTo(t *testing.T) {
	tests := []struct {
		name     string
		from     Status
		to       Status
		expected bool
	}{
		{"pending to picked_up", StatusPending, StatusPickedUp, true},
		{"pending to cancelled", StatusPending, StatusCancelled, true},
		{"pending to in_transit", StatusPending, StatusInTransit, false},
		{"pending to delivered", StatusPending, StatusDelivered, false},
		{"picked_up to in_transit", StatusPickedUp, StatusInTransit, true},
		{"picked_up to cancelled", StatusPickedUp, StatusCancelled, true},
		{"picked_up to delivered", StatusPickedUp, StatusDelivered, false},
		{"in_transit to delivered", StatusInTransit, StatusDelivered, true},
		{"in_transit to cancelled", StatusInTransit, StatusCancelled, true},
		{"in_transit to picked_up", StatusInTransit, StatusPickedUp, false},
		{"delivered to any", StatusDelivered, StatusPending, false},
		{"cancelled to any", StatusCancelled, StatusPending, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, tt.from.CanTransitionTo(tt.to))
		})
	}
}

func TestStatus_ValidateTransition(t *testing.T) {
	t.Run("valid transition", func(t *testing.T) {
		err := StatusPending.ValidateTransition(StatusPickedUp)
		assert.NoError(t, err)
	})

	t.Run("invalid transition", func(t *testing.T) {
		err := StatusPending.ValidateTransition(StatusDelivered)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "invalid transition")
	})

	t.Run("invalid status", func(t *testing.T) {
		err := StatusPending.ValidateTransition(Status("bogus"))
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "invalid status")
	})
}

func TestStatus_IsTerminal(t *testing.T) {
	assert.True(t, StatusDelivered.IsTerminal())
	assert.True(t, StatusCancelled.IsTerminal())
	assert.False(t, StatusPending.IsTerminal())
	assert.False(t, StatusPickedUp.IsTerminal())
	assert.False(t, StatusInTransit.IsTerminal())
}

func TestNewShipment(t *testing.T) {
	t.Run("valid shipment", func(t *testing.T) {
		s, err := NewShipment("REF-001", "New York", "Los Angeles", "John", "UNIT-1", 1000, 500)
		assert.NoError(t, err)
		assert.Equal(t, "REF-001", s.ReferenceNumber)
		assert.Equal(t, "New York", s.Origin)
		assert.Equal(t, "Los Angeles", s.Destination)
		assert.Equal(t, StatusPending, s.Status)
		assert.Equal(t, "John", s.DriverName)
		assert.Equal(t, "UNIT-1", s.UnitNumber)
		assert.Equal(t, 1000.0, s.ShipmentAmount)
		assert.Equal(t, 500.0, s.DriverRevenue)
		assert.False(t, s.CreatedAt.IsZero())
		assert.False(t, s.UpdatedAt.IsZero())
	})

	t.Run("missing reference number", func(t *testing.T) {
		_, err := NewShipment("", "A", "B", "", "", 0, 0)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "reference number")
	})

	t.Run("missing origin", func(t *testing.T) {
		_, err := NewShipment("REF", "", "B", "", "", 0, 0)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "origin")
	})

	t.Run("missing destination", func(t *testing.T) {
		_, err := NewShipment("REF", "A", "", "", "", 0, 0)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "destination")
	})
}
