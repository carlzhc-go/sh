// Golang library for various shell commands
package sh

import (
	"flag"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/icza/gog"
)

/*
command: command [-pVv] command [arg ...]

	Execute a simple command or display information about commands.

	Runs COMMAND with ARGS suppressing  shell function lookup, or display
	information about the specified COMMANDs.  Can be used to invoke commands
	on disk when a function with the same name exists.

	Options:
	  -p    use a default value for PATH that is guaranteed to find all of
	        the standard utilities (Ignored and always true)
	  -v    print a description of COMMAND similar to the `type' builtin
	  -V    print a more verbose description of each COMMAND (Not implemented)

	Exit Status:
	Returns exit status of COMMAND, or failure if COMMAND is not found.
*/
func Command(args ...string) (result []byte, err error) {

	var (
		opt_p   bool
		opt_v   bool
		cmdLine = flag.NewFlagSet("sh.Command", flag.ExitOnError)
	)
	cmdLine.BoolVar(&opt_p, "p", false, "use ad efault value for PATH that is guaranteed to find all the standard utilities")
	cmdLine.BoolVar(&opt_v, "v", false, "print a description of COMMAND similar to the `type' builtin")
	cmdLine.Parse(args)

	if opt_v {
		for _, cmd := range cmdLine.Args() {
			cmdPath, err := exec.LookPath(cmd)
			if err != nil {
				return nil, err
			}
			result = append(result, []byte(cmdPath)...)
			result = append(result, 0xa) // newline
		}

	} else if len(cmdLine.Args()) != 0 {
		args := cmdLine.Args()
		runCmd := exec.Command(args[0], args[1:]...)

		result, err = runCmd.Output()
	}

	return result, err
}

func Test(args ...string) (result bool) {
	var op, file string
	var idx = 0

	if args[idx] == "!" {
		defer func() {
			result = !result
		}()

		idx += 1
	}

	op = args[idx]
	file = args[idx+1]

	if len(op) != 2 || op[0] != '-' {
		panic("Wrong format of operator: " + op)
	}

	fileInfo, err := os.Stat(file)
	if err != nil {
		result = false
		return
	}

	switch op[1] {
	case 'a', 'e':
		result = true
	case 'b':
		result = (fileInfo.Mode()&fs.ModeDevice != 0) && (fileInfo.Mode()&fs.ModeCharDevice == 0)
	case 'c':
		result = fileInfo.Mode()&fs.ModeCharDevice != 0
	case 'd':
		result = fileInfo.IsDir()
	case 'f':
		result = fileInfo.Mode().IsRegular()
	case 'g':
		result = fileInfo.Mode()&fs.ModeSetgid != 0
	case 'h':
		result = fileInfo.Mode()&fs.ModeSymlink != 0
	case 'L':
		result = fileInfo.Mode()&fs.ModeSymlink != 0
	case 'k':
		result = fileInfo.Mode()&fs.ModeSticky != 0
	case 'p':
		result = fileInfo.Mode()&fs.ModeNamedPipe != 0
	case 'r':
		result = fileInfo.Mode()&0404 != 0
	case 's':
		result = fileInfo.Size() > 0
	case 'S':
		result = fileInfo.Mode()&fs.ModeSocket != 0
	case 'u':
		result = fileInfo.Mode()&fs.ModeSetuid != 0
	case 'w':
		result = fileInfo.Mode()&0202 != 0
	case 'x':
		result = fileInfo.Mode()&0101 != 0
	default:
		panic("Unknown operator: " + op)
	}

	return
}

// Shell variable substitution
// implemented: #, ##, %, %%, ^, ^^, ',', ',,', :, /
func Subst(str, pattern string) (result string) {
	strLen := len(str)
	patLen := len(pattern)
	if patLen == 0 {
		return str
	}

	// for '#'
	stripLeft := func(pat string) string {
		if strLen == 0 {
			return str
		}

		var i int
		for i = 0; i < strLen; i++ {
			if matched, _ := filepath.Match(pat, str[0:i]); matched {
				break
			}
		}

		return str[i:]
	}

	// for '##'
	stripLeftGreedy := func(pat string) string {
		if strLen == 0 {
			return str
		}

		var i int
		for i = strLen; i >= 0; i-- {
			if matched, _ := filepath.Match(pat, str[:i]); matched {
				break
			}

		}

		if i < 0 {
			// Not matched
			return str
		}

		return str[i:]
	}

	// for '%'
	stripRight := func(pat string) string {
		if strLen == 0 {
			return str
		}

		var i int
		for i = strLen; i >= 0; i-- {
			if matched, _ := filepath.Match(pat, str[i:]); matched {
				break
			}

		}

		if i < 0 {
			// Not matched
			return str
		}

		return str[:i]

	}

	// for '%%'
	stripRightGreedy := func(pat string) string {
		if strLen == 0 {
			return str
		}

		var i int
		for i = 0; i < strLen; i++ {
			if matched, _ := filepath.Match(pat, str[i:]); matched {
				break
			}
		}
		return str[:i]
	}

	// for '/'
	subst := func(pat, repl string) (replaced, left string) {
		var i int
		// find '*' in pattern
		for i = len(pat) - 1; i >= 0; i-- {
			if pat[i] == '*' {
				break
			}
		}

		foundGlob := (i != -1)

		var result strings.Builder
		var start int
		left = ""
		if foundGlob { // find match from end
			for start = 0; start < strLen; {
				for i = strLen; i >= start; i-- {
					if matched, _ := filepath.Match(pat, str[start:i]); matched {
						break
					}
				}

				if i < start { // no match found
					result.WriteByte(str[start])
					start++
				} else {
					// found match
					result.WriteString(repl)
					left = str[i:]
					break
				}
			}

		} else { // find match from beginning
			if len(pat) == 0 {
				for start = 0; start < strLen; start++ {
					result.WriteString(repl)
					result.WriteByte(str[start])
				}
			} else {
				for start = 0; start < strLen; start++ {
					for i = start + 1; i <= strLen; i++ {
						if matched, _ := filepath.Match(pat, str[start:i]); matched {
							start = i - 1
							break
						}
					}
					if i > strLen { // not matched
						result.WriteByte(str[start])
					} else {
						result.WriteString(repl)
						left = str[i:]
						break
					}
				}
			}

		}

		return result.String(), left

	}

	// for '//'
	substAll := func(pat, repl string) string {
		var result strings.Builder
		for {
			replaced, left := subst(pat, repl)
			result.WriteString(replaced)
			if left == "" {
				return result.String()
			}
			str = left
			strLen = len(left)
		}
	}

	switch pattern[0] {
	case '#':
		if patLen > 1 && pattern[1] == '#' {
			result = stripLeftGreedy(pattern[2:])
		} else {
			result = stripLeft(pattern[1:])
		}
	case '%':
		if patLen > 1 && pattern[1] == '%' {
			result = stripRightGreedy(pattern[2:])
		} else {
			result = stripRight(pattern[1:])
		}
	case '/':
		skip := 1
		if patLen > 1 && pattern[1] == '/' {
			skip = 2
		}

		var i int
		for i = 0; i < len(pattern); i++ {
			if i < skip {
				continue
			} // skip '//'
			if pattern[i] == '/' && pattern[i-1] != '\\' {
				break // found '/'
			}
		}

		j := i + 1
		if i == len(pattern) { // not found '/'
			j = i
		}

		if skip == 2 {
			result = substAll(pattern[skip:i], pattern[j:])
		} else {
			replaced, left := subst(pattern[skip:i], pattern[j:])
			result = replaced + left
		}

	case '^':
		if patLen > 1 && pattern[1] == '^' {
			result = strings.ToUpper(str)
		} else {
			result = (strings.ToUpper(str[:1]) + str[1:])
		}

	case ',':
		if patLen > 1 && pattern[1] == ',' {
			result = strings.ToLower(str)
		} else {
			result = (strings.ToLower(str[:1]) + str[1:])
		}

	case '-', '=':
		result = gog.If(str == "", pattern[1:], str)
	case '+':
		result = gog.If(str == "", str, pattern[1:])
	case '?':
		if str == "" {
			panic(pattern + ": Bad substitution")
		} else {
			result = str
		}
	case ':':
		if patLen == 1 {
			panic(pattern + ": Bad substitution")
		}

		// Find the second ':'
		var i int
		for i = 1; i < patLen; i++ {
			if pattern[i] == ':' {
				break
			}
		}

		var p1, p2 string
		p1 = pattern[1:i]
		p2 = pattern[i:]

		var start, end int
		if p1 == "" {
			start = 0
		} else {
			start = max(gog.Must(strconv.Atoi(p1)), 0)
			start = min(start, strLen)
		}

		if p2 == "" {
			end = strLen
		} else if p2 == ":" {
			end = start
		} else {
			l := gog.Must(strconv.Atoi(p2[1:]))
			end = gog.If(l < 0, strLen+l, start+l)
		}

		if end < start {
			panic(pattern + ": Bad substitution")
		}

		result = str[start:end]

	default:
		panic("Unknown pattern: " + pattern)
	}
	return
}
