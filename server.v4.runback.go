package socketio

import (
	"fmt"

	siop "github.com/njones/socketio/protocol"
	siot "github.com/njones/socketio/transport"
)

func runV4(v4 *ServerV4) func(r *Request, socketID siot.SocketID) error {
	return func(r *Request, socketID siot.SocketID) error {
		v2 := v4.prev.prev
		v1 := v2.prev
	Receive:
		for socket := range v1.transport.Receive(socketID) {
			switch socket.Type {
			case siop.ConnectPacket.Byte():
				if err := v1.doConnectPacket(r, socketID, socket); err != nil {
					v1.transport.Send(socketID, serviceError(err), siop.WithType(byte(siop.ErrorPacket)))
				}
			case siop.DisconnectPacket.Byte():
				break Receive
			case siop.EventPacket.Byte():
				if err := v1.doEventPacket(socket); err != nil {
					v1.transport.Send(socketID, serviceError(err), siop.WithType(byte(siop.ErrorPacket)))
				}
			case siop.AckPacket.Byte():
				if err := v1.doAckPacket(socket); err != nil {
					v1.transport.Send(socketID, serviceError(err), siop.WithType(byte(siop.ErrorPacket)))
				}
			case siop.ErrorPacket.Byte():
				if e, ok := socket.Data.(error); ok {
					return e
				}
			case siop.BinaryEventPacket.Byte():
				if err := v2.doBinaryEventPacket(socket); err != nil {
					v1.transport.Send(socketID, serviceError(err), siop.WithType(byte(siop.ErrorPacket)))
				}
			default:
				return fmt.Errorf("invalid packet type: %#v", socket)
			}
		}
		return nil
	}
}

func doConnectPacketV4(v4 *ServerV4) func(*Request, SocketID, siot.Socket) error {
	return func(req *Request, socketID SocketID, socket siot.Socket) (err error) {
		v1 := v4.prev.prev.prev
		transport := v1.tr()
		transport.Join(socket.Namespace, socketID, socketID.Room(socketIDPrefix))

		socketV4 := &SocketV4{ID: socketID, req: req}
		return v4.onConnect[socket.Namespace](socketV4)
	}
}
