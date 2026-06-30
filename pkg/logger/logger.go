package logger

import "go.uber.org/zap"

var log *zap.Logger

func Init() {
	var err error
	log, err = zap.NewProduction()
	if err != nil {
		panic(err)
	}
}

// get mengembalikan logger; auto-init bila belum diinisialisasi (mis. dipanggil
// dari script seed/release yang tidak memanggil Init lebih dulu).
func get() *zap.Logger {
	if log == nil {
		Init()
	}
	return log
}

// Sync flush buffer log sebelum aplikasi mati.
func Sync() {
	if log != nil {
		_ = log.Sync()
	}
}

func Info(msg string, fields ...zap.Field)  { get().Info(msg, fields...) }
func Error(msg string, fields ...zap.Field) { get().Error(msg, fields...) }
func Warn(msg string, fields ...zap.Field)  { get().Warn(msg, fields...) }
func Debug(msg string, fields ...zap.Field) { get().Debug(msg, fields...) }

func String(key, value string) zap.Field      { return zap.String(key, value) }
func Bool(key string, value bool) zap.Field   { return zap.Bool(key, value) }
func Int64(key string, value int64) zap.Field { return zap.Int64(key, value) }
func Int(key string, value int) zap.Field     { return zap.Int(key, value) }
func Uint(key string, value uint) zap.Field   { return zap.Uint(key, value) }
func Err(err error) zap.Field                 { return zap.Error(err) }
func Any(key string, value any) zap.Field     { return zap.Any(key, value) }
