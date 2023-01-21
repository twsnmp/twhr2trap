package main

import (
	"flag"
	"fmt"
	"log"
	"net"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/gosnmp/gosnmp"
	"github.com/shirou/gopsutil/v3/cpu"
	"github.com/shirou/gopsutil/v3/disk"
	"github.com/shirou/gopsutil/v3/load"
	"github.com/shirou/gopsutil/v3/mem"
)

var version = "v1.0.0"
var commit = ""
var trap = ""
var community = ""
var user = ""
var passwd = ""
var engineID = ""
var snmpMode = "v2c"
var syslog = ""
var interval = 10
var cpuTh = 0
var memTh = 0
var diskTh = 0
var loadTh = 0
var all = false

func init() {
	flag.StringVar(&trap, "trap", "", "trap destnation list")
	flag.StringVar(&snmpMode, "mode", "v2c", "snmp trap mode (v2c|v3Auth|v3AuthPriv)")
	flag.StringVar(&community, "community", "", "snmp v2c  trap community")
	flag.StringVar(&user, "user", "", "snmp v3 user")
	flag.StringVar(&passwd, "password", "", "snmp v3 password")
	flag.StringVar(&engineID, "eid", "", "snmp v3 engine ID")
	flag.StringVar(&syslog, "syslog", "", "syslog destnation list")
	flag.IntVar(&interval, "interval", 10, "check interval(sec)")
	flag.IntVar(&cpuTh, "cpu", 0, "cpu usage threshold 0=disable")
	flag.IntVar(&memTh, "mem", 0, "memory usage threshold 0=disable")
	flag.IntVar(&diskTh, "disk", 0, "disk usage threshold 0=disable")
	flag.IntVar(&loadTh, "load", 0, "load usage threshold 0=disable")
	flag.BoolVar(&all, "all", false, "send trap continuously")
	flag.VisitAll(func(f *flag.Flag) {
		if s := os.Getenv("TWHR2TRAP_" + strings.ToUpper(f.Name)); s != "" {
			f.Value.Set(s)
		}
	})
	flag.Parse()
}

type logWriter struct {
}

func (writer logWriter) Write(bytes []byte) (int, error) {
	return fmt.Print(time.Now().Format("2006-01-02T15:04:05.999 ") + string(bytes))
}

func main() {
	log.SetFlags(0)
	log.SetOutput(new(logWriter))
	log.Printf("version=%s", fmt.Sprintf("%s(%s)", version, commit))
	if trap == "" && syslog == "" {
		log.Fatalln("no trap or syslog distenation")
	}
	if engineID == "" {
		engineID = fmt.Sprintf("17861:%d", time.Now().Unix())
	}
	snmpMode = strings.ToLower(snmpMode)
	for {
		checkHostResource()
		time.Sleep(time.Second * time.Duration(interval))
	}
}

func checkHostResource() {
	checkCPU()
	checkMemory()
	checkLoad()
	checkDisk()
}

var sentCPU = false

func checkCPU() {
	if cpuTh == 0 {
		return
	}
	cpus, err := cpu.Percent(0, false)
	if err != nil {
		log.Printf("checkCPU err=%v", err)
		return
	}
	intCPU := int(cpus[0])
	if cpuTh < intCPU {
		if all || !sentCPU {
			send("cpu", cpuTh, intCPU, false)
			sentCPU = true
		}
	} else if sentCPU {
		send("cpu", cpuTh, intCPU, true)
		sentCPU = false
	}
}

var sentMem = false

func checkMemory() {
	if memTh == 0 {
		return
	}
	mems, err := mem.VirtualMemory()
	if err != nil {
		log.Printf("checkMemory err=%v", err)
		return
	}
	intMem := int(mems.UsedPercent)
	if memTh < intMem {
		if !sentMem || all {
			send("mem", memTh, intMem, false)
			sentMem = true
		}
	} else if sentMem {
		send("mem", memTh, intMem, true)
		sentMem = false
	}
}

var sentLoad = false

func checkLoad() {
	if loadTh == 0 {
		return
	}
	loads, err := load.Avg()
	if err != nil {
		log.Printf("checkLoad err=%v", err)
		return
	}
	intLoad := int(loads.Load1)
	if loadTh < intLoad {
		if !sentLoad || all {
			send("load", loadTh, intLoad, false)
			sentLoad = true
		}
	} else if sentLoad {
		send("load", loadTh, intLoad, true)
		sentLoad = false
	}
}

var sentDiskMap = make(map[string]bool)

func checkDisk() {
	if diskTh == 0 {
		return
	}
	paths, err := disk.Partitions(false)
	if err != nil {
		log.Printf("checkDisk err=%v", err)
		return
	}
	for _, p := range paths {
		us, err := disk.Usage(p.Mountpoint)
		if err != nil {
			log.Printf("checkDisk err=%v", err)
			continue
		}
		if isExcludeDisk(p) {
			continue
		}
		sentDisk, ok := sentDiskMap[p.Mountpoint]
		if !ok {
			sentDisk = false
			sentDiskMap[p.Mountpoint] = false
		}
		intDisk := int(us.UsedPercent)
		if diskTh < intDisk {
			if all || !sentDisk {
				send("disk:"+p.Mountpoint, diskTh, intDisk, false)
				sentDiskMap[p.Mountpoint] = true
			}
		} else if sentDisk {
			send("disk:"+p.Mountpoint, diskTh, intDisk, true)
			sentDiskMap[p.Mountpoint] = false
		}
	}
}

func isExcludeDisk(p disk.PartitionStat) bool {
	if p.Fstype == "devfs" {
		return true
	}
	if strings.HasPrefix(p.Mountpoint, "/System/Volumes") {
		return true
	}
	return false
}

func send(hrName string, th, val int, normal bool) {
	var err error
	if syslog != "" {
		if normal {
			err = sendSyslog(fmt.Sprintf("%s back to normal %d%% > %d%%", hrName, th, val), 6)
		} else {
			err = sendSyslog(fmt.Sprintf("%s over thresold %d%% < %d%%", hrName, th, val), 3)
		}
	}
	if trap != "" {
		dsts := strings.Split(trap, ",")
		for _, t := range dsts {
			err = sendTrap(t, hrName, th, val, normal)
		}
	}
	if err != nil {
		log.Printf("send err=%v", err)
	}
}

func sendTrap(target, hrName string, th, val int, normal bool) error {
	log.Printf("sendTrap %s %s %d %d %v", target, hrName, th, val, normal)
	port := 162
	ta := strings.SplitN(target, ":", 2)
	if len(ta) > 1 {
		target = ta[0]
		if v, err := strconv.ParseInt(ta[1], 10, 64); err == nil && v > 0 && v < 0xfffe {
			port = int(v)
		}
	}
	gosnmp.Default.Target = target
	gosnmp.Default.Port = uint16(port)
	gosnmp.Default.Timeout = time.Duration(3) * time.Second
	switch snmpMode {
	case "v3auth":
		gosnmp.Default.Version = gosnmp.Version3
		gosnmp.Default.SecurityModel = gosnmp.UserSecurityModel
		gosnmp.Default.MsgFlags = gosnmp.AuthNoPriv
		gosnmp.Default.SecurityParameters = &gosnmp.UsmSecurityParameters{
			UserName:                 user,
			AuthoritativeEngineID:    engineID,
			AuthenticationProtocol:   gosnmp.SHA,
			AuthenticationPassphrase: passwd,
		}
	case "v3authpriv":
		gosnmp.Default.Version = gosnmp.Version3
		gosnmp.Default.SecurityModel = gosnmp.UserSecurityModel
		gosnmp.Default.MsgFlags = gosnmp.AuthPriv
		gosnmp.Default.SecurityParameters = &gosnmp.UsmSecurityParameters{
			UserName:                 user,
			AuthoritativeEngineID:    engineID,
			AuthenticationProtocol:   gosnmp.SHA,
			AuthenticationPassphrase: passwd,
			PrivacyProtocol:          gosnmp.AES,
			PrivacyPassphrase:        passwd,
		}
	default:
		gosnmp.Default.Version = gosnmp.Version2c
		gosnmp.Default.Community = community
	}
	err := gosnmp.Default.Connect()
	if err != nil {
		return err
	}
	defer gosnmp.Default.Conn.Close()
	hrType := hrName
	a := strings.SplitN(hrName, ":", 2)
	if len(a) > 1 {
		hrType = a[0]
	}
	vbs := []gosnmp.SnmpPDU{}
	if normal {
		vbs = append(vbs,
			gosnmp.SnmpPDU{
				Name:  ".1.3.6.1.6.3.1.1.4.1.0",
				Type:  gosnmp.ObjectIdentifier,
				Value: ".1.3.6.1.4.1.17861.1.10.0.5",
			})
	} else {
		switch hrType {
		case "cpu":
			vbs = append(vbs,
				gosnmp.SnmpPDU{
					Name:  ".1.3.6.1.6.3.1.1.4.1.0",
					Type:  gosnmp.ObjectIdentifier,
					Value: ".1.3.6.1.4.1.17861.1.10.0.1",
				})
		case "mem":
			vbs = append(vbs,
				gosnmp.SnmpPDU{
					Name:  ".1.3.6.1.6.3.1.1.4.1.0",
					Type:  gosnmp.ObjectIdentifier,
					Value: ".1.3.6.1.4.1.17861.1.10.0.2",
				})
		case "load":
			vbs = append(vbs,
				gosnmp.SnmpPDU{
					Name:  ".1.3.6.1.6.3.1.1.4.1.0",
					Type:  gosnmp.ObjectIdentifier,
					Value: ".1.3.6.1.4.1.17861.1.10.0.3",
				})
		case "disk":
			vbs = append(vbs,
				gosnmp.SnmpPDU{
					Name:  ".1.3.6.1.6.3.1.1.4.1.0",
					Type:  gosnmp.ObjectIdentifier,
					Value: ".1.3.6.1.4.1.17861.1.10.0.4",
				})
		default:
			return fmt.Errorf("invalid hrType=%s", hrType)
		}
	}
	vbs = append(vbs,
		gosnmp.SnmpPDU{
			Name:  "..1.3.6.1.4.1.17861.1.10.1.1.0",
			Type:  gosnmp.OctetString,
			Value: hrName,
		})
	vbs = append(vbs,
		gosnmp.SnmpPDU{
			Name:  "..1.3.6.1.4.1.17861.1.10.1.2.0",
			Type:  gosnmp.Integer,
			Value: th,
		})
	vbs = append(vbs,
		gosnmp.SnmpPDU{
			Name:  "..1.3.6.1.4.1.17861.1.10.1.3.0",
			Type:  gosnmp.Integer,
			Value: val,
		})

	trap := gosnmp.SnmpTrap{
		Variables: vbs,
	}
	_, err = gosnmp.Default.SendTrap(trap)
	return err
}

func sendSyslog(msg string, severity int) error {
	var ret error
	log.Printf("sendSyslog %s %d", msg, severity)
	host, err := os.Hostname()
	if err != nil {
		host = "localhost"
	}
	dsts := strings.Split(syslog, ",")
	for _, d := range dsts {
		if !strings.Contains(d, ":") {
			d += ":514"
		}
		s, err := net.Dial("udp", d)
		if err != nil {
			log.Printf("sendSyslog err=%v", err)
			if ret == nil {
				ret = err
			}
			continue
		}
		m := fmt.Sprintf("<%d>%s %s twhr2trap: %s", 21*8+severity, time.Now().Format("2006-01-02T15:04:05-07:00"), host, msg)
		s.Write([]byte(m))
		s.Close()
	}
	return ret
}
