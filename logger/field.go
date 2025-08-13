package logger

func Error(err error) Field {
	return Field{
		Key:  "error",
		Data: err,
	}
}
func String(key string, data string) Field {
	return Field{
		Key:  key,
		Data: data,
	}
}
func Int64(key string, data int64) Field {
	return Field{
		Key:  key,
		Data: data,
	}
}

func Int32(key string, data int32) Field {
	return Field{
		Key:  key,
		Data: data,
	}
}
