package main
import (
	"net"
	"math/big"
	"fmt"
)
type IPRange struct {
	start net.IP
	end net.IP
	startBig *big.Int
	endBig *big.Int
}

// new 创建一个IP范围对象。
//
// 参数:
//   ip []byte: 表示IP地址的字节切片，长度应为4。
//   mask []byte: 表示子网掩码的字节切片，长度应为4。
//
// 返回值:
//   *IPRange: 指向新创建的IP范围对象的指针。
//   error: 如果IP地址或子网掩码无效，则返回错误对象。
//
func NewIpRange(ip []byte, mask []byte) *IPRange {
	// 计算起始IP地址。
	start := append([]byte{}, ip...)
	for i := 0; i < 4; i++ {
		start[i] = ip[i] & mask[i] // 通过与操作计算起始IP。
		mask[i] = mask[i] ^ 0xff   // 更新mask，为计算结束IP做准备。
	}

	// 计算结束IP地址。
	end := append([]byte{}, ip...)
	for i := 0; i < 4; i++ {
		end[i] = ip[i] | mask[i] // 通过或操作计算结束IP。
	}

	// 返回IP范围对象。
	return &IPRange{
		start: start,
		end: end,
		startBig: big.NewInt(0).SetBytes([]byte(start)),
		endBig: big.NewInt(0).SetBytes([]byte(end)),
	}
}


// Next 返回范围内的下一个IP地址。
//
// 返回值:
//   [4]byte: 表示IP地址的字节切片。
//   error: 如果没有更多的IP地址，则返回错误对象。
func (r *IPRange) Next() (net.IP, error) {
	r.startBig.Add(r.startBig, big.NewInt(1)) // 计算下一个IP地址。
	startBig := r.startBig
	endBig := r.endBig
	if startBig.Cmp(endBig) > 0 {
		return nil, fmt.Errorf("no more IP addresses")
	}
	return net.IP(r.startBig.Bytes()), nil // 返回下一个IP地址。
}

func (r *IPRange) Reset() {
	r.startBig = big.NewInt(0).SetBytes([]byte(r.start))
}