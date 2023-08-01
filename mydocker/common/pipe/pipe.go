package pipe

import (
	"fmt"
	"io/ioutil"
	"os"
)

// NewPipe 创建管道
func NewPipe() (*os.File, *os.File, error) {
	r, w, err := os.Pipe()
	if err != nil {
		return nil, nil, err
	}
	return r, w, err
}

// WritePipe 写管道
func WritePipe(writePipe *os.File, msg string) error {
	if _, err := writePipe.WriteString(msg); err != nil {
		return fmt.Errorf("Write pipe error: ", err.Error())
	}
	return nil
}

// ReadPipe 读管道
func ReadPipe(readPipe *os.File) ([]byte, error) {
	msg, err := ioutil.ReadAll(readPipe)
	if err != nil {
		return []byte{}, fmt.Errorf("Read pipe error:", err.Error())
	}
	return msg, nil
}

func ClosePipe(pipe *os.File) error {
	return pipe.Close()
}
