package src

import (
	"bufio"
	"errors"
	"net"
	"os"
	"os/exec"
	"os/signal"
	"os/user"
	"path/filepath"
	"strconv"
	"strings"
	"syscall"
)

const (
	pathOsReleaseFile = "/etc/os-release"

	osReleasePrettyName = "PRETTY_NAME"
	osReleaseName       = "NAME"

	devPath = "/dev/"
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

// Prepares *exec.Cmd to be started as sysuser.
func cmdAsUser(usr *sysuser, name string, arg ...string) *exec.Cmd {
	if strings.Contains(name, " ") {
		nameArgs := strings.Split(name, " ")
		name = nameArgs[0]
		arg = append(nameArgs[1:], arg...)
	}
	cmd := exec.Command(name, arg...)
	cmd.Env = usr.environ()
	cmd.SysProcAttr = &syscall.SysProcAttr{}
	cmd.SysProcAttr.Credential = &syscall.Credential{Uid: usr.uidu32(), Gid: usr.gidu32(), Groups: usr.gidsu32}
	return cmd
}

// Checks, if array contains value
func contains(array []string, value string) bool {
	for _, v := range array {
		if v == value {
			return true
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

	output, err := execCmd.Output()
	if err == nil {
		return strings.TrimSpace(string(output))
	}
	return ""
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
			values[key] = value[1 : len(value)-1]
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

// Make channel for catching interrupts.
func makeInterruptChannel() chan os.Signal {
	c := make(chan os.Signal, 10)
	signal.Notify(c, os.Interrupt, syscall.SIGHUP, syscall.SIGINT, syscall.SIGKILL, syscall.SIGQUIT, syscall.SIGTERM)
	return c
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
