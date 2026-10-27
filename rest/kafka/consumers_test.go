package kafka

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSafeTopicBoundsAndEscapes(t *testing.T) {
	topic := safeTopic("events\n" + strings.Repeat("x", 200))
	assert.NotContains(t, topic, "\n")
	assert.LessOrEqual(t, len(topic), 128)
	assert.LessOrEqual(t, len(safeTopic(strings.Repeat("é", 100))), 128)
}
