package workqueue

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNopRateLimiterImpl_When(t *testing.T) {
	rl := NewNopRateLimiterImpl()

	// Call the When method
	delay := rl.When(nil).Milliseconds()

	assert.Equal(t, int64(0), delay, "delay should be 0")
}

func TestBucketRateLimiterImpl_When(t *testing.T) {
	rl := NewBucketRateLimiterImpl(5, 1)

	for i := 0; i < 10; i++ {
		// Call the When method
		delay := rl.When(nil).Milliseconds()

		assert.GreaterOrEqual(t, delay, int64(i*200-2), "delay should be greater than or equal to %d", i*200-2)
		assert.LessOrEqual(t, delay, int64(i*200), "delay should be less than or equal to %d", i*200)
	}
}
