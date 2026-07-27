package utils_test

import (
	"github.com/PxyUp/fitter/pkg/builder"
	"github.com/PxyUp/fitter/pkg/utils"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestProcessCondition(t *testing.T) {
	index := uint32(2)

	cases := []struct {
		name       string
		expression string
		result     builder.Interfacable
		index      *uint32
		expected   bool
		hasError   bool
	}{
		{
			name:       "true condition on number",
			expression: "fRes > 5",
			result:     builder.Number(10),
			expected:   true,
		},
		{
			name:       "false condition on number",
			expression: "fRes > 5",
			result:     builder.Number(3),
			expected:   false,
		},
		{
			name:       "string equality",
			expression: `fRes == "green"`,
			result:     builder.String("green"),
			expected:   true,
		},
		{
			name:       "condition by index",
			expression: "fIndex != 1",
			result:     builder.Number(1),
			index:      &index,
			expected:   true,
		},
		{
			name:       "non-boolean result is false",
			expression: "fRes + 1",
			result:     builder.Number(1),
			expected:   false,
		},
		{
			name:       "invalid expression returns error",
			expression: "fRes >>> nonsense",
			result:     builder.Number(1),
			expected:   false,
			hasError:   true,
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			pass, err := utils.ProcessCondition(c.expression, c.result, c.index, nil)
			if c.hasError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
			assert.Equal(t, c.expected, pass)
		})
	}
}
