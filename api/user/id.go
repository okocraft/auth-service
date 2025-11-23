package user

import (
	"strconv"
)

type ID int32

func (id ID) String() string {
	return strconv.FormatInt(int64(id), 10)
}
