//go:build linux

package cli

import (
	"os"
	"regexp"
	"strconv"
	"strings"

	"github.com/pkg/errors"
)

var rePerProcess = regexp.MustCompile(`/proc/\d+/(.+)`)

func (c *CLI) createInputIter(query string, args []string) (inputIter, error) {
	if len(args) < 1 {
		return nil, errors.New("file name argument is required")
	}

	fname := args[0]
	f, err := os.Open(fname)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	submatch := rePerProcess.FindStringSubmatch(fname)
	if len(submatch) == 2 {
		switch submatch[1] {
		case "io":
			return newProcMapIter(fname, f, createMapParser(createLineParser(splitLineByColon, parseProcPidIoValue)))
		case "limits":
			return newProcTableIter(fname, f, createTableParser(skipTableHeader(1), createTableRowParser(splitProcPidLimitsColumns, parseProcPidLimitsColumns)))
		}
	}

	switch fname {
	case "/proc/cpuinfo":
		return newProcArrayIter(fname, f, createChunkParser(createLineParser(splitLineByColon, parseProcCpuinfoValue)))
	case "/proc/crypto":
		return newProcArrayIter(fname, f, createChunkParser(createLineParser(splitLineByColon, parseProcCryptoValue)))
	case "/proc/diskstats":
		return newProcTableIter(fname, f, createTableParser(noTableHeader, createTableRowParser(splitColumnsBySpace, parseProcDiskstatsColumns)))
	case "/proc/meminfo":
		return newProcMapIter(fname, f, createMapParser(createLineParser(splitLineByColon, parseProcMeminfoValue)))
	case "/proc/modules":
		return newProcTableIter(fname, f, createTableParser(noTableHeader, createTableRowParser(splitColumnsBySpace, parseProcModulesColumns)))
	case "/proc/mounts":
		return newProcTableIter(fname, f, createTableParser(noTableHeader, createTableRowParser(splitColumnsBySpace, parseProcMountsColumns)))
	case "/proc/net/arp", "/proc/self/net/arp":
		return newProcTableIter(fname, f, createTableParser(skipTableHeader(1), createTableRowParser(splitColumnsBySpace, parseProcNetArpColumns)))
	case "/proc/net/dev", "/proc/self/net/dev":
		return newProcTableIter(fname, f, createTableParser(skipTableHeader(2), createTableRowParser(splitColumnsByColonAndSpace, parseProcNetDevColumns)))
	case "/proc/net/netlink", "/proc/self/net/netlink":
		return newProcTableIter(fname, f, createTableParser(skipTableHeader(1), createTableRowParser(splitColumnsBySpace, parseProcNetNetlinkColumns)))
	case "/proc/vmstat":
		return newProcMapIter(fname, f, createMapParser(createLineParser(splitLineBySpace, parseProcVmstatValue)))
	}
	return nil, errors.Errorf("%s is not supported", fname)
}

func parseProcPidIoValue(key, valueStr string) (interface{}, error) {
	return strconv.ParseInt(valueStr, 10, 64)
}

func splitProcPidLimitsColumns(row string) ([]string, error) {
	if len(row) < 68 {
		return nil, errors.New("unexpected /proc/{pid}/limits format")
	}

	var columns []string

	columns = append(columns, strings.TrimSpace(row[:26]))
	columns = append(columns, strings.TrimSpace(row[26:47]))
	columns = append(columns, strings.TrimSpace(row[47:68]))
	columns = append(columns, strings.TrimSpace(row[68:]))

	return columns, nil
}

func parseProcPidLimitsColumns(header, columns []string) (map[string]interface{}, error) {
	sl, err := unlimitedOrInteger(columns[1])
	if err != nil {
		return nil, err
	}

	hl, err := unlimitedOrInteger(columns[2])
	if err != nil {
		return nil, err
	}

	result := map[string]interface{}{
		"limit":      columns[0],
		"soft_limit": sl,
		"hard_limit": hl,
	}

	if columns[3] != "" {
		result["units"] = columns[3]
	}

	return result, nil
}

func unlimitedOrInteger(s string) (interface{}, error) {
	if s == "unlimited" {
		return s, nil
	}

	return strconv.ParseInt(s, 10, 64)
}

func parseProcCpuinfoValue(key, valueStr string) (interface{}, error) {
	switch key {

	// strings
	case
		"address sizes",   // amd64
		"cache size",      // amd64
		"CPU implementer", // arm64
		"CPU part",        // arm64
		"CPU variant",     // arm64
		"fpu",             // amd64
		"fpu_exception",   // amd64
		"isa",             // riscv
		"microcode",       // amd64
		"mmu",             // riscv
		"model name",      // amd64
		"vendor_id",       // amd64
		"wp":              // amd64

		return valueStr, nil

	// integers
	case
		"apicid",           // amd64
		"cache_alignment",  // amd64
		"clflush size",     // amd64
		"core id",          // amd64
		"CPU architecture", // arm64
		"CPU revision",     // arm64
		"cpu cores",        // amd64
		"cpu family",       // amd64
		"cpuid level",      // amd64
		"hart",             // riscv
		"initial apicid",   // amd64
		"model",            // amd64
		"physical id",      // amd64
		"processor",        // amd64
		"siblings",         // amd64
		"stepping":         // amd64

		val, err := strconv.ParseInt(valueStr, 10, 64)
		if err != nil {
			return nil, err
		}
		return val, nil

	// floats
	case
		"bogomips", // amd64
		"BogoMIPS", // arm64
		"cpu MHz":  // amd64

		val, err := strconv.ParseFloat(valueStr, 64)
		if err != nil {
			return nil, err
		}
		return val, nil

	// space separated strings
	case
		"bugs",      // amd64
		"Features",  // arm64
		"flags",     // amd64
		"vmx flags": // amd64

		var val []interface{}
		s := strings.Split(valueStr, " ")
		for _, v := range s {
			val = append(val, v)
		}
		return val, nil

	// comma separated strings
	case
		"uarch": // riscv

		var val []interface{}
		s := strings.Split(valueStr, ",")
		for _, v := range s {
			val = append(val, v)
		}
		return val, nil

	// unknown
	case
		"power management": // amd64

		return valueStr, nil

	default:
		return nil, errors.Errorf("unknown cpuinfo key: %s", key)
	}
}

func parseProcCryptoValue(key, valueStr string) (interface{}, error) {
	switch key {

	// strings
	case
		"async",
		"driver",
		"geniv",
		"internal",
		"module",
		"name",
		"selftest",
		"type":

		return valueStr, nil

	// integers
	case
		"blocksize",
		"chunksize",
		"digestsize",
		"ivsize",
		"max keysize",
		"maxauthsize",
		"min keysize",
		"priority",
		"refcnt",
		"seedsize",
		"walksize":

		val, err := strconv.ParseInt(valueStr, 10, 64)
		if err != nil {
			return nil, err
		}
		return val, nil

	default:
		return nil, errors.Errorf("unknown crypto key: %s", key)
	}
}

func parseProcDiskstatsColumns(_ []string, columns []string) (map[string]interface{}, error) {
	if len(columns) < 20 {
		return nil, errors.Errorf("unknown /proc/diskstats format: found %d columns", len(columns))
	}

	labels := []string{"major", "minor", "name", "ios_read", "merges_read", "sectors_read", "msecs_read", "ios_write", "merges_write", "sectors_write", "msecs_write", "inflight", "io_ticks", "msecs_total", "ios_discard", "merges_discard", "sectors_discard", "msec_discard", "ios_flush", "msec_flush"}

	result := make(map[string]interface{})
	for i, col := range columns {
		if i != 2 {
			val, err := strconv.ParseInt(col, 10, 64)
			if err != nil {
				return nil, err
			}
			result[labels[i]] = val
		} else {
			result[labels[i]] = col
		}
	}
	return result, nil
}

func parseProcMeminfoValue(key, valueStr string) (interface{}, error) {
	if strings.HasSuffix(valueStr, " kB") {
		val, err := strconv.ParseInt(strings.TrimSuffix(valueStr, " kB"), 10, 64)
		if err != nil {
			return nil, err
		}
		return map[string]interface{}{
			"value": val,
			"unit":  "kB",
		}, nil
	}

	if isLikelyInteger(valueStr) {
		val, err := strconv.ParseInt(valueStr, 10, 64)
		if err != nil {
			return nil, err
		}
		return map[string]interface{}{
			"value": val,
		}, nil
	}

	return map[string]interface{}{
		"value": valueStr,
	}, nil
}

func parseProcModulesColumns(_ []string, columns []string) (map[string]interface{}, error) {
	if len(columns) < 6 {
		return nil, errors.Errorf("unknown /proc/modules format. expected 6 columns but got %d columns", len(columns))
	}

	var err error
	result := make(map[string]interface{})
	result["name"] = columns[0]
	result["size"], err = strconv.ParseInt(columns[1], 10, 64)
	if err != nil {
		return nil, err
	}
	result["ref_count"], err = strconv.ParseInt(columns[2], 10, 64)
	if err != nil {
		return nil, err
	}

	var sources []interface{}
	if strings.HasSuffix(columns[3], ",") {
		s := strings.Split(strings.TrimSuffix(columns[3], ","), ",")
		for _, src := range s {
			sources = append(sources, src)
		}
	}
	if sources != nil {
		result["sources"] = sources
	}

	result["state"] = columns[4]
	result["base"] = columns[5]

	return result, nil
}

func parseProcMountsColumns(_ []string, columns []string) (map[string]interface{}, error) {
	if len(columns) < 6 {
		return nil, errors.Errorf("unknown /proc/mounts format. expected 6 columns but got %d columns", len(columns))
	}

	result := make(map[string]interface{})
	result["device"] = columns[0]
	result["mount_point"] = columns[1]
	result["type"] = columns[2]

	var options []interface{}
	for _, opt := range strings.Split(columns[3], ",") {
		options = append(options, opt)
	}
	result["options"] = options

	return result, nil
}

func parseProcNetArpColumns(_ []string, columns []string) (map[string]interface{}, error) {
	if len(columns) < 6 {
		return nil, errors.Errorf("unknown /proc/net/arp format. expected 6 columns but got %d columns", len(columns))
	}

	result := make(map[string]interface{})
	result["ip_address"] = columns[0]
	result["hw_type"] = columns[1]
	result["flags"] = columns[2]
	result["hw_address"] = columns[3]
	result["mask"] = columns[4]
	result["device"] = columns[5]

	return result, nil
}

func parseProcNetDevColumns(_ []string, columns []string) (map[string]interface{}, error) {
	if len(columns) < 17 {
		return nil, errors.Errorf("unknown /proc/net/dev format. expected 17 columns but got %d columns", len(columns))
	}

	var err error
	result := make(map[string]interface{})
	result["interface"] = columns[0]
	receive := make(map[string]interface{})
	receive["bytes"], err = strconv.ParseInt(columns[1], 10, 64)
	if err != nil {
		return nil, err
	}
	receive["packets"], err = strconv.ParseInt(columns[2], 10, 64)
	if err != nil {
		return nil, err
	}
	receive["errs"], err = strconv.ParseInt(columns[3], 10, 64)
	if err != nil {
		return nil, err
	}
	receive["drop"], err = strconv.ParseInt(columns[4], 10, 64)
	if err != nil {
		return nil, err
	}
	receive["fifo"], err = strconv.ParseInt(columns[5], 10, 64)
	if err != nil {
		return nil, err
	}
	receive["frame"], err = strconv.ParseInt(columns[6], 10, 64)
	if err != nil {
		return nil, err
	}
	receive["compressed"], err = strconv.ParseInt(columns[7], 10, 64)
	if err != nil {
		return nil, err
	}
	receive["multicast"], err = strconv.ParseInt(columns[8], 10, 64)
	if err != nil {
		return nil, err
	}
	result["receive"] = receive
	transmit := make(map[string]interface{})
	transmit["bytes"], err = strconv.ParseInt(columns[9], 10, 64)
	if err != nil {
		return nil, err
	}
	transmit["packets"], err = strconv.ParseInt(columns[10], 10, 64)
	if err != nil {
		return nil, err
	}
	transmit["errs"], err = strconv.ParseInt(columns[11], 10, 64)
	if err != nil {
		return nil, err
	}
	transmit["drop"], err = strconv.ParseInt(columns[12], 10, 64)
	if err != nil {
		return nil, err
	}
	transmit["fifo"], err = strconv.ParseInt(columns[13], 10, 64)
	if err != nil {
		return nil, err
	}
	transmit["frame"], err = strconv.ParseInt(columns[14], 10, 64)
	if err != nil {
		return nil, err
	}
	transmit["compressed"], err = strconv.ParseInt(columns[15], 10, 64)
	if err != nil {
		return nil, err
	}
	transmit["multicast"], err = strconv.ParseInt(columns[16], 10, 64)
	if err != nil {
		return nil, err
	}
	result["transmit"] = transmit

	return result, nil
}

func parseProcNetNetlinkColumns(_ []string, columns []string) (map[string]interface{}, error) {
	if len(columns) < 10 {
		return nil, errors.Errorf("unknown /proc/net/netlink format. expected 10 columns but got %d columns", len(columns))
	}

	var err error
	result := make(map[string]interface{})
	result["sk"] = columns[0]
	result["eth"], err = strconv.ParseInt(columns[1], 10, 64)
	if err != nil {
		return nil, err
	}
	result["pid"], err = strconv.ParseInt(columns[2], 10, 64)
	if err != nil {
		return nil, err
	}
	result["groups"] = columns[3]
	result["rmem"], err = strconv.ParseInt(columns[4], 10, 64)
	if err != nil {
		return nil, err
	}
	result["wmem"], err = strconv.ParseInt(columns[5], 10, 64)
	if err != nil {
		return nil, err
	}
	result["dump"], err = strconv.ParseInt(columns[6], 10, 64)
	if err != nil {
		return nil, err
	}
	result["locks"], err = strconv.ParseInt(columns[7], 10, 64)
	if err != nil {
		return nil, err
	}
	result["drops"], err = strconv.ParseInt(columns[8], 10, 64)
	if err != nil {
		return nil, err
	}
	result["inode"], err = strconv.ParseInt(columns[9], 10, 64)
	if err != nil {
		return nil, err
	}

	return result, nil
}

func parseProcVmstatValue(key, valueStr string) (interface{}, error) {
	return strconv.ParseInt(valueStr, 10, 64)
}
