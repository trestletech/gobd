package gobd

type SerialPort interface {
	Read([]byte) (int, error)
	Write([]byte) (int, error)
}
