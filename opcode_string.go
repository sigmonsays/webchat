// Code generated by "stringer -type=OpCode"; DO NOT EDIT

package webchat

import "fmt"

const _OpCode_name = "InvalidOpRegisterOpUnregisterOpMessageOpNoticeOpJoinOpNickOpPingOp"

var _OpCode_index = [...]uint8{0, 9, 19, 31, 40, 48, 54, 60, 66}

func (i OpCode) String() string {
	if i < 0 || i >= OpCode(len(_OpCode_index)-1) {
		return fmt.Sprintf("OpCode(%d)", i)
	}
	return _OpCode_name[_OpCode_index[i]:_OpCode_index[i+1]]
}
