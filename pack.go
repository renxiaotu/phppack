package phppack

import (
	"encoding/hex"
	"errors"
	"fmt"
	"github.com/renxiaotu/dtc/tobytes"
	"reflect"
	"strings"
)

func PackByStruct(data interface{}) ([]byte, error) {
	b := make([]byte, 0)
	value := reflect.ValueOf(data)
	for value.Kind() == reflect.Ptr {
		next := value.Elem().Kind()
		if next == reflect.Struct || next == reflect.Ptr {
			value = value.Elem()
		} else {
			break
		}
	}
	pts := make([]packType, 0)
	err := errors.New("")
	switch value.Kind() {
	case reflect.Struct:
		pts, err = parseTypes(value)
		if err != nil {
			return nil, err
		}
		break
	default:
		return nil, errors.New(PackageName + ":unsupported data type")
	}

	fmt.Println(pts)

	for i := 0; i < len(pts); i++ {
		v := value.FieldByName(pts[i].Name).Interface()
		pt := pts[i]
		sub, err := pack(&b, pt, v)
		if err != nil {
			return b, err
		}
		b = append(b, sub...)
	}

	return b, nil
}

func PackByFormat(f string, args ...interface{}) ([]byte, error) {
	b := make([]byte, 0)
	pts, err := parsePackFormats(f)
	if err != nil {
		return nil, err
	}

	for i := 0; i < len(pts); i++ {
		pt := pts[i]
		if pt.tag.Size == -1 {
			if len(args) == 0 {
				return nil, errNea
			}
			if strings.Contains("aAhHxX@", pt.tag.Type) {
				pt.tag.Size = len(args[0].(string))
			} else {
				pt.tag.Size = len(args)
			}
		}

		if pt.tag.Size > 1 && !strings.Contains("aAhHxX@", pt.tag.Type) {
			pt.tag.Size--
			i--
		}

		v := args[i]

		sub, err := pack(&b, pt, v)
		if err != nil {
			return b, err
		}

		b = append(b, sub...)
	}

	return b, nil
}

func pack(b *[]byte, pt packType, v interface{}) ([]byte, error) {
	switch pt.tag.Type {
	//--------------------------------------------字符串--------------------------
	case "a": //以NUL字节填充字符串
		return i2a(v, pt, false)
	case "A": //以SPACE(空格)填充字符串
		return i2a(v, pt, true)

	//--------------------------------------------hex--------------------------
	case "h": //十六进制字符串，低位在前
		return i2h(v, pt, false)
	case "H": //十六进制字符串，高位在前
		return i2h(v, pt, true)

	//--------------------------------------------8bit--------------------------
	case "c": //有符号字符 int8
		return interface2c(v)
	case "C": //无符号字符 uint8
		return interface2C(v)

	//--------------------------------------------16bit--------------------------
	case "s": //有符号短整型(16位，主机字节序)
		return interface2s(v)
	case "S": //无符号短整型(16位，主机字节序)
		return interface2S(v)

	case "n": //无符号短整型(16位，大端字节序)
		return interface2n(v)
	case "v": //无符号短整型(16位，小端字节序)
		return interface2v(v)

	//--------------------------------------------this bit--------------------------
	case "i": //有符号整型(机器相关大小字节序)
		return interface2i(v)
	case "I": //无符号整型(机器相关大小字节序)
		return interface2I(v)

	//--------------------------------------------32bit--------------------------
	case "l": //有符号长整型(32位，主机字节序)
		return interface2l(v)
	case "L": //无符号长整型(32位，主机字节序)
		return interface2L(v)
	case "N": //无符号长整型(32位，大端字节序)
		return interface2N(v)
	case "V": //无符号长整型(32位，小端字节序)
		return interface2V(v)

	//--------------------------------------------64bit--------------------------
	case "q": //有符号长长整型(64位，主机字节序)
		return interface2q(v)
	case "Q": //无符号长长整型(64位，主机字节序)
		return interface2Q(v)

	case "J": //无符号长长整型(64位，大端字节序)
		return interface2J(v)
	case "P": //无符号长长整型(64位，小端字节序)
		return interface2P(v)

	//--------------------------------------------float--------------------------
	case "f": //单精度浮点型(主机字节序)
		return interface2f(v)
	case "g": //单精度浮点型(小端字节序)
		return interface2g(v)
	case "G": //单精度浮点型(大端字节序)
		return interface2G(v)

	case "d": //双精度浮点型(主机字节序)
		return interface2d(v)
	case "e": //双精度浮点型(小端字节序)
		return interface2e(v)
	case "E": //双精度浮点型(大端字节序)
		return interface2E(v)

	//--------------------------------------------other--------------------------
	case "x": //NUL字节
		return x(pt.tag.Size), nil
	case "X": //回退字节
		X(b, pt.tag.Size)
		return make([]byte, 0), nil
	case "Z": //a的别名
		return i2a(v, pt, false)
	case "@": //在绝对位置填充0到末尾
		at(b, pt.tag.Size)
		return make([]byte, 0), nil
	default: //不支持的格式
		return nil, errors.New("format contains characters that are not supported")
	}
}

func i2a(v interface{}, pt packType, isSpace bool) ([]byte, error) {
	str := ""
	switch v.(type) {
	case string:
		str = v.(string)
		break
	default:
		return nil, errT()
	}
	l := pt.tag.Size
	if l == -1 {
		l = len(str)
	}
	b := make([]byte, l)

	if isSpace {
		for i := 0; i < len(b); i++ {
			b[i] = 32
		}
	}

	copy(b, str)
	return b, nil
}

func i2h(v interface{}, pt packType, Big bool) ([]byte, error) {
	str := ""
	switch v.(type) {
	case string:
		str = v.(string)
		break
	default:
		return nil, errT()
	}

	l := pt.tag.Size
	if l == -1 {
		l = len(str)
	}
	b := make([]byte, l/2+l%2)
	bLen := len(b) * 2
	h := []byte(str)
	if len(h) > bLen {
		h = h[:bLen]
	}
	if len(h) < l {
		return b, errL()
	}
	if len(h) > l {
		h[l] = 48
	}
	if !Big {
		for bLen > 0 {
			tmp := h[bLen-1]
			h[bLen-1] = h[bLen-2]
			h[bLen-2] = tmp
			bLen -= 2
		}
	}
	data, err := hex.DecodeString(string(h))
	if err != nil {
		return b, errT()
	}
	copy(b, data)
	return b, nil
}

func interface2c(v interface{}) ([]byte, error) {
	n := int8(0)
	switch v.(type) {
	case string:
	case int:
		n = int8(v.(int))
		break
	case int8:
		n = v.(int8)
		break
	default:
		return nil, errT()
	}
	b := make([]byte, 0)
	b = append(b, byte(n))
	return b, nil
}

func interface2C(v interface{}) ([]byte, error) {
	n := uint8(0)
	switch v.(type) {
	case string:
	case int:
		n = uint8(v.(int))
		break
	case uint8:
		n = v.(uint8)
		break
	default:
		return nil, errT()
	}
	b := make([]byte, 0)
	b = append(b, n)
	return b, nil
}

func interface2Int16(v interface{}, e tobytes.Endian) ([]byte, error) {
	n := int16(0)
	switch v.(type) {
	case string:
	case int:
		n = int16(v.(int))
		break
	case int16:
		n = v.(int16)
		break
	default:
		return nil, errT()
	}
	return tobytes.Int16ToBytes(n, e), nil
}

func interface2Uint16(v interface{}, e tobytes.Endian) ([]byte, error) {
	n := uint16(0)
	switch v.(type) {
	case string:
	case int:
		n = uint16(v.(int))
		break
	case uint16:
		n = v.(uint16)
		break
	default:
		return nil, errT()
	}
	return tobytes.Uint16ToBytes(n, e), nil
}

func interface2Int(v interface{}, e tobytes.Endian) ([]byte, error) {
	n := 0
	switch v.(type) {
	case string:
	case int:
		n = v.(int)
		break
	default:
		return nil, errT()
	}
	return tobytes.IntToBytes(n, e), nil
}

func interface2Uint(v interface{}, e tobytes.Endian) ([]byte, error) {
	n := uint(0)
	switch v.(type) {
	case string:
	case int:
		n = uint(v.(int))
		break
	case int16:
		n = v.(uint)
		break
	default:
		return nil, errT()
	}
	return tobytes.UintToBytes(n, e), nil
}

func interface2Int32(v interface{}, e tobytes.Endian) ([]byte, error) {
	n := int32(0)
	switch v.(type) {
	case string:
	case int:
		n = int32(v.(int))
		break
	case int16:
		n = v.(int32)
		break
	default:
		return nil, errT()
	}
	return tobytes.Int32ToBytes(n, e), nil
}

func interface2Uint32(v interface{}, e tobytes.Endian) ([]byte, error) {
	n := uint32(0)
	switch v.(type) {
	case string:
	case int:
		n = uint32(v.(int))
		break
	case uint32:
		n = v.(uint32)
		break
	default:
		return nil, errT()
	}
	return tobytes.Uint32ToBytes(n, e), nil
}

func interface2Int64(v interface{}, e tobytes.Endian) ([]byte, error) {
	n := int64(0)
	switch v.(type) {
	case string:
	case int:
		n = int64(v.(int))
		break
	case int16:
		n = v.(int64)
		break
	default:
		return nil, errT()
	}
	return tobytes.Int64ToBytes(n, e), nil
}

func interface2Uint64(v interface{}, e tobytes.Endian) ([]byte, error) {
	n := uint64(0)
	switch v.(type) {
	case string:
	case int:
		n = uint64(v.(int))
		break
	case int16:
		n = v.(uint64)
		break
	default:
		return nil, errT()
	}
	return tobytes.Uint64ToBytes(n, e), nil
}

func interface2Float32(v interface{}, e tobytes.Endian) ([]byte, error) {
	n := float32(0)
	switch v.(type) {
	case string:
	case int:
		n = float32(v.(int))
		break
	case float32:
		n = v.(float32)
		break
	case float64:
		n = float32(v.(float64))
		break
	default:
		return nil, errT()
	}
	return tobytes.Float32ToBytes(n, e), nil
}

func interface2Float64(v interface{}, e tobytes.Endian) ([]byte, error) {
	n := float64(0)
	switch v.(type) {
	case string:
	case int:
		n = float64(v.(int))
		break
	case float32:
		n = float64(v.(float32))
		break
	case float64:
		n = v.(float64)
		break
	default:
		return nil, errT()
	}
	return tobytes.Float64ToBytes(n, e), nil
}

func interface2s(v interface{}) ([]byte, error) {
	return interface2Int16(v, tobytes.ThisEndian)
}

func interface2S(v interface{}) ([]byte, error) {
	return interface2Uint16(v, tobytes.ThisEndian)
}

func interface2n(v interface{}) ([]byte, error) {
	return interface2Uint16(v, tobytes.BigEndian)
}

func interface2v(v interface{}) ([]byte, error) {
	return interface2Uint16(v, tobytes.LittleEndian)
}

func interface2i(v interface{}) ([]byte, error) {
	return interface2Int(v, tobytes.ThisEndian)
}

func interface2I(v interface{}) ([]byte, error) {
	return interface2Uint(v, tobytes.ThisEndian)
}

func interface2l(v interface{}) ([]byte, error) {
	return interface2Int32(v, tobytes.ThisEndian)
}

func interface2L(v interface{}) ([]byte, error) {
	return interface2Uint32(v, tobytes.ThisEndian)
}

func interface2N(v interface{}) ([]byte, error) {
	return interface2Uint32(v, tobytes.BigEndian)
}

func interface2V(v interface{}) ([]byte, error) {
	return interface2Uint32(v, tobytes.LittleEndian)
}

func interface2q(v interface{}) ([]byte, error) {
	return interface2Int64(v, tobytes.ThisEndian)
}

func interface2Q(v interface{}) ([]byte, error) {
	return interface2Uint64(v, tobytes.ThisEndian)
}

func interface2J(v interface{}) ([]byte, error) {
	return interface2Uint64(v, tobytes.BigEndian)
}

func interface2P(v interface{}) ([]byte, error) {
	return interface2Uint64(v, tobytes.LittleEndian)
}

func interface2f(v interface{}) ([]byte, error) {
	return interface2Float32(v, tobytes.ThisEndian)
}

func interface2g(v interface{}) ([]byte, error) {
	return interface2Float32(v, tobytes.LittleEndian)
}

func interface2G(v interface{}) ([]byte, error) {
	return interface2Float32(v, tobytes.BigEndian)
}

func interface2d(v interface{}) ([]byte, error) {
	return interface2Float64(v, tobytes.ThisEndian)
}

func interface2e(v interface{}) ([]byte, error) {
	return interface2Float64(v, tobytes.LittleEndian)
}

func interface2E(v interface{}) ([]byte, error) {
	return interface2Float64(v, tobytes.BigEndian)
}

func x(l int) []byte {
	return make([]byte, l)
}

func X(b *[]byte, l int) {
	*b = (*b)[0 : len(*b)-l]
}

func at(b *[]byte, l int) {
	s := make([]byte, len(*b)-l)
	copy((*b)[l:], s)
}
