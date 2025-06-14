package clock

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestClock(t *testing.T) {
	provider := New()

	t1 := provider.Now()
	time.Sleep(time.Millisecond)
	t2 := provider.Now()

	assert.True(t, t2.After(t1), "second call to Now() should return later time")
}

func TestMockClock(t *testing.T) {
	mock := NewMockClock(t)
	baseTime := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)

	// Configure all expectations at the start
	mock.EXPECT().Now().Return(baseTime).Once()
	mock.EXPECT().Now().Return(baseTime.Add(time.Hour)).Once()
	mock.EXPECT().Now().Return(baseTime.Add(2 * time.Hour)).Once()
	mock.EXPECT().Now().Return(baseTime.Add(3 * time.Hour)).Once()
	mock.EXPECT().Now().Return(baseTime.Add(4 * time.Hour)).Once()

	// Test initial time
	assert.Equal(t, baseTime, mock.Now())

	// Test next call with different time
	assert.Equal(t, baseTime.Add(time.Hour), mock.Now())

	// Test multiple sequential calls
	assert.Equal(t, baseTime.Add(2*time.Hour), mock.Now())
	assert.Equal(t, baseTime.Add(3*time.Hour), mock.Now())
	assert.Equal(t, baseTime.Add(4*time.Hour), mock.Now())
}
