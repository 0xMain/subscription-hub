package pagination

const (
	DefaultLimit = 10
	MaxLimit     = 100
)

func Normalize(limit, offset int) (int, int) {
	if limit <= 0 {
		limit = DefaultLimit
	}

	if limit > MaxLimit {
		limit = MaxLimit
	}

	if offset < 0 {
		offset = 0
	}

	return limit, offset
}

func CalculateCapacity(total int64, limit, offset int) int {
	if total <= 0 || int64(offset) >= total {
		return 0
	}

	remaining := int(total - int64(offset))
	return min(remaining, limit)
}
