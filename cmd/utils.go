package wakehostapi


func Contains(hosts []Host, host Host) bool {
	for _, v := range hosts {
		if v == host {
			return true
		}
	}
	return false
}
func Compare(host1 Host, host Host) bool {

	if host1.Name == host.Name {
		return true
	}
	if host1.MacAddress == host.MacAddress {
		return true
	}
	if host1.IpAddress == host.IpAddress {
		return true
	}
	return false
}