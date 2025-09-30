package store

func nullableJSON(data []byte) interface{} {
	if len(data) == 0 {
		return nil
	}
	return data
}
