package logger

import "time"

func Error(err error) Field {
	return Field{
		Key:  "error",
		Data: err.Error(), // 一般日志里只打 string，避免 interface{} 打出来一坨
	}
}

func String(key, data string) Field {
	return Field{
		Key:  key,
		Data: data,
	}
}

func Bool(key string, data bool) Field {
	return Field{
		Key:  key,
		Data: data,
	}
}

func Int(key string, data int) Field {
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

func Int64(key string, data int64) Field {
	return Field{
		Key:  key,
		Data: data,
	}
}

func Uint64(key string, data uint64) Field {
	return Field{
		Key:  key,
		Data: data,
	}
}

func Float32(key string, data float32) Field {
	return Field{
		Key:  key,
		Data: data,
	}
}

func Float64(key string, data float64) Field {
	return Field{
		Key:  key,
		Data: data,
	}
}

func Duration(key string, data time.Duration) Field {
	return Field{
		Key:  key,
		Data: data.String(),
	}
}

func Time(key string, t time.Time) Field {
	return Field{
		Key:  key,
		Data: t.Format(time.RFC3339), // 日志里一般转成 string，方便收集
	}
}

func Any(key string, data any) Field {
	return Field{
		Key:  key,
		Data: data,
	}
}
