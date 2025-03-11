package common

import "fmt"

func MemoryText(m int) string {
	m = m / 1024
	if m < 1024 {
		return fmt.Sprintf("%d KB", m)
	} else if m < 1024*1024 {
		return fmt.Sprintf("%.2f MB", float64(m)/1024)
	} else if m < 1024*1024*1024 {
		return fmt.Sprintf("%.2f GB", float64(m)/(1024*1024))
	} else {
		return fmt.Sprintf("%.2f TB", float64(m)/(1024*1024*1024))
	}
}

func TrafficText(t int) string {
	t = t / 125
	if t < 1000 {
		return fmt.Sprintf("%d Kbps", t)
	} else if t < 1000*1000 {
		return fmt.Sprintf("%.2f Mbps", float64(t)/1000)
	} else {
		return fmt.Sprintf("%.2f Gbps", float64(t)/(1000*1000))
	}
}
