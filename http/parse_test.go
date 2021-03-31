package http

import (
	"bytes"
	"testing"
)

type IPRegToIPs struct {
	name  string
	raw   string
	wants map[string]struct{}
}

func TestParseIP4(t *testing.T) {
	var ip4 = []IPRegToIPs{
		{
			"'-'功能:连续的IP地址",
			"192.168.10-13.0",
			map[string]struct{}{"192.168.10.0": struct{}{}, "192.168.11.0": struct{}{}, "192.168.12.0": struct{}{}, "192.168.13.0": struct{}{}},
		},
		{
			"'[]'功能:端点功能",
			"192.168.[10-13].0",
			map[string]struct{}{"192.168.10.0": struct{}{}, "192.168.13.0": struct{}{}},
		},
		{
			"'/'功能:CIDR表示法",
			"192.168.10.0/29",
			map[string]struct{}{"192.168.10.1": struct{}{}, "192.168.10.2": struct{}{}, "192.168.10.3": struct{}{}, "192.168.10.4": struct{}{}, "192.168.10.5": struct{}{}, "192.168.10.6": struct{}{}},
		},
		{
			"'/'功能:CIDR表示法",
			"192.168.10.8/29",
			map[string]struct{}{"192.168.10.9": struct{}{}, "192.168.10.10": struct{}{}, "192.168.10.11": struct{}{}, "192.168.10.12": struct{}{}, "192.168.10.13": struct{}{}, "192.168.10.14": struct{}{}},
		},
	}

	for _, val := range ip4 {
		t.Run(val.name, func(t *testing.T) {
			ips := ParseIP(val.raw)
			if len(ips) != len(val.wants) {
				t.Errorf("预期和实际解析得到的IP总数不一致: wants: %d, get: %d\n", len(val.wants), len(ips))
				return
			}
			for _, ip := range ips {
				if _, ok := val.wants[ip]; !ok {
					t.Errorf("预期和实际解析得到的IP不一致: wants: %v, get: %v\n", val.wants, ips)
				}
			}
		})
	}

}

func TestParseIP(t *testing.T) {
	var ips = []IPRegToIPs{
		{
			"','功能:解析多个IP表示式",
			"192.168.15.5-7,192.168.10-13.0",
			map[string]struct{}{"192.168.10.0": struct{}{}, "192.168.11.0": struct{}{}, "192.168.12.0": struct{}{}, "192.168.13.0": struct{}{}, "192.168.15.5": struct{}{}, "192.168.15.6": struct{}{}, "192.168.15.7": struct{}{}},
		},
		{
			"多个IP表达式结果去重",
			"192.168.10|13.0,192.168.10.0",
			map[string]struct{}{"192.168.10.0": struct{}{}, "192.168.13.0": struct{}{}},
		},
	}

	for _, val := range ips {
		t.Run(val.name, func(t *testing.T) {
			ips := ParseIP(val.raw)
			if len(ips) != len(val.wants) {
				t.Errorf("预期和实际解析得到的IP总数不一致: wants: %d, get: %d\n", len(val.wants), len(ips))
			}
			for _, ip := range ips {
				if _, ok := val.wants[ip]; !ok {
					t.Errorf("预期和实际解析得到的IP不一致: wants: %v, get: %v\n", val.wants, ips)
					return
				}
			}
		})
	}
}

func TestUint32ToBytes(t *testing.T) {
	var data = []struct {
		n    uint32
		want []byte
		get  []byte
	}{
		{0, []byte{0, 0, 0, 0}, []byte{}},
		{0x14, []byte{0, 0, 0, 0x14}, []byte{}},
		{0x2300, []byte{0, 0, 0x23, 0}, []byte{}},
		{0x120000, []byte{0, 0x12, 0, 0}, []byte{}},
		{0xff000000, []byte{0xff, 0, 0, 0}, []byte{}},
		{0xff102314, []byte{0xff, 0x10, 0x23, 0x14}, []byte{}},
	}

	for _, val := range data {
		val.get = uint32ToBytes(val.n)
		if !bytes.Equal(val.want, val.get) {
			t.Errorf("uint32数字转换成[]byte与预期不一致, %+v", val)
		}
	}
}

func TestBytesToUint32(t *testing.T) {
	var data = []struct {
		b    []byte
		want uint32
		get  uint32
	}{
		{[]byte{0, 0, 0, 0}, 0, 0},
		{[]byte{0, 0, 0, 0x14}, 0x14, 0},
		{[]byte{0, 0, 0x23, 0}, 0x2300, 0},
		{[]byte{0, 0x12, 0, 0}, 0x120000, 0},
		{[]byte{0xff, 0, 0, 0}, 0xff000000, 0},
		{[]byte{0xff, 0x10, 0x23, 0x14}, 0xff102314, 0},
	}

	for _, val := range data {
		val.get = bytesToUint32(val.b)
		if val.want != val.get {
			t.Errorf("[]byte转换成uint32数字与预期不一致, %+v", val)
		}
	}
}

func TestStrNumToByte(t *testing.T) {
	var data = []struct {
		str  string
		want byte
		get  byte
	}{
		{"0", 0, 0},
		{"10", 10, 0},
		{"192", 192, 0},
		{"255", 255, 0},
	}

	for _, val := range data {
		val.get = strNumToByte(val.str)
		if val.want != val.get {
			t.Errorf("[]byte转换成uint32数字与预期不一致, %+v", val)
		}
	}
}

func TestSearchIP4ByCIDR(t *testing.T) {
	var data = []struct {
		ip, maskstr string
		want        [][]byte
		get         [][]byte
	}{
		{"192.168.10.0", "29", [][]byte{[]byte{192, 168, 10, 1}, []byte{192, 168, 10, 2}, []byte{192, 168, 10, 3}, []byte{192, 168, 10, 4}, []byte{192, 168, 10, 5}, []byte{192, 168, 10, 6}}, [][]byte{}},
	}

	for _, val := range data {
		val.get = searchIP4ByCIDR(val.ip, val.maskstr)
		if !twoDimenByte(val.want).Equal(twoDimenByte(val.get)) {
			t.Errorf("[]byte转换成uint32数字与预期不一致, %+v", val)
		}
	}
}

func TestSearchIP4(t *testing.T) {
	var data = []struct {
		reg  string
		want [][]byte
		get  [][]byte
	}{
		{"192.168.[10-13].0", [][]byte{[]byte{192, 168, 10, 0}, []byte{192, 168, 13, 0}}, [][]byte{}},
		{"192.168.10-13.0", [][]byte{[]byte{192, 168, 10, 0}, []byte{192, 168, 11, 0}, []byte{192, 168, 12, 0}, []byte{192, 168, 13, 0}}, [][]byte{}},
	}
	for _, val := range data {
		val.get = searchIP4(val.reg)
		if !twoDimenByte(val.want).Equal(twoDimenByte(val.get)) {
			t.Errorf("[]byte转换成uint32数字与预期不一致, %+v", val)
		}
	}
}

type twoDimenByte [][]byte

func (t twoDimenByte) Equal(b twoDimenByte) bool {
	if len(t) != len(b) {
		return false
	}

	buff := make(map[string]struct{})
	for _, val := range t {
		buff[string(val)] = struct{}{}
	}
	for _, val := range b {
		if _, ok := buff[string(val)]; !ok {
			return false
		}
	}
	return true
}
