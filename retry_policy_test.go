package workqueue

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestNopRetryPolicyImpl_NextDelay(t *testing.T) {
	policy := NewNopRetryPolicyImpl()

	delay, retry := policy.NextDelay("task", 1, nil)
	assert.Equal(t, time.Duration(0), delay)
	assert.False(t, retry)
}

func TestExponentialRetryPolicyImpl_NextDelay(t *testing.T) {
	policy := NewExponentialRetryPolicy(100*time.Millisecond, time.Second, 3)

	delay1, retry1 := policy.NextDelay("task", 1, nil)
	assert.True(t, retry1)
	assert.Equal(t, 100*time.Millisecond, delay1)

	delay2, retry2 := policy.NextDelay("task", 2, nil)
	assert.True(t, retry2)
	assert.Equal(t, 200*time.Millisecond, delay2)

	delay3, retry3 := policy.NextDelay("task", 3, nil)
	assert.True(t, retry3)
	assert.Equal(t, 400*time.Millisecond, delay3)

	delay4, retry4 := policy.NextDelay("task", 4, nil)
	assert.False(t, retry4)
	assert.Equal(t, time.Duration(0), delay4)
}

func TestExponentialRetryPolicyImpl_DelayCap(t *testing.T) {
	policy := NewExponentialRetryPolicy(500*time.Millisecond, 750*time.Millisecond, 10)

	delay, retry := policy.NextDelay("task", 2, nil)
	assert.True(t, retry)
	assert.Equal(t, 750*time.Millisecond, delay)
}
