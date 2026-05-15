package voyage_test

import (
	"testing"
	"time"

	"github.com/roma-glushko/cargo/internal/voyage"
	"github.com/stretchr/testify/require"
)

func TestBuilder(t *testing.T) {
	dep1 := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
	arr1 := time.Date(2024, 1, 2, 0, 0, 0, 0, time.UTC)
	dep2 := time.Date(2024, 1, 3, 0, 0, 0, 0, time.UTC)
	arr2 := time.Date(2024, 1, 4, 0, 0, 0, 0, time.UTC)

	v := voyage.NewBuilder("V100", "CNHKG").
		AddMovement("JNTKO", dep1, arr1).
		AddMovement("USNYC", dep2, arr2).
		Build()

	require.Equal(t, voyage.Number("V100"), v.Number)
	require.Len(t, v.Schedule.Movements, 2)

	m0 := v.Schedule.Movements[0]
	require.Equal(t, "CNHKG", string(m0.DepartureLocation))
	require.Equal(t, "JNTKO", string(m0.ArrivalLocation))
	require.Equal(t, dep1, m0.DepartureTime)
	require.Equal(t, arr1, m0.ArrivalTime)

	m1 := v.Schedule.Movements[1]
	require.Equal(t, "JNTKO", string(m1.DepartureLocation))
	require.Equal(t, "USNYC", string(m1.ArrivalLocation))
	require.Equal(t, dep2, m1.DepartureTime)
	require.Equal(t, arr2, m1.ArrivalTime)
}
