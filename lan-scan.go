package main

import (
	"fmt"
	"net"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/go-ping/ping"
	"github.com/mostlygeek/arp"
)

type Host struct {
	Ip  []uint8
	Mac string
}

type Hosts []Host

func (h Hosts) Len() int {
	return len(h)
}

func (h Hosts) Swap(i, j int) {
	h[i], h[j] = h[j], h[i]
}

func (h Hosts) Less(i, j int) bool {
	ih := h[i].Ip
	jh := h[j].Ip
	for index := range ih {
		if ih[index] == jh[index] {
			continue
		}
		return ih[index] < jh[index]
	}
	return true
}

func (h Hosts) String() string {
	if len(h) > 0 {
		return fmt.Sprintf("%v", h[0])
	}
	return ""
}

func ArpTable() Hosts {
	hs := Hosts{}
	for ipString := range arp.Table() {
		ipParts := strings.Split(ipString, ".")
		ip := make([]uint8, len(ipParts))
		for i, part := range ipParts {
			num, _ := strconv.Atoi(part)
			ip[i] = uint8(num)
		}
		h := Host{
			Ip:  ip,
			Mac: arp.Search(ipString),
		}
		hs = append(hs, h)
	}
	sort.Sort(Hosts(hs))
	return hs
}

// main is the entry point of the program.
// It performs the following tasks:
// - Retrieves the ARP table using the ArpTable function.
// - Filters out hosts with MAC address "00:00:00:00:00:00".
// - Prints the filteredHost hosts.
// - Retrieves the IP addresses of the network interfaces.
// - Creates a channel to store IP addresses.
// - Spawns goroutines to iterate over IP ranges and send IP addresses to the channel.
// - Prints the IP addresses received from the channel.
// - Spawns worker goroutines to perform tasks on each IP address.
// - Waits for all worker goroutines to finish.
func main() {
	addrs, err := net.InterfaceAddrs()
	if err != nil {
		fmt.Println("Error getting IP addresses:", err)
		return
	}
	availableIps := make([]net.IP, 0)
	var wg sync.WaitGroup
	for _, addr := range addrs {
		ipNet, ok := addr.(*net.IPNet)
		if ok && !ipNet.IP.IsLoopback() && ipNet.IP.To4() != nil {
			ips := make(chan net.IP, 16)
			ipRange := NewIpRange(ipNet.IP.To4(), ipNet.Mask)
			go func() {
				for {
					ip, err := ipRange.Next()
					if err != nil {
						close(ips)
						return
					}
					wg.Add(1)
					ips <- ip
				}
			}()
			for ip := range ips {
				go func(ip net.IP) {
					if IpAvaiable(ip) {
						availableIps = append(availableIps, ip)
					}
					wg.Done()
				}(ip)
			}
		}
	}
	wg.Wait()
	hosts := ArpTable()
	length := hosts.Len()
	fmt.Printf("ip\t\tmac\n")
	fmt.Println("---------------------------------")
	for i := 0; i < length; i++ {
		h := hosts[i]
		// 过滤 虚拟网卡
		if h.Mac == "00:00:00:00:00:00" {
			continue
		}
		fmt.Printf("%v\t%v\n", net.IP(h.Ip), h.Mac)
	}
}

// IpAvaiable checks if the given IP address is available by sending a ping request.
// It returns true if the IP address is available, and false otherwise.
func IpAvaiable(ip net.IP) bool {
	// 创建一个新的ping器用于向指定IP发送ping请求
	pinger, err := ping.NewPinger(ip.String())
	if err != nil {
		// 如果创建ping器时发生错误，打印错误信息并返回false
		return false
	}
	// 设置ping器的参数
	// 将要发送的echo请求的数量设置为3
	pinger.Count = 1
	// 设置每次请求之间的间隔为1秒
	pinger.Interval = time.Second
	// 设置每个ping响应的超时时间为3秒
	pinger.Timeout = time.Second * 1

	// 执行ping操作
	err = pinger.Run()
	if err != nil {
		// 如果在ping过程中出现错误，返回false
		return false
	}
	stats := pinger.Statistics()
	return stats.PacketsRecv != 0
}
