package bidirpc

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
