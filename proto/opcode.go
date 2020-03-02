package proto

const (
	// SM_DETACH_STATUS is server message.
	//
	// Data structure:
	//  0 byte: uint8, 1 detached 0 attached
	SM_DETACH_STATUS uint16 = iota + 1

	// SM_ATTACH_ACK is server message.
	//
	// Data structure:
	//  0 byte: uint8, 1 accept, otherwise denied
	//  1 byte: []byte, reason for denied
	SM_ATTACH_ACK

	// CM_SCREEN_SIZE is client message.
	//
	// Data stucture:
	//  0-1 byte: uint16, screen rows number
	//  2-3 byte: uint16, screen columns number
	CM_SCREEN_SIZE

	// CM_USER_INPUT is client message.
	//
	// Data structure:
	//  []byte, command string
	CM_USER_INPUT

	// CM_QUERY_DETACH_STATUS is client message.
	//
	// Data structure:
	//  no data
	CM_QUERY_DETACH_STATUS

	// CM_ATTACH_REQ is client message.
	//
	// Data structure:
	//  0 byte: uint8, 1 detach other and attach, 0 only attach
	CM_ATTACH_REQ
)
