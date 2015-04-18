package java

import "log"

type opcode struct {
	byte  byte
	index int
	name  string
	args  []byte
}

func (o *opcode) width() int {
	return len(o.args) + 1
}

func bytesToOpcodes(bytes []byte) []opcode {
	opcodes := make([]opcode, 0, len(bytes))
	i := 0
	for i < len(bytes) {
		b := bytes[i]
		switch b {
		case 0:
			o := opcode{b, i, "nop", nil}
			opcodes = append(opcodes, o)
			i += o.width()
		case 4:
			o := opcode{b, i, "iconst_1", nil}
			opcodes = append(opcodes, o)
			i += o.width()
		case 5:
			o := opcode{b, i, "iconst_2", nil}
			opcodes = append(opcodes, o)
			i += o.width()
		case 8:
			o := opcode{b, i, "iconst_5", nil}
			opcodes = append(opcodes, o)
			i += o.width()
		case 16:
			o := opcode{b, i, "bipush", bytes[i+1 : i+2]}
			opcodes = append(opcodes, o)
			i += o.width()
		case 18:
			o := opcode{b, i, "ldc", bytes[i+1 : i+2]}
			opcodes = append(opcodes, o)
			i += o.width()
		case 26:
			o := opcode{b, i, "iload_0", nil}
			opcodes = append(opcodes, o)
			i += o.width()
		case 27:
			o := opcode{b, i, "iload_1", nil}
			opcodes = append(opcodes, o)
			i += o.width()
		case 42:
			o := opcode{b, i, "aload_0", nil}
			opcodes = append(opcodes, o)
			i += o.width()
		case 43:
			o := opcode{b, i, "aload_1", nil}
			opcodes = append(opcodes, o)
			i += o.width()
		case 60:
			o := opcode{b, i, "istore_1", nil}
			opcodes = append(opcodes, o)
			i += o.width()
		case 76:
			o := opcode{b, i, "astore_1", nil}
			opcodes = append(opcodes, o)
			i += o.width()
		case 89:
			o := opcode{b, i, "dup", nil}
			opcodes = append(opcodes, o)
			i += o.width()
		case 96:
			o := opcode{b, i, "iadd", nil}
			opcodes = append(opcodes, o)
			i += o.width()
		case 100:
			o := opcode{b, i, "isub", nil}
			opcodes = append(opcodes, o)
			i += o.width()
		case 104:
			o := opcode{b, i, "imul", nil}
			opcodes = append(opcodes, o)
			i += o.width()
		case 108:
			o := opcode{b, i, "idiv", nil}
			opcodes = append(opcodes, o)
			i += o.width()
		case 172:
			o := opcode{b, i, "ireturn", nil}
			opcodes = append(opcodes, o)
			i += o.width()
		case 177:
			o := opcode{b, i, "return", nil}
			opcodes = append(opcodes, o)
			i += o.width()
		case 178:
			o := opcode{b, i, "getstatic", bytes[i+1 : i+3]}
			opcodes = append(opcodes, o)
			i += o.width()
		case 179:
			o := opcode{b, i, "putstatic", bytes[i+1 : i+3]}
			opcodes = append(opcodes, o)
			i += o.width()
		case 180:
			o := opcode{b, i, "getfield", bytes[i+1 : i+3]}
			opcodes = append(opcodes, o)
			i += o.width()
		case 181:
			o := opcode{b, i, "putfield", bytes[i+1 : i+3]}
			opcodes = append(opcodes, o)
			i += o.width()
		case 182:
			o := opcode{b, i, "invokevirtual", bytes[i+1 : i+3]}
			opcodes = append(opcodes, o)
			i += o.width()
		case 183:
			o := opcode{b, i, "invokespecial", bytes[i+1 : i+3]}
			opcodes = append(opcodes, o)
			i += o.width()
		case 184:
			o := opcode{b, i, "invokestatic", bytes[i+1 : i+3]}
			opcodes = append(opcodes, o)
			i += o.width()
		case 187:
			o := opcode{b, i, "new", bytes[i+1 : i+3]}
			opcodes = append(opcodes, o)
			i += o.width()
		default:
			log.Panicf("Unknown instruction: %v", b)
		}
	}
	return opcodes
}
