package ofswitch

import (
	"encoding/binary"
	"fmt"
	"math"
	"net"
	"strconv"
	"sync"

	"github.com/golang-collections/go-datastructures/bitarray"
	"github.com/vishvananda/netlink"
)

type IP4AddrPool struct {
	subnet         *net.IPNet
	subnetLength   int
	hostCount      uint64
	allocationPool bitarray.BitArray
	mu             sync.Mutex
}

func NewIP4AddrPool(subnet *netlink.Addr) *IP4AddrPool {
	subnetLength, _ := subnet.IPNet.Mask.Size()
	hostLength := 32 - subnetLength
	hostCount := uint64(math.Pow(2, float64(hostLength)))
	pool := &IP4AddrPool{
		subnet:         subnet.IPNet,
		subnetLength:   subnetLength,
		hostCount:      hostCount,
		allocationPool: bitarray.NewBitArray(hostCount),
		mu:             sync.Mutex{},
	}
	pool.allocationPool.SetBit(0)
	pool.allocationPool.SetBit(uint64(hostCount - 1))

	return pool
}

func (p *IP4AddrPool) LeaseWithAddr(addr *netlink.Addr) error {
	p.mu.Lock()
	defer p.mu.Unlock()

	if !p.subnet.Contains(addr.IP) {
		return fmt.Errorf("IP %v is out of range %v", addr.IP.String(), p.subnet.String())
	}

	hostAddr := uint64(getHostAddr(addr.IP, p.subnetLength))
	isSet, err := p.allocationPool.GetBit(hostAddr)
	if err != nil {
		return err
	}
	if isSet {
		return fmt.Errorf("IP %v is already assigned", addr.IP.String())
	}

	return p.allocationPool.SetBit(hostAddr)
}

func (p *IP4AddrPool) Lease() (*netlink.Addr, error) {
	p.mu.Lock()
	defer p.mu.Unlock()

	addr := uint32(0)
	for i := uint32(0); i < uint32(p.hostCount); i++ {
		isSet, err := p.allocationPool.GetBit(uint64(i))
		if err != nil {
			return nil, err
		}
		if isSet {
			continue
		}
		subnetAddr := getSubnetAddr(p.subnet.IP, p.subnetLength)
		addr = subnetAddr + i
		p.allocationPool.SetBit(uint64(i))
		break
	}

	if addr == 0 {
		return nil, fmt.Errorf("no available address")
	}

	ipByte := make([]byte, 4)
	binary.BigEndian.PutUint32(ipByte, addr)
	ip := net.IP(ipByte)
	ipString := ip.To4().String() + "/" + strconv.Itoa(p.subnetLength)
	ipAddr, err := netlink.ParseAddr(ipString)
	return ipAddr, err
}

func (p *IP4AddrPool) Release(addr net.IP) error {
	p.mu.Lock()
	defer p.mu.Unlock()

	if !p.subnet.Contains(addr) {
		return fmt.Errorf("IP %v is out of range %v", addr.String(), p.subnet.String())
	}

	hostAddr := uint64(getHostAddr(addr, p.subnetLength))
	isSet, err := p.allocationPool.GetBit(hostAddr)
	if err != nil {
		return err
	}
	if !isSet {
		return fmt.Errorf("IP %v is not assigned", addr.String())
	}

	return p.allocationPool.ClearBit(hostAddr)
}

func getHostAddr(addr net.IP, subnetLength int) uint32 {
	ipUint32 := binary.BigEndian.Uint32(addr.To4())
	hostMask := uint32(math.Pow(2, float64(32-subnetLength)) - 1)

	return hostMask & ipUint32
}

func getSubnetAddr(addr net.IP, subnetLength int) uint32 {
	ipUint32 := binary.BigEndian.Uint32(addr.To4())
	hostLength := uint32(32 - subnetLength)
	return (ipUint32 >> hostLength) << hostLength
}
