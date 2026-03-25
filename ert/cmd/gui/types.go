package main

type SystemInfo struct {
	Hostname     string `json:"hostname"`
	OSName       string `json:"os_name"`
	OSVersion    string `json:"os_version"`
	Architecture string `json:"architecture"`
	BootTime     string `json:"boot_time"`
	CurrentUser  string `json:"current_user"`
	CPUCount     int    `json:"cpu_count"`
	MemoryTotal  uint64 `json:"memory_total"`
	DiskTotal    uint64 `json:"disk_total"`
	IsDomain     bool   `json:"is_domain"`
	DomainName   string `json:"domain_name"`
}
