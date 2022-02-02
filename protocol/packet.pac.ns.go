package protocol

type packetNS string

func (x packetNS) Read(p []byte) (n int, err error) {
	if len(x) == 0 {
		return
	}

	if x == "/" {
		return // a single "/" is the same as empty
	}

	if n = copy(p, x); n < len(x) {
		return n, PacketError{str: "buffer namespace for read", buffer: []byte(x)[n:], errs: []error{ErrShortRead}}
	}

	return n, nil
}

func (x *packetNS) Write(p []byte) (n int, err error) {
	if len(p) == 0 {
		if len(*x) == 0 {
			*x = "/"
		}
		return
	}

	var data []byte
	if x != nil && len(*x) > 0 { // this means we have a short write, so pick up where we left off...
		data = []byte(*x)
	}

	var size = len(data)
	for i, val := range p {
		if i == 0 && val != '/' && len(data) == 0 {
			*x = packetNS("/")
			return 0, nil // ErrNoNamespace
		}
		switch val {
		case ',':
			*x = packetNS(string(data))
			return i + 1, nil
		}
		data = append(data, val)
	}

	if (len(data) - size) == len(p) {
		*x = packetNS(string(data))
		return len(data) - size, nil
	}

	return len(data) - size, ErrShortWrite
}
