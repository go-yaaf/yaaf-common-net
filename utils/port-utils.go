package utils

import (
	"sync"

	"github.com/go-yaaf/yaaf-common/entity"
)

type portDesc entity.Tuple[string, string]

var portUtilsOnce sync.Once
var portUtilsInst *PortUtilsStruct = nil

type PortUtilsStruct struct {
	portsMap map[int]portDesc
}

// PortUtils is a factory method that acts as a static member
func PortUtils() *PortUtilsStruct {
	portUtilsOnce.Do(func() {
		portUtilsInst = &PortUtilsStruct{portsMap: make(map[int]portDesc)}
		portUtilsInst.initWellKnownPorts()
	})
	return portUtilsInst
}

// GetName get well-known port name
func (t *PortUtilsStruct) GetName(port int) string {
	if desc, ok := t.portsMap[port]; ok {
		return desc.Key
	} else {
		return ""
	}
}

// GetDescription get well-known port description
func (t *PortUtilsStruct) GetDescription(port int) string {
	if desc, ok := t.portsMap[port]; ok {
		return desc.Value
	} else {
		return ""
	}
}

func (t *PortUtilsStruct) initWellKnownPorts() {

	t.portsMap[20] = portDesc{Key: "FTP", Value: "FTP Data Transfer"}                                   // TCP
	t.portsMap[21] = portDesc{Key: "FTPC", Value: "FTP Control"}                                        // TCP
	t.portsMap[22] = portDesc{Key: "SSH", Value: "Remote Shell Protocol"}                               // TCP/UDP
	t.portsMap[23] = portDesc{Key: "TELNET", Value: "Telnet"}                                           // TCP
	t.portsMap[24] = portDesc{Key: "LMTP", Value: "Local Mail Transfer Protocol"}                       // TCP
	t.portsMap[25] = portDesc{Key: "SMTP", Value: "SMTP Email Routing"}                                 // TCP
	t.portsMap[37] = portDesc{Key: "TIME", Value: "Time Protocol"}                                      // TCP/UDP
	t.portsMap[43] = portDesc{Key: "WHOIS", Value: "WHOIS Protocol"}                                    // TCP
	t.portsMap[53] = portDesc{Key: "DNS", Value: "DNS Service"}                                         // TCP/UDP
	t.portsMap[67] = portDesc{Key: "DHCP", Value: "DHCP Server"}                                        // UDP
	t.portsMap[68] = portDesc{Key: "DHCP", Value: "DHCP Client"}                                        // UDP
	t.portsMap[69] = portDesc{Key: "TFTP", Value: "Trivial File Transfer"}                              // UDP
	t.portsMap[70] = portDesc{Key: "GOPHER", Value: "Gopher protocol"}                                  // TCP
	t.portsMap[79] = portDesc{Key: "FINGER", Value: "Finger protocol"}                                  // TCP
	t.portsMap[80] = portDesc{Key: "HTTP", Value: "HTTP Web Traffic"}                                   // TCP
	t.portsMap[88] = portDesc{Key: "KERBEROS", Value: "Kerberos authentication"}                        // TCP/UDP
	t.portsMap[95] = portDesc{Key: "SUPDUP", Value: "Terminal remote login"}                            // TCP
	t.portsMap[101] = portDesc{Key: "NIC", Value: "NIC host name"}                                      // TCP
	t.portsMap[102] = portDesc{Key: "TSAP", Value: "Service access point"}                              // TCP
	t.portsMap[104] = portDesc{Key: "DICOM", Value: "Digital Imagine Communication"}                    // TCP/UDP
	t.portsMap[107] = portDesc{Key: "RTELNET", Value: "Remote User Telnet service"}                     // TCP/UDP
	t.portsMap[108] = portDesc{Key: "SNA", Value: "IBM SNA gateway"}                                    // TCP/UDP
	t.portsMap[109] = portDesc{Key: "POP2", Value: "Post Office protocol V2"}                           // TCP
	t.portsMap[110] = portDesc{Key: "POP3", Value: "Post Office protocol V3"}                           // TCP
	t.portsMap[111] = portDesc{Key: "ONC RPC", Value: "Sun RPC"}                                        // TCP/UDP
	t.portsMap[115] = portDesc{Key: "SFTP", Value: "Simple File Transfer Protocol"}                     // TCP
	t.portsMap[119] = portDesc{Key: "NNTP", Value: "NNTP Usenet"}                                       // TCP
	t.portsMap[123] = portDesc{Key: "NTP", Value: "NTP Time Synchronization"}                           // UDP
	t.portsMap[135] = portDesc{Key: "DCE/RPC", Value: "Microsoft RPC"}                                  // TCP/UDP
	t.portsMap[137] = portDesc{Key: "NetBIOS", Value: "NetBIOS Name Service"}                           // UDP
	t.portsMap[138] = portDesc{Key: "NetBIOS", Value: "NetBIOS Datagram Service"}                       // UDP
	t.portsMap[139] = portDesc{Key: "NetBIOS", Value: "NetBIOS Session Service"}                        // TCP
	t.portsMap[143] = portDesc{Key: "IMAP", Value: "IMAP Email Retrieval"}                              // TCP
	t.portsMap[161] = portDesc{Key: "SNMP", Value: "Simple Network Management Protocol"}                // UDP
	t.portsMap[162] = portDesc{Key: "SNMP Trap", Value: "Simple Network Management Protocol"}           // UDP
	t.portsMap[179] = portDesc{Key: "BGP", Value: "BGP Routing Protocol"}                               // TCP
	t.portsMap[389] = portDesc{Key: "LDAP", Value: "LDAP Directory Services"}                           // TCP/UDP
	t.portsMap[443] = portDesc{Key: "HTTPS", Value: "HTTPS Secure Web Traffic"}                         // TCP
	t.portsMap[445] = portDesc{Key: "SMB", Value: "Microsoft SMB File Sharing"}                         // TCP
	t.portsMap[465] = portDesc{Key: "SMTP TLS", Value: "SMTP Submission over TLS"}                      // TCP
	t.portsMap[500] = portDesc{Key: "ISAKMP", Value: "ISAKMP / IKE VPN Key Exchange"}                   // UDP
	t.portsMap[514] = portDesc{Key: "SYSLOG", Value: "Syslog"}                                          // UDP
	t.portsMap[520] = portDesc{Key: "RIP", Value: "RIP Routing Protocol"}                               // UDP
	t.portsMap[587] = portDesc{Key: "SMTP Submission", Value: "SMTP Submission"}                        // TCP
	t.portsMap[601] = portDesc{Key: "RSYSLOG", Value: "Reliable Syslog Service"}                        // TCP
	t.portsMap[631] = portDesc{Key: "IPP", Value: "Internet Printing Protocol"}                         // TCP
	t.portsMap[636] = portDesc{Key: "LDAPS", Value: "Secure LDAP"}                                      // TCP
	t.portsMap[684] = portDesc{Key: "CORBA", Value: "CORBA IIOP SSL"}                                   // TCP/UDP
	t.portsMap[802] = portDesc{Key: "MODBUS", Value: "MODBUS TCP Security"}                             // TCP/UDP
	t.portsMap[829] = portDesc{Key: "CMP", Value: "Certificate Management Protocol"}                    // TCP
	t.portsMap[853] = portDesc{Key: "DNSSEC", Value: "DNS over TLS"}                                    // TCP/UDP
	t.portsMap[873] = portDesc{Key: "RSYNC", Value: "file synchronization protocol"}                    // TCP
	t.portsMap[989] = portDesc{Key: "FTPS", Value: "FTP data over TLS/SSL"}                             // TP/UDP
	t.portsMap[990] = portDesc{Key: "FTPS", Value: "FTP control over TLS/SSL"}                          // TP/UDP
	t.portsMap[991] = portDesc{Key: "NAS", Value: "Netnews Administration System"}                      // TP/UDP
	t.portsMap[992] = portDesc{Key: "TELNETS", Value: "Telnet over TLS/SSL"}                            // TP/UDP
	t.portsMap[993] = portDesc{Key: "IMAPS", Value: "IMAP over TLS/SSL"}                                // TCP/UDP
	t.portsMap[995] = portDesc{Key: "POP3S", Value: "POP3 over TLS/SSL"}                                // TCP/UDP
	t.portsMap[1194] = portDesc{Key: "OpenVPN", Value: "Open VPN"}                                      // TCP/UDP
	t.portsMap[1293] = portDesc{Key: "IPSEC", Value: "Internet Protocol Security"}                      // TCP/UDP
	t.portsMap[1433] = portDesc{Key: "MSSQL", Value: "Microsoft SQL Server"}                            // TCP/UDP
	t.portsMap[1434] = portDesc{Key: "MSSQLMon", Value: "Microsoft SQL Monitor"}                        // TCP/UDP
	t.portsMap[1512] = portDesc{Key: "WINS", Value: "Microsoft Windows Internet Name Service"}          // TCP/UDP
	t.portsMap[1521] = portDesc{Key: "ORACLEDB", Value: "Oracle Database Listener"}                     // TCP
	t.portsMap[1527] = portDesc{Key: "SQL*Net", Value: "Oracle Net Services"}                           // TCP/UDP
	t.portsMap[1589] = portDesc{Key: "VQP", Value: "Cisco VLAN Query protocol"}                         // TCP/UDP
	t.portsMap[1701] = portDesc{Key: "L2F", Value: "Layer 2 Forwarding protocol"}                       // TCP/UDP
	t.portsMap[1723] = portDesc{Key: "PPTP", Value: "PTP Tunneling"}                                    // TCP
	t.portsMap[1891] = portDesc{Key: "MSMQ", Value: "Microsoft Message Queue"}                          // TCP/UDP
	t.portsMap[1812] = portDesc{Key: "RADIUS Auth", Value: "RADIUS Authentication"}                     // UDP
	t.portsMap[1813] = portDesc{Key: "RADIUS Acc", Value: "RADIUS Accounting"}                          // UDP
	t.portsMap[1883] = portDesc{Key: "MQTT", Value: "MQ Telemetry Transport"}                           // TCP/UDP
	t.portsMap[1900] = portDesc{Key: "SSDP", Value: "SSDP UPnP Discovery"}                              // UDP
	t.portsMap[1998] = portDesc{Key: "XOT", Value: "Cisco X.25 over TCP"}                               // TCP/UDP
	t.portsMap[2000] = portDesc{Key: "SCCP", Value: "Cisco Skinny Client Control Protocol "}            // TCP/UDP
	t.portsMap[2049] = portDesc{Key: "NFS", Value: "Network File System"}                               // TCP/UDP
	t.portsMap[2083] = portDesc{Key: "RADSEC", Value: "Secure RADIUS Service"}                          // TCP/UDP
	t.portsMap[2181] = portDesc{Key: "ZooKeeper", Value: "Apache ZooKeeper"}                            // TCP
	t.portsMap[2375] = portDesc{Key: "DOCKER", Value: "Docker Remote API (Unsecured)"}                  // TCP
	t.portsMap[2375] = portDesc{Key: "DOCKERS", Value: "Docker Remote API (TLS)"}                       // TCP
	t.portsMap[2761] = portDesc{Key: "ISCL", Value: "DICOM over Integrated Secure Communication Layer"} // TCP/UDP
	t.portsMap[2762] = portDesc{Key: "DICOMTLS", Value: "DICOM over TLS"}                               // TCP/UDP
	t.portsMap[2775] = portDesc{Key: "SMPP", Value: "Short Message P2P"}                                // TCP/UDP
	t.portsMap[3268] = portDesc{Key: "GCLDAP", Value: "Global Catalog LDAP"}                            // TCP
	t.portsMap[3306] = portDesc{Key: "MySQL", Value: "MySQL Database"}                                  // TCP
	t.portsMap[3389] = portDesc{Key: "RDP", Value: "Microsoft Remote Desktop"}                          // TCP
	t.portsMap[3478] = portDesc{Key: "STUN", Value: "STUN / TURN (WebRTC)"}                             // UDP
	t.portsMap[3868] = portDesc{Key: "DIAMETER", Value: "Diameter base protocol"}                       // TCP
	t.portsMap[4195] = portDesc{Key: "AWS", Value: "AWS protocol for cloud remoting"}                   // TCP/ UDP
	t.portsMap[4460] = portDesc{Key: "NTS", Value: "Network Time Security Key Establishment"}           // TCP
	t.portsMap[4486] = portDesc{Key: "ICMS", Value: "Integrated Client Message Service"}                // TCP/UDP
	t.portsMap[4500] = portDesc{Key: "IPSec NAT-T", Value: "IPSec NAT-Traversal"}                       // UDP
	t.portsMap[5000] = portDesc{Key: "UPnP", Value: "UPnP / Flask / Docker"}                            // TCP
	t.portsMap[5004] = portDesc{Key: "RTP", Value: "Real-time Transport media data"}                    // TCP/UDP
	t.portsMap[5005] = portDesc{Key: "RTCP", Value: "Real-time Transport media control"}                // TCP/UDP
	t.portsMap[5060] = portDesc{Key: "SIP", Value: "SIP VoIP Signaling"}                                // TCP/UDP
	t.portsMap[5269] = portDesc{Key: "XMPP", Value: "Extensible Messaging and Presence Protocol"}       // TCP
	t.portsMap[5280] = portDesc{Key: "XMPP", Value: "Extensible Messaging and Presence Protocol"}       // TCP
	t.portsMap[5281] = portDesc{Key: "XMPP", Value: "Extensible Messaging and Presence Protocol"}       // TCP
	t.portsMap[5298] = portDesc{Key: "XMPP", Value: "Extensible Messaging and Presence Protocol"}       // TCP/UDP
	t.portsMap[5349] = portDesc{Key: "STUN", Value: "STUN/TURN over TLS"}                               // TCP/UDP
	t.portsMap[5353] = portDesc{Key: "mDNS", Value: "Multicast DNS / Bonjour"}                          // UDP
	t.portsMap[5402] = portDesc{Key: "MFTP", Value: "Multicast File Transfer Protocol"}                 // TCP/UDP
	t.portsMap[5432] = portDesc{Key: "PGSQL", Value: "PostgreSQL Database"}                             // TCP
	t.portsMap[5670] = portDesc{Key: "FILEMQ", Value: "ZeroMQ File Message Queuing"}                    // TCP
	t.portsMap[5671] = portDesc{Key: "AMQPS", Value: "AMQP over TLS"}                                   // TCP
	t.portsMap[5672] = portDesc{Key: "AMQP", Value: "RabbitMQ / AMQP"}                                  // TCP
	t.portsMap[5900] = portDesc{Key: "VNC", Value: "VNC Remote Desktop"}                                // TCP/UDP
	t.portsMap[6343] = portDesc{Key: "SFLOW", Value: "sFlow traffic monitoring"}                        // UDP
	t.portsMap[6379] = portDesc{Key: "REDIS", Value: "Redis Database"}                                  // TCP
	t.portsMap[6443] = portDesc{Key: "Kubernetes", Value: "Kubernetes API Server"}                      // TCP
	t.portsMap[6514] = portDesc{Key: "SYSLOG-TLS", Value: "Syslog over TLS"}                            // TCP/UDP
	t.portsMap[6622] = portDesc{Key: "MFTP", Value: "Multicast File Transfer Protocol"}                 // TCP/UDP
	t.portsMap[6653] = portDesc{Key: "OpenFlow", Value: "OpenFlow"}                                     // TCP
	t.portsMap[6697] = portDesc{Key: "IRC-SSL", Value: "IRC over SSL"}                                  // TCP
	t.portsMap[6881] = portDesc{Key: "BitTorrent", Value: "BitTorrent"}                                 // TCP/UDP
	t.portsMap[6882] = portDesc{Key: "BitTorrent", Value: "BitTorrent"}                                 // TCP/UDP
	t.portsMap[6883] = portDesc{Key: "BitTorrent", Value: "BitTorrent"}                                 // TCP/UDP
	t.portsMap[6884] = portDesc{Key: "BitTorrent", Value: "BitTorrent"}                                 // TCP/UDP
	t.portsMap[6885] = portDesc{Key: "BitTorrent", Value: "BitTorrent"}                                 // TCP/UDP
	t.portsMap[6886] = portDesc{Key: "BitTorrent", Value: "BitTorrent"}                                 // TCP/UDP
	t.portsMap[6887] = portDesc{Key: "BitTorrent", Value: "BitTorrent"}                                 // TCP/UDP
	t.portsMap[6888] = portDesc{Key: "BitTorrent", Value: "BitTorrent"}                                 // TCP/UDP
	t.portsMap[6889] = portDesc{Key: "BitTorrent", Value: "BitTorrent"}                                 // TCP/UDP
	t.portsMap[6890] = portDesc{Key: "BitTorrent", Value: "BitTorrent"}                                 // TCP/UDP
	t.portsMap[7400] = portDesc{Key: "RTPS", Value: "RTPS DDS Discovery"}                               // TCP/UDP
	t.portsMap[7401] = portDesc{Key: "RTPS", Value: "RTPS DDS User Traffic"}                            // TCP/UDP
	t.portsMap[7402] = portDesc{Key: "RTPS", Value: "RTPS DDS Meta Traffic"}                            // TCP/UDP
	t.portsMap[8008] = portDesc{Key: "HTTP", Value: "HTTP Alternative port"}                            // TCP
	t.portsMap[8080] = portDesc{Key: "HTTP", Value: "HTTP Alternative port"}                            // TCP
	t.portsMap[8443] = portDesc{Key: "HTTPS", Value: "HTTPS Alternative port"}                          // TCP
	t.portsMap[8883] = portDesc{Key: "SMQTT", Value: "MQTT over TLS"}                                   // TCP/UDP
	t.portsMap[8888] = portDesc{Key: "WebRTC", Value: "WebRTC"}                                         // TCP
	t.portsMap[8889] = portDesc{Key: "WebRTC", Value: "WebRTC"}                                         // TCP
	t.portsMap[9150] = portDesc{Key: "TOR", Value: "Tor Browser"}                                       // TCP
	t.portsMap[9092] = portDesc{Key: "KASKA", Value: "Apache Kafka"}                                    // TCP
	t.portsMap[9200] = portDesc{Key: "Elasticsearch", Value: "Elasticsearch REST API"}                  // TCP
	t.portsMap[9899] = portDesc{Key: "SCTP", Value: "SCTP tunneling"}                                   // UDP
	t.portsMap[10443] = portDesc{Key: "SSLVPN", Value: "Secured VPN access over SSL"}                   // TCP
	t.portsMap[11211] = portDesc{Key: "Memcached", Value: "Memcached"}                                  // TCP
	t.portsMap[11214] = portDesc{Key: "Memcached", Value: "Memcached Incoming SSL proxy"}               // TCP
	t.portsMap[11215] = portDesc{Key: "Memcached", Value: "Memcached Outgoing SSL proxy"}               // TCP
}
