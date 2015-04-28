package java

import "log"

type opcode struct {
	byte byte
	name string
	args []byte
}

func (op *opcode) width() int {
	return len(op.args) + 1
}

func (op opcode) int16() int16 {
	var x int16
	x |= int16(op.args[0]) << 8
	x |= int16(op.args[1])
	return x
}

func (op opcode) uint16() uint16 {
	var x uint16
	x |= uint16(op.args[0]) << 8
	x |= uint16(op.args[1])
	return x
}

func (op opcode) int8() byte {
	return op.args[0]
}

func bytesToOpcode(bytes []byte) opcode {
	b := bytes[0]
	switch b {
	case 0:
		return opcode{b, "nop", nil}
	case 3:
		return opcode{b, "iconst_0", nil}
	case 4:
		return opcode{b, "iconst_1", nil}
	case 5:
		return opcode{b, "iconst_2", nil}
	case 6:
		return opcode{b, "iconst_3", nil}
	case 7:
		return opcode{b, "iconst_4", nil}
	case 8:
		return opcode{b, "iconst_5", nil}
	case 16:
		return opcode{b, "bipush", bytes[1:2]}
	case 18:
		return opcode{b, "ldc", bytes[1:2]}
	case 21:
		return opcode{b, "iload", bytes[1:2]}
	case 26:
		return opcode{b, "iload_0", nil}
	case 27:
		return opcode{b, "iload_1", nil}
	case 28:
		return opcode{b, "iload_2", nil}
	case 29:
		return opcode{b, "iload_3", nil}
	case 42:
		return opcode{b, "aload_0", nil}
	case 43:
		return opcode{b, "aload_1", nil}
	case 44:
		return opcode{b, "aload_2", nil}
	case 45:
		return opcode{b, "aload_3", nil}
	case 52:
		return opcode{b, "caload", nil}
	case 54:
		return opcode{b, "istore", bytes[1:2]}
	case 60:
		return opcode{b, "istore_1", nil}
	case 61:
		return opcode{b, "istore_2", nil}
	case 62:
		return opcode{b, "istore_3", nil}
	case 76:
		return opcode{b, "astore_1", nil}
	case 78:
		return opcode{b, "astore_3", nil}
	case 85:
		return opcode{b, "castore", nil}
	case 89:
		return opcode{b, "dup", nil}
	case 96:
		return opcode{b, "iadd", nil}
	case 100:
		return opcode{b, "isub", nil}
	case 104:
		return opcode{b, "imul", nil}
	case 108:
		return opcode{b, "idiv", nil}
	case 132:
		return opcode{b, "iinc", bytes[1:3]}
	case 154:
		return opcode{b, "ifne", bytes[1:3]}
	case 162:
		return opcode{b, "if_icmpge", bytes[1:3]}
	case 164:
		return opcode{b, "if_icmple", bytes[1:3]}
	case 167:
		return opcode{b, "goto", bytes[1:3]}
	case 172:
		return opcode{b, "ireturn", nil}
	case 176:
		return opcode{b, "areturn", nil}
	case 177:
		return opcode{b, "return", nil}
	case 178:
		return opcode{b, "getstatic", bytes[1:3]}
	case 179:
		return opcode{b, "putstatic", bytes[1:3]}
	case 180:
		return opcode{b, "getfield", bytes[1:3]}
	case 181:
		return opcode{b, "putfield", bytes[1:3]}
	case 182:
		return opcode{b, "invokevirtual", bytes[1:3]}
	case 183:
		return opcode{b, "invokespecial", bytes[1:3]}
	case 184:
		return opcode{b, "invokestatic", bytes[1:3]}
	case 187:
		return opcode{b, "new", bytes[1:3]}
	case 188:
		return opcode{b, "newarray", bytes[1:2]}
	case 190:
		return opcode{b, "arraylength", nil}
	default:
		log.Panicf("Unknown instruction: %v", b)
	}
	return opcode{}
}

type programCounter struct {
	bytecodes []byte
	pos       int
	lastop    opcode
}

func newProgramCounter(bytes []byte) programCounter {
	return programCounter{bytes, 0, opcode{}}
}

func (pc *programCounter) next() opcode {
	op := bytesToOpcode(pc.bytecodes[pc.pos:])
	pc.pos += op.width()
	pc.lastop = op
	return op
}

func (pc *programCounter) jump(offset int) {
	pc.pos -= pc.lastop.width()
	pc.pos += offset
}
