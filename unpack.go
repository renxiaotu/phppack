package phppack

import (
	"bytes"
	"encoding/hex"
	"errors"
	"github.com/renxiaotu/dtc/frombytes"
	"reflect"
	"strconv"
)

func UnpackByStruct(data interface{}, b []byte) error {
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
			return err
		}
		break
	default:
		return errors.New(PackageName + ":unsupported data type")
	}

	for i := 0; i < len(pts); i++ {
		pt := pts[i]
		v, err := unpack(&b, pt)
		if err != nil {
			return err
		}
		if v != nil {
			value.FieldByName(pts[i].Name).Set(reflect.ValueOf(v))
		}
	}
	return nil
}

func UnpackByFormat(f string, b []byte) (map[string]interface{}, error) {
	m := make(map[string]interface{})
	pts, err := parseUnPackFormats(f)
	if err != nil {
		return nil, err
	}

	index := 1
	for i := 0; i < len(pts); i++ {
		pt := pts[i]
		if pt.Name == "" {
			pt.Name = strconv.Itoa(index)
			index++
		}
		v, err := unpack(&b, pt)
		if err != nil {
			return nil, err
		}
		if v != nil {
			m[pt.Name] = v
		}
	}
	return m, nil
}

func unpack(b *[]byte, pt packType) (interface{}, error) {
	switch pt.tag.Type {
	//--------------------------------------------字符串--------------------------
	case "a": //以NUL字节填充字符串
		return un2a(b, pt)
	case "A": //以SPACE(空格)填充字符串
		return un2A(b, pt)

		//--------------------------------------------hex--------------------------
	case "h": //十六进制字符串，低位在前
		return un2h(b, pt)
	case "H": //十六进制字符串，高位在前
		return un2H(b, pt)

		//--------------------------------------------8bit--------------------------
	case "c": //有符号字符 int8
		return un2c(b)
	case "C": //无符号字符 uint8
		return un2C(b)

		//--------------------------------------------16bit--------------------------
	case "s": //有符号短整型(16位，主机字节序)
		return un2s(b)
	case "S": //无符号短整型(16位，主机字节序)
		return un2S(b)

	case "n": //无符号短整型(16位，大端字节序)
		return un2n(b)
	case "v": //无符号短整型(16位，小端字节序)
		return un2v(b)

		//--------------------------------------------this bit--------------------------
	case "i": //有符号整型(机器相关大小字节序)
		return un2i(b)
	case "I": //无符号整型(机器相关大小字节序)
		return un2I(b)

		//--------------------------------------------32bit--------------------------
	case "l": //有符号长整型(32位，主机字节序)
		return un2l(b)
	case "L": //无符号长整型(32位，主机字节序)
		return un2L(b)
	case "N": //无符号短整型(16位，大端字节序)
		return un2N(b)
	case "V": //无符号长整型(32位，小端字节序)
		return un2V(b)

		//--------------------------------------------64bit--------------------------
	case "q": //有符号长长整型(64位，主机字节序)
		return un2q(b)
	case "Q": //无符号长长整型(64位，主机字节序)
		return un2Q(b)

	case "J": //无符号长长整型(64位，大端字节序)
		return un2J(b)
	case "P": //无符号长长整型(64位，小端字节序)
		return un2P(b)

		//--------------------------------------------float--------------------------
	case "f": //单精度浮点型(主机字节序)
		return un2f(b)
	case "g": //单精度浮点型(小端字节序)
		return un2g(b)
	case "G": //单精度浮点型(大端字节序)
		return un2G(b)

	case "d": //双精度浮点型(主机字节序)
		return un2d(b)
	case "e": //双精度浮点型(小端字节序)
		return un2e(b)
	case "E": //双精度浮点型(大端字节序)
		return un2E(b)

		//--------------------------------------------other--------------------------
	case "x": //NUL字节
		return nil, un2x(b, pt)
	case "X": //回退字节
		return nil, nil
	case "Z": //a的别名
		return un2a(b, pt)
	case "@": //a的别名
		return nil, nil
	default: //不支持的格式
		return nil, errors.New("format contains characters that are not supported")
	}
}

func un2a(b *[]byte, pt packType) (string, error) {
	s := ""
	l := pt.tag.Size
	if l == -1 {
		l = len(*b)
	}
	if l == 0 {
		return s, nil
	}
	if len(*b) == 0 {
		return s, errNea
	}
	s = string((*b)[0:l])
	*b = (*b)[l:]
	s = string(bytes.TrimRight([]byte(s), string(byte(0))))
	return s, nil
}

func un2A(b *[]byte, pt packType) (string, error) {
	s := ""
	l := pt.tag.Size
	if l == -1 {
		l = len(*b)
	}
	if l == 0 {
		return s, nil
	}
	if len(*b) == 0 {
		return s, errNea
	}
	s = string((*b)[0:l])
	*b = (*b)[l:]
	s = string(bytes.TrimRight([]byte(s), " "))
	return s, nil
}

func un2h(b *[]byte, pt packType) (string, error) {
	s := ""
	l := pt.tag.Size
	if l == -1 {
		l = len(*b)
	}
	if l == 0 {
		return s, nil
	}
	if len(*b) == 0 {
		return s, errNea
	}

	if len(*b) == 0 {
		return s, errNea
	}
	l = l/2 + l%2
	vb := (*b)[:l]
	if l > len(vb) {
		vb = append(vb, byte(48))
	}
	v := []byte(hex.EncodeToString(vb))
	for l > 0 {
		tmp := v[l*2-1]
		v[l*2-1] = v[l*2-2]
		v[l*2-2] = tmp
		l--
	}
	s = string(v)
	*b = (*b)[l:]
	return s[:pt.tag.Size], nil
}

func un2H(b *[]byte, pt packType) (string, error) {
	s := ""
	l := pt.tag.Size
	if l == -1 {
		l = len(*b)
	}
	if l == 0 {
		return s, nil
	}
	if len(*b) == 0 {
		return s, errNea
	}

	if len(*b) == 0 {
		return s, errNea
	}
	l = l/2 + l%2
	v := (*b)[:l]
	if l > len(v) {
		v = append(v, byte(48))
	}
	s = hex.EncodeToString(v)
	*b = (*b)[l:]
	return s[:pt.tag.Size], nil
}

func un2c(b *[]byte) (int8, error) {
	bl := 1
	n := int8(0)
	if len(*b) < bl {
		return n, errNea
	}
	n = int8((*b)[0])
	*b = (*b)[bl:]
	return n, nil
}

func un2C(b *[]byte) (uint8, error) {
	bl := 1
	n := uint8(0)
	if len(*b) < bl {
		return n, errNea
	}
	n = (*b)[0]
	*b = (*b)[bl:]
	return n, nil
}

func un2Int16(b *[]byte, e frombytes.Endian) (int16, error) {
	bl := 2
	if len(*b) < bl {
		return 0, errNea
	}
	n := frombytes.BytesToInt16((*b)[0:bl], e)
	*b = (*b)[bl:]
	return n, nil
}

func un2Uint16(b *[]byte, e frombytes.Endian) (uint16, error) {
	bl := 2
	if len(*b) < bl {
		return 0, errNea
	}
	n := frombytes.BytesToUint16((*b)[0:bl], e)
	*b = (*b)[bl:]
	return n, nil
}

func un2Int(b *[]byte, e frombytes.Endian) (int, error) {
	bl := strconv.IntSize / 8
	if len(*b) < bl {
		return 0, errNea
	}
	n := frombytes.BytesToInt((*b)[0:bl], e)
	*b = (*b)[bl:]
	return n, nil
}

func un2Uint(b *[]byte, e frombytes.Endian) (uint, error) {
	bl := strconv.IntSize / 8
	if len(*b) < bl {
		return 0, errNea
	}
	n := frombytes.BytesToUint((*b)[0:bl], e)
	*b = (*b)[bl:]
	return n, nil
}

func un2Int32(b *[]byte, e frombytes.Endian) (int32, error) {
	bl := 4
	if len(*b) < bl {
		return 0, errNea
	}
	n := frombytes.BytesToInt32((*b)[0:bl], e)
	*b = (*b)[bl:]
	return n, nil
}

func un2Uint32(b *[]byte, e frombytes.Endian) (uint32, error) {
	bl := 4
	if len(*b) < bl {
		return 0, errNea
	}
	n := frombytes.BytesToUint32((*b)[0:bl], e)
	*b = (*b)[bl:]
	return n, nil
}

func un2Int64(b *[]byte, e frombytes.Endian) (int64, error) {
	bl := 8
	if len(*b) < bl {
		return 0, errNea
	}
	n := frombytes.BytesToInt64((*b)[0:bl], e)
	*b = (*b)[bl:]
	return n, nil
}

func un2Uint64(b *[]byte, e frombytes.Endian) (uint64, error) {
	bl := 8
	if len(*b) < bl {
		return 0, errNea
	}
	n := frombytes.BytesToUint64((*b)[0:bl], e)
	*b = (*b)[bl:]
	return n, nil
}

func un2Float32(b *[]byte, e frombytes.Endian) (float32, error) {
	bl := 2
	if len(*b) < bl {
		return 0, errNea
	}
	n := frombytes.BytesToFloat32((*b)[0:bl], e)
	*b = (*b)[bl:]
	return n, nil
}

func un2Float64(b *[]byte, e frombytes.Endian) (float64, error) {
	bl := 2
	if len(*b) < bl {
		return 0, errNea
	}
	n := frombytes.BytesToFloat64((*b)[0:bl], e)
	*b = (*b)[bl:]
	return n, nil
}

func un2s(b *[]byte) (int16, error) {
	return un2Int16(b, frombytes.ThisEndian)
}

func un2S(b *[]byte) (uint16, error) {
	return un2Uint16(b, frombytes.ThisEndian)
}

func un2n(b *[]byte) (uint16, error) {
	return un2Uint16(b, frombytes.BigEndian)
}

func un2v(b *[]byte) (uint16, error) {
	return un2Uint16(b, frombytes.LittleEndian)
}

func un2i(b *[]byte) (int, error) {
	return un2Int(b, frombytes.ThisEndian)
}

func un2I(b *[]byte) (uint, error) {
	return un2Uint(b, frombytes.ThisEndian)
}

func un2l(b *[]byte) (int32, error) {
	return un2Int32(b, frombytes.ThisEndian)
}

func un2L(b *[]byte) (uint32, error) {
	return un2Uint32(b, frombytes.ThisEndian)
}

func un2N(b *[]byte) (uint32, error) {
	return un2Uint32(b, frombytes.BigEndian)
}

func un2V(b *[]byte) (uint32, error) {
	return un2Uint32(b, frombytes.LittleEndian)
}

func un2q(b *[]byte) (int64, error) {
	return un2Int64(b, frombytes.ThisEndian)
}

func un2Q(b *[]byte) (uint64, error) {
	return un2Uint64(b, frombytes.ThisEndian)
}

func un2J(b *[]byte) (uint64, error) {
	return un2Uint64(b, frombytes.BigEndian)
}

func un2P(b *[]byte) (uint64, error) {
	return un2Uint64(b, frombytes.LittleEndian)
}

func un2f(b *[]byte) (float32, error) {
	return un2Float32(b, frombytes.ThisEndian)
}

func un2g(b *[]byte) (float32, error) {
	return un2Float32(b, frombytes.LittleEndian)
}

func un2G(b *[]byte) (float32, error) {
	return un2Float32(b, frombytes.BigEndian)
}

func un2d(b *[]byte) (float64, error) {
	return un2Float64(b, frombytes.ThisEndian)
}

func un2e(b *[]byte) (float64, error) {
	return un2Float64(b, frombytes.LittleEndian)
}

func un2E(b *[]byte) (float64, error) {
	return un2Float64(b, frombytes.BigEndian)
}

func un2x(b *[]byte, pt packType) error {
	l := pt.tag.Size
	if l == -1 {
		l = len(*b)
	}
	if l > len(*b) {
		return errNea
	}
	if l > 0 {
		*b = (*b)[l:]
	}
	return nil
}
