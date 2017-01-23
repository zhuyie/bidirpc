package bidirpc

import "io"

func encodeHeader(buffer []byte, streamType byte, bodyLen int) {
	buffer[0] = streamType
	buffer[1] = byte(bodyLen >> 16)
	buffer[2] = byte(bodyLen >> 8)
	buffer[3] = byte(bodyLen)
}

func decodeHeader(buffer []byte) (streamType byte, bodyLen int) {
	streamType = buffer[0]
	bodyLen = int(buffer[1])<<16 | int(buffer[2])<<8 | int(buffer[3])
	return
}

func copyAtLeast(dst io.Writer, src io.Reader, min int) (n int, err error) {
	for n < min && err == nil {
		var nn int64
		nn, err = io.Copy(dst, src)
		if nn == 0 && err == nil {
			err = io.ErrUnexpectedEOF
		}
		n += int(nn)
	}
	if n >= min {
		err = nil
	} else if n > 0 && err == io.EOF {
		err = io.ErrUnexpectedEOF
	}
	return
}
