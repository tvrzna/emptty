package src

import (
	"bufio"
	"errors"
	"fmt"
	"net"
	"os"
	"os/exec"
	"os/user"
	"path/filepath"
	"strconv"
	"strings"
	"syscall"
	"unsafe"
)

const (
	pathOsReleaseFile = "/etc/os-release"

	osReleasePrettyName = "PRETTY_NAME"
	osReleaseName       = "NAME"

	devPath = "/dev/"

	_KDGKBTYPE     = 0x4B33
	_VT_ACTIVATE   = 0x5606
	_VT_WAITACTIVE = 0x5607

	_KB_101 = 0x02
	_KB_84  = 0x01

	currentTty = "/dev/tty"
	devConsole = "/dev/console"
)

// propertyFunc defines method to be invoked during readProperties method for each record.
type propertyFunc func(key, value string)

// readProperties reads defined filePath per line and parses each key-value pair.
// These pairs are used as parameters for invoking propertyFunc
func readProperties(filePath string, method propertyFunc) error {
	return readPropertiesWithSupport(filePath, method, false)
}

// readPropertiesWithSupport reads defined filePath per line and parses each key-value pair with possible fish shell support.
// These pairs are used as parameters for invoking propertyFunc
func readPropertiesWithSupport(filePath string, method propertyFunc, fishSupport bool) error {
	file, err := os.Open(filePath)
	if err != nil {
		return errors.New("Could not open file " + filePath)
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	requiresFishSupport := false
	isFirstLine := true
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if isFirstLine {
			if fishSupport && strings.HasPrefix(line, "#!") && strings.Contains(line, "/fish") {
				requiresFishSupport = true
			}
			isFirstLine = false
		}

		readPropertyLine(line, method, requiresFishSupport)
	}
	return scanner.Err()
}

// Reads single property line and parses its content into key-value pair.
// The pair is used as parameter for invoking propertyFunc.
func readPropertyLine(line string, method propertyFunc, fishSupport bool) {
	if !strings.HasPrefix(line, "#") && ((!fishSupport && strings.Contains(line, "=")) || fishSupport && strings.HasPrefix(line, "set")) {
		var splitIndex int
		if fishSupport {
			line = line[4:]
			splitIndex = strings.Index(line, " ")
		} else {
			splitIndex = strings.Index(line, "=")
		}

		key := strings.ReplaceAll(line[:splitIndex], "export ", "")
		value := line[splitIndex+1:]
		if strings.Contains(value, "#") {
			value = value[:strings.Index(value, "#")]
		}
		key = strings.ToUpper(strings.TrimSpace(key))
		value = strings.TrimSpace(value)
		for (strings.HasPrefix(value, "\"") && strings.HasSuffix(value, "\"")) || (strings.HasPrefix(value, "'") && strings.HasSuffix(value, "'")) {
			value = value[1 : len(value)-1]
		}
		method(key, value)
	}
}

// Reads properties from defined filePath into key-value map pair.
// The result map is returned, if no error appears.
func readPropertiesToMap(filePath string) (result map[string]string, err error) {
	result = make(map[string]string)
	err = readProperties(filePath, func(key, value string) {
		result[key] = value
	})
	if err != nil {
		return nil, err
	}
	return result, nil
}

// Checks, if file on path exists.
func fileExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

// Checks, if file on path exists and is executable.
func fileIsExecutable(path string) bool {
	stat, err := os.Stat(path)
	return err == nil && (stat.Mode()&0100 == 0100)
}

// Sanitize value.
func sanitizeValue(value, defaultValue string) string {
	if value == "" {
		return defaultValue
	}
	return strings.TrimSpace(value)
}

// Makes directories up to last part of path (to make sure to not make dir, that is named as result file)
func mkDirsForFile(path string, perm os.FileMode) error {
	if !fileExists(path) && path != "" {
		return os.MkdirAll(path[:strings.LastIndex(path, "/")], perm)
	}
	return nil
}

// Converts color by name into ANSI color number.
func convertColor(name string, isForeground bool) string {
	colorName := strings.ToUpper(name)
	isLight := strings.HasPrefix(colorName, "LIGHT_")
	colorName = strings.Replace(colorName, "LIGHT_", "", -1)
	colorNumber := 0

	switch colorName {
	case "":
		colorNumber = 0
	case "BLACK":
		colorNumber = 30
	case "RED":
		colorNumber = 31
	case "GREEN":
		colorNumber = 32
	case "YELLOW":
		colorNumber = 33
	case "BLUE":
		colorNumber = 34
	case "MAGENTA":
		colorNumber = 35
	case "CYAN":
		colorNumber = 36
	case "WHITE":
		colorNumber = 37
	default:
		return ""
	}

	if colorNumber > 0 {
		if !isForeground {
			colorNumber += 10
		}
		if isLight {
			colorNumber += 60
		}
	}

	return strconv.Itoa(colorNumber)
}

func stringColor(value string, isForeground bool) string {
	colorNumber, err := strconv.Atoi(value)
	if err != nil {
		return ""
	}

	if !isForeground {
		colorNumber -= 10
	}

	isLight := false
	if colorNumber >= 90 {
		isLight = true
		colorNumber -= 60
	}

	colorValue := ""
	switch colorNumber {
	case 0:
		return ""
	case 30:
		colorValue = "BLACK"
	case 31:
		colorValue = "RED"
	case 32:
		colorValue = "GREEN"
	case 33:
		colorValue = "YELLOW"
	case 34:
		colorValue = "BLUE"
	case 35:
		colorValue = "MAGENTA"
	case 36:
		colorValue = "CYAN"
	case 37:
		colorValue = "WHITE"
	default:
		return ""
	}

	if isLight {
		colorValue = "LIGHT_" + colorValue
	}
	return colorValue
}

// Prepares *exec.Cmd to be started as sysuser.
func cmdAsUser(usr *sysuser, name string, arg ...string) *exec.Cmd {
	if strings.Contains(name, " ") {
		nameArgs := parseExec(name)
		name = nameArgs[0]
		arg = append(nameArgs[1:], arg...)
	}
	cmd := exec.Command(name, arg...)
	cmd.Env = usr.environ()
	cmd.SysProcAttr = &syscall.SysProcAttr{}
	cmd.SysProcAttr.Credential = &syscall.Credential{Uid: usr.uidu32(), Gid: usr.gidu32(), Groups: usr.gidsu32}
	return cmd
}

// Processes selected command as exec.Cmd
func processCommandAsCmd(name string) error {
	var arg []string
	if strings.Contains(name, " ") {
		nameArgs := parseExec(name)
		name = nameArgs[0]
		arg = append(nameArgs[1:], arg...)
	}
	name, _ = exec.LookPath(name)
	cmd := exec.Command(name, arg...)
	return cmd.Run()
}

// Splits execString by spaces respecting double quotes.
func parseExec(execString string) (result []string) {
	var sb strings.Builder
	inQuotes := false
	escapeNext := false

	for _, r := range execString {
		switch {
		case escapeNext:
			escapeNext = false
		case r == '\\':
			escapeNext = true
		case r == '"':
			inQuotes = !inQuotes
		case r == ' ' && !inQuotes:
			if sb.Len() > 0 {
				result = append(result, sb.String())
				sb.Reset()
				continue
			}
		}
		sb.WriteRune(r)
	}
	if sb.Len() > 0 {
		result = append(result, sb.String())
	}
	return
}

// Applies current resource limits
func applyRlimits() {
	rlimits := []int{syscall.RLIMIT_AS, syscall.RLIMIT_CORE, syscall.RLIMIT_CPU, syscall.RLIMIT_DATA, syscall.RLIMIT_FSIZE, syscall.RLIMIT_NOFILE, syscall.RLIMIT_STACK}

	rlimit := &syscall.Rlimit{}
	for _, r := range rlimits {
		if err := syscall.Getrlimit(r, rlimit); err != nil {
			logPrintf("could not get rlimit %d", r)
			continue
		}
		if err := syscall.Setrlimit(r, rlimit); err != nil {
			logPrintf("could not set rlimit %d(soft: %d, max: %d)", r, rlimit.Cur, rlimit.Max)
		}
	}
}

// Checks, if array contains any of values
func contains(array []string, values ...string) bool {
	for _, a := range array {
		for _, v := range values {
			if a == v {
				return true
			}
		}
	}
	return false
}

// Parse boolean values.
func parseBool(strBool, defaultValue string) bool {
	val, err := strconv.ParseBool(sanitizeValue(strBool, defaultValue))
	if err != nil {
		return false
	}
	return val
}

// Runs simple command and returns its output as string
func runSimpleCmd(cmd ...string) string {
	return runSimpleCmdAsUser(nil, cmd...)
}

// Runs simple command as user and returns its output as string
func runSimpleCmdAsUser(usr *sysuser, cmd ...string) string {
	path, err := exec.LookPath(cmd[0])
	if err != nil {
		logPrintf("Could not find command '%s' on PATH", cmd[0])
		return ""
	}

	execCmd := exec.Command(path, cmd[1:]...)

	if usr != nil {
		execCmd.Env = usr.environ()
		execCmd.SysProcAttr = &syscall.SysProcAttr{}
		execCmd.SysProcAttr.Credential = &syscall.Credential{Uid: usr.uidu32(), Gid: usr.gidu32(), Groups: usr.gidsu32}
	}

	if output, err := execCmd.Output(); err == nil {
		return strings.TrimSpace(string(output))
	}
	return ""
}

// Look for path of cmd, if not found, use fallback
func lookPath(cmd string, fallback string) string {
	path, err := exec.LookPath(cmd)
	if err != nil {
		logPrintf("Could not find command '%s' on PATH, using fallback '%s'", cmd, fallback)
		return fallback
	}
	return path
}

// Tries to find corresponding interface and its IP address
func getIpAddress(name string, ipType byte) string {
	if name == "" {
		ifaces, err := net.Interfaces()
		if err != nil {
			logPrint(err)
			return ""
		}
		for _, iface := range ifaces {
			if iface.Flags&net.FlagUp > 0 && iface.Flags&net.FlagLoopback == 0 {
				return getIpAddressFromIface(&iface, ipType)
			}
		}
	} else {
		iface, err := net.InterfaceByName(name)
		if err != nil {
			logPrint(err)
			return ""
		}
		return getIpAddressFromIface(iface, ipType)
	}

	return ""
}

// Gets corresponding IP address from interface
func getIpAddressFromIface(iface *net.Interface, ipType byte) string {
	if iface.Flags&net.FlagUp == 0 || iface.Flags&net.FlagLoopback != 0 {
		return ""
	}
	addrs, err := iface.Addrs()
	if err != nil {
		logPrint(err)
		return ""
	}
	for _, addr := range addrs {
		var ip net.IP
		switch v := addr.(type) {
		case *net.IPNet:
			ip = v.IP
		case *net.IPAddr:
			ip = v.IP
		}
		if ip == nil {
			return ""
		}
		if ipType == '4' {
			if ip.To4() != nil {
				return ip.To4().String()
			}
		} else {
			if ip.To4() == nil {
				return ip.To16().String()
			}
		}
	}
	return ""
}

// Gets value from /etc/os-release. If no name is defined, it assumes PRETTY_NAME or NAME, if PRETTY_NAME is not defined.
func getOsReleaseValue(name string) string {
	var values = make(map[string]string)
	readProperties(pathOsReleaseFile, func(key, value string) {
		if len(value) > 1 {
			values[key] = value
		}
	})

	if name == "" {
		if values[osReleasePrettyName] != "" {
			return values[osReleasePrettyName]
		}
		return values[osReleaseName]
	}
	return values[name]
}

// Do operation as user and then reverts to previous user.
func doAsUser(usr *sysuser, fce func()) {
	currentUser, _ := user.Current()
	previousUser := getSysuser(currentUser)

	setFsUser(usr)

	fce()

	setFsUser(previousUser)
}

// Gets current TTY name
func getCurrentTTYName(fallback string, fullname bool) string {
	if name, err := filepath.EvalSymlinks(os.Stdout.Name()); err == nil {
		if fullname {
			return name
		}
		return name[strings.LastIndex(name, devPath)+len(devPath):]
	}
	// if tty name fails, try to run ps command
	if result := runSimpleCmd("ps", "-p", strconv.Itoa(os.Getpid()), "-o", "tty", "--no-headers"); result != "" {
		if fullname {
			return filepath.Join(devPath, result)
		}
		return result
	}
	if fullname {
		return filepath.Join(devPath, fallback)
	}
	return fallback
}

// Gets DNS domain name of current machine
func getDnsDomainName() string {
	if host, err := os.Hostname(); err == nil {
		var domain string
		if canonname, err := net.LookupCNAME(host); err == nil {
			domain = canonname[strings.Index(canonname, ".")+1:]
		}
		if domain == "" {
			if ip, err := net.LookupHost(host); err == nil && len(ip) > 0 {
				if domains, err := net.LookupAddr(ip[0]); err == nil {
					for _, d := range domains {
						if d[len(d)-1:] == "." {
							domain = d[strings.Index(d, ".")+1:]
							break
						}
					}
				}
			}
		}
		if domain != "" && domain[len(domain)-1:] == "." {
			return domain[:len(domain)-1]
		}
	}
	return "unknown_domain"
}

// Opens console by its path
func openConsole(path string) *os.File {
	for _, flag := range []int{os.O_RDWR, os.O_RDONLY, os.O_WRONLY} {
		if c, err := os.OpenFile(path, flag, 0700); err == nil {
			return c
		}
	}
	return nil
}

// Checks, if used fd is a console
func isConsole(fd uintptr) bool {
	flag := 0
	if _, _, errNo := syscall.Syscall(syscall.SYS_IOCTL, fd, uintptr(_KDGKBTYPE), uintptr(unsafe.Pointer(&flag))); errNo == 0 {
		return flag == _KB_101 || flag == _KB_84
	}
	return false
}

// Gets console to change the TTY
func getConsole() *os.File {
	for _, name := range []string{currentTty, currentVc, devConsole} {
		if c := openConsole(name); c != nil {
			if isConsole(c.Fd()) {
				return c
			}
			c.Close()
		}
	}
	return nil
}

// Performs chvt command using ioctl
func chvt(tty int) bool {
	if c := getConsole(); c != nil {
		defer c.Close()
		if _, _, errNo := syscall.Syscall(syscall.SYS_IOCTL, uintptr(c.Fd()), uintptr(_VT_ACTIVATE), uintptr(tty)); errNo > 0 {
			return false
		}
		if _, _, errNo := syscall.Syscall(syscall.SYS_IOCTL, uintptr(c.Fd()), uintptr(_VT_WAITACTIVE), uintptr(tty)); errNo > 0 {
			return false
		}
	}
	return true
}

func waitForReturnToExit(code int) {
	fmt.Printf("\nPress Enter to continue...")
	if !TEST_MODE {
		bufio.NewReader(os.Stdin).ReadString('\n')
		os.Exit(code)
	}
}
