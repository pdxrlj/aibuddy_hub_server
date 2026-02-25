package aiframe

type Frame interface {
	Encode() ([]byte, error)
	Decode(data []byte) error
}
