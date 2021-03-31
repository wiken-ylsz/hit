package util

import (
	"strconv"

	"github.com/sony/sonyflake"
)

var flake = sonyflake.NewSonyflake(sonyflake.Settings{})

// UUID 唯一ID
func UUID() string {
	var id uint64
	for {
		_id, err := flake.NextID()
		if err == nil {
			id = _id
			break
		}
	}

	return strconv.FormatUint(id, 32)
}
