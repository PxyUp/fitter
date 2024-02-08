package builder_test

import (
	"bytes"
	"encoding/json"
	"github.com/PxyUp/fitter/pkg/builder"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestNumber(t *testing.T) {
	num := builder.Number(5)
	bb, err := json.Marshal(5)
	require.NoError(t, err)
	assert.True(t, bytes.Equal(bb, num.Raw()))
}
