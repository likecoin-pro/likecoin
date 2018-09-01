package hex

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestBytes_MarshalJSON(t *testing.T) {
	v := Bytes("\x00\x00\x00\x00\xff\xff\xff\xff")

	data, err := json.Marshal(v)

	assert.NoError(t, err)
	assert.Equal(t, `"00000000ffffffff"`, string(data))
}
