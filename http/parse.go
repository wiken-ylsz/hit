package http

import (
	"errors"
	"net"
	"strings"
)

// ErrInvaildIPAddr 无效的IP地址
var ErrInvaildIPAddr = errors.New("无效的IP地址")

// ParseIP 解析IP表达式
func ParseIP(reg string) (ips []string) {
	// 解析IP4地址
	for _, val := range strings.Split(reg, ",") {
		if strings.Contains(reg, ".") {
			ips = append(ips, parseIP4(val)...)
		}
	}

	// TODO: 解析IP6地址
	return
}

// parseIP4 解析IP4地址
// 1.支持CIDR表示法,解析'192.168.0.10/29'得到: 192.168.0.9、192.168.0.10、192.168.0.11、192.168.0.12、192.168.0.13、192.168.0.14
// 2.解析'192.168.0.10-13'得到: 192.168.0.10、 192.168.0.11、192.168.0.12、192.168.0.13
// 3.解析'192.168.0.[10-13]'得到: 192.168.0.10、192.168.0.13
func parseIP4(str string) (ips []string) {
	var _bytes [][]byte
	if i := strings.Index(str, "/"); i > 0 { // CIDR 表示法
		_bytes = searchIP4ByCIDR(str[0:i], str[i+1:])
	} else {
		_bytes = searchIP4(str)
	}
	for _, b := range _bytes {
		ips = append(ips, net.IP(b).String())
	}
	return
}

// MaxUnit32 32位最大无符号数
const MaxUnit32 uint32 = 1<<32 - 1

// searchIP4ByCIDR 解析CIDR表示法,返回可用的IP
func searchIP4ByCIDR(ip, maskstr string) (ips [][]byte) {
	buff := make([]byte, 0, 4)
	for _, val := range strings.Split(ip, ".") {
		buff = append(buff, strNumToByte(val))
	}
	if len(buff) != 4 {
		return
	}

	// 1.计算掩码
	mask := ^(MaxUnit32 >> strNumToByte(maskstr))

	// 2.计算网络号
	b := []byte{buff[0], buff[1], buff[2], buff[3]}
	netid := bytesToUint32(b) & mask

	// 3.计算第一可用IP
	min := netid + 1

	// 4.计算最后可用IP
	max := (netid | (MaxUnit32 >> strNumToByte(maskstr))) - 1

	// 5.计算可用IP
	for i := min; i <= max; i++ {
		ips = append(ips, uint32ToBytes(i))
	}
	return
}

// searchIP4 解析IP地址段, 返回可用的IP
func searchIP4(reg string) (ips [][]byte) {
	strs := strings.Split(reg, ".")
	if len(strs) != 4 {
		return
	}

	res := make([]uint32, 1)
	for _, val := range strs {
		leng := len(val)
		isEndPoint := false
		if val[0] == '[' {
			val = val[1:]
			leng--
			isEndPoint = true
		}
		if val[leng-1] == ']' {
			val = val[:leng-1]
			isEndPoint = true
		}

		buff := make([]uint32, 0)
		if i := strings.Index(val, "-"); i > 0 {
			min, max := strNumToByte(val[:i]), strNumToByte(val[i+1:])
			if min > max {
				min, max = max, min
			}
			b := make([]uint32, 0)
			switch {
			case isEndPoint:
				b = append(b, uint32(min), uint32(max))
			default:
				for i := min; i <= max; i++ {
					b = append(b, uint32(i))
				}
			}
			buff = append(buff, b...)
		} else {
			buff = append(buff, uint32(strNumToByte(val)))
		}

		_res := make([]uint32, 0, len(res)*len(buff))

		for i := 0; i < len(res); i++ {
			for j := 0; j < len(buff); j++ {
				_res = append(_res, (res[i]<<8)|buff[j])
			}
		}
		res = _res
	}

	for _, n := range res {
		ips = append(ips, uint32ToBytes(n))
	}

	return
}

func uint32ToBytes(n uint32) (b []byte) {
	b = make([]byte, 4)
	b[0] = byte((n >> 24))
	b[1] = byte((n << 8) >> 24)
	b[2] = byte((n << 16) >> 24)
	b[3] = byte((n << 24) >> 24)
	return
}

func bytesToUint32(b []byte) (n uint32) {
	for _, i := range b {
		n = (n << 8) | uint32(i)
	}
	return
}

func strNumToByte(str string) (b byte) {
	for _, s := range []byte(str) {
		switch s {
		case '0', '1', '2', '3', '4', '5', '6', '7', '8', '9':
			b = 10*b + (s - '0')
		}
	}
	return
}
