package java

import (
	"fmt"
	"log"
	"strings"
)

type OpCode struct {
	byte byte
	name string
	args []byte
}

func (op OpCode) Name() string {
	return op.name
}

func (op OpCode) String() string {
	result := make([]string, len(op.args)+1)
	result[0] = op.Name()
	for i, a := range op.args {
		result[i+1] = fmt.Sprintf("%v", a)
	}
	return strings.Join(result, " ")
}

func (op OpCode) Width() int {
	return len(op.args) + 1
}

func (op OpCode) int16() int16 {
	var x int16
	x |= int16(op.args[0]) << 8
	x |= int16(op.args[1])
	return x
}

func (op OpCode) uint16() uint16 {
	var x uint16
	x |= uint16(op.args[0]) << 8
	x |= uint16(op.args[1])
	return x
}

func (op OpCode) int8() byte {
	return op.args[0]
}

func bytesToOpcode(bytes []byte) OpCode {
	b := bytes[0]
	switch b {
	case 0:
		return OpCode{b, "nop", nil}
	case 1:
		return OpCode{b, "aconst_null", nil}
	case 3:
		return OpCode{b, "iconst_0", nil}
	case 4:
		return OpCode{b, "iconst_1", nil}
	case 5:
		return OpCode{b, "iconst_2", nil}
	case 6:
		return OpCode{b, "iconst_3", nil}
	case 7:
		return OpCode{b, "iconst_4", nil}
	case 8:
		return OpCode{b, "iconst_5", nil}
	case 13:
		return OpCode{b, "fconst_2", nil}
	case 15:
		return OpCode{b, "dconst_1", nil}
	case 16:
		return OpCode{b, "bipush", bytes[1:2]}
	case 18:
		return OpCode{b, "ldc", bytes[1:2]}
	case 20:
		return OpCode{b, "ldc2_w", bytes[1:3]}
	case 21:
		return OpCode{b, "iload", bytes[1:2]}
	case 25:
		return OpCode{b, "aload", bytes[1:2]}
	case 26:
		return OpCode{b, "iload_0", nil}
	case 27:
		return OpCode{b, "iload_1", nil}
	case 28:
		return OpCode{b, "iload_2", nil}
	case 29:
		return OpCode{b, "iload_3", nil}
	case 30:
		return OpCode{b, "lload_0", nil}
	case 31:
		return OpCode{b, "lload_1", nil}
	case 32:
		return OpCode{b, "lload_2", nil}
	case 33:
		return OpCode{b, "lload_3", nil}
	case 34:
		return OpCode{b, "fload_0", nil}
	case 35:
		return OpCode{b, "fload_1", nil}
	case 42:
		return OpCode{b, "aload_0", nil}
	case 43:
		return OpCode{b, "aload_1", nil}
	case 44:
		return OpCode{b, "aload_2", nil}
	case 45:
		return OpCode{b, "aload_3", nil}
	case 52:
		return OpCode{b, "caload", nil}
	case 54:
		return OpCode{b, "istore", bytes[1:2]}
	case 58:
		return OpCode{b, "astore", bytes[1:2]}
	case 60:
		return OpCode{b, "istore_1", nil}
	case 61:
		return OpCode{b, "istore_2", nil}
	case 62:
		return OpCode{b, "istore_3", nil}
	case 76:
		return OpCode{b, "astore_1", nil}
	case 77:
		return OpCode{b, "astore_2", nil}
	case 78:
		return OpCode{b, "astore_3", nil}
	case 85:
		return OpCode{b, "castore", nil}
	case 87:
		return OpCode{b, "pop", nil}
	case 89:
		return OpCode{b, "dup", nil}
	case 90:
		return OpCode{b, "dup_x1", nil}
	case 96:
		return OpCode{b, "iadd", nil}
	case 97:
		return OpCode{b, "ladd", nil}
	case 98:
		return OpCode{b, "fadd", nil}
	case 100:
		return OpCode{b, "isub", nil}
	case 101:
		return OpCode{b, "lsub", nil}
	case 102:
		return OpCode{b, "fsub", nil}
	case 104:
		return OpCode{b, "imul", nil}
	case 105:
		return OpCode{b, "lmul", nil}
	case 106:
		return OpCode{b, "fmul", nil}
	case 108:
		return OpCode{b, "idiv", nil}
	case 109:
		return OpCode{b, "ldiv", nil}
	case 110:
		return OpCode{b, "fdiv", nil}
	case 132:
		return OpCode{b, "iinc", bytes[1:3]}
	case 153:
		return OpCode{b, "ifeq", bytes[1:3]}
	case 154:
		return OpCode{b, "ifne", bytes[1:3]}
	case 156:
		return OpCode{b, "ifge", bytes[1:3]}
	case 157:
		return OpCode{b, "ifgt", bytes[1:3]}
	case 158:
		return OpCode{b, "ifle", bytes[1:3]}
	case 160:
		return OpCode{b, "if_icmpne", bytes[1:3]}
	case 162:
		return OpCode{b, "if_icmpge", bytes[1:3]}
	case 163:
		return OpCode{b, "if_icmpgt", bytes[1:3]}
	case 164:
		return OpCode{b, "if_icmple", bytes[1:3]}
	case 167:
		return OpCode{b, "goto", bytes[1:3]}
	case 172:
		return OpCode{b, "ireturn", nil}
	case 173:
		return OpCode{b, "lreturn", nil}
	case 174:
		return OpCode{b, "freturn", nil}
	case 176:
		return OpCode{b, "areturn", nil}
	case 177:
		return OpCode{b, "return", nil}
	case 178:
		return OpCode{b, "getstatic", bytes[1:3]}
	case 179:
		return OpCode{b, "putstatic", bytes[1:3]}
	case 180:
		return OpCode{b, "getfield", bytes[1:3]}
	case 181:
		return OpCode{b, "putfield", bytes[1:3]}
	case 182:
		return OpCode{b, "invokevirtual", bytes[1:3]}
	case 183:
		return OpCode{b, "invokespecial", bytes[1:3]}
	case 184:
		return OpCode{b, "invokestatic", bytes[1:3]}
	case 185:
		return OpCode{b, "invokeinterface", bytes[1:5]}
	case 187:
		return OpCode{b, "new", bytes[1:3]}
	case 188:
		return OpCode{b, "newarray", bytes[1:2]}
	case 189:
		return OpCode{b, "anewarray", bytes[1:3]}
	case 190:
		return OpCode{b, "arraylength", nil}
	case 191:
		return OpCode{b, "athrow", nil}
	case 192:
		return OpCode{b, "checkcast", bytes[1:3]}
	case 193:
		return OpCode{b, "instanceof", bytes[1:3]}
	case 194:
		return OpCode{b, "monitorenter", nil}
	case 195:
		return OpCode{b, "monitorexit", nil}
	case 198:
		return OpCode{b, "ifnull", bytes[1:3]}
	case 199:
		return OpCode{b, "ifnonnull", bytes[1:3]}
	default:
		log.Panicf("Unknown instruction: %v", b)
	}
	return OpCode{}
}

type ProgramCounter struct {
	RawByteCodes     []byte
	RawByteCodeIndex int
	OpCodeIndex      int
	OpCodes          []OpCode
}

func opsFromBytes(bytes []byte) []OpCode {
	var ops []OpCode
	i := 0
	for i < len(bytes) {
		op := bytesToOpcode(bytes[i:])
		ops = append(ops, op)
		i += op.Width()
	}
	return ops
}

func newProgramCounter(bytes []byte) ProgramCounter {
	return ProgramCounter{bytes, 0, 0, opsFromBytes(bytes)}
}

func (pc *ProgramCounter) OpCode() OpCode {
	return pc.OpCodes[pc.OpCodeIndex]
}

func (pc *ProgramCounter) CurrentByteCodeIndex() int {
	return pc.RawByteCodeIndex - pc.OpCode().Width()
}

func (pc *ProgramCounter) next() OpCode {
	op := pc.OpCode()
	pc.RawByteCodeIndex += op.Width()
	pc.OpCodeIndex++
	return op
}

func (pc *ProgramCounter) jumpTo(index int) {
	pc.jump(index - pc.RawByteCodeIndex + pc.OpCode().Width())
}

func (pc *ProgramCounter) DebugOut() {
	index := 0
	for i, o := range pc.OpCodes {
		prefix := "\t"
		if i == pc.OpCodeIndex-1 {
			prefix = "->\t"
		}
		log.Printf("%s %3d %3d %v", prefix, index, i, o)
		index += o.Width()
	}
}

func (pc *ProgramCounter) jump(offset int) {
	pc.OpCodeIndex--
	pc.RawByteCodeIndex -= pc.OpCode().Width()
	pc.RawByteCodeIndex += offset

	if offset < 0 {
		for offset < 0 {
			pc.OpCodeIndex--
			offset += pc.OpCode().Width()
		}
	} else {
		for offset > 0 {
			offset -= pc.OpCode().Width()
			pc.OpCodeIndex++
		}
	}
}
