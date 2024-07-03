package sh

import (
	"os"
	"strings"
	"testing"
)

func TestPrintf(t *testing.T) {
	runTest := func(exp string, format string, args ...any) {
		var res string
		res, _ = Printf(format, args...)
		if  res != exp {
			t.Errorf("Failed on Printf %s %q, expect: '%v', got: '%v'", format, args, exp, res)
		}

	}

	runTest("", "")
	runTest("", "", "hello")
	runTest("hello", "hello")
	runTest("helloworld", "%s", "hello", "world")
	runTest("hello%!s(MISSING)", "%s%s", "hello")
	runTest("hello%", "%s%%", "hello")
	runTest("helloworld\nworldhello\n", "%s%s\n", "hello", "world", "world", "hello")
}

func TestCommand(t *testing.T) {
	cmds := []string{"-v", "cp", "rm"} // unix/msys command 'cp', 'rm'
	var out []byte
	var err error
	var msg string
	var exp string

	out, err = Command(cmds...)
	msg = string(out)
	if err != nil || strings.Count(string(out), "\n") != 2 {
		t.Errorf("Failed command%v: %v, %v", cmds, msg, err)
	}

	cmds = []string{"-p", "hostname"} // run command: hostname
	out, err = Command(cmds...)

	var err1 error
	exp, err1 = os.Hostname()
	if err1 != nil {
		t.Fatalf("Exception: Failed to get hostname via os.Hostname(): %v", err1)
	}

	msg = strings.TrimSpace(string(out))
	if err != nil || msg != exp {
		t.Errorf("Failed on Command%v:\noutput: %v\nexpected: %v\nerror: %v", cmds, msg, exp, err)
	}

	cmds = []string{"-v", "./nonexist"}
	_, err = Command(cmds...)
	if err == nil {
		t.Errorf("Failed on Command%v", cmds)
	}
}

func TestTest(t *testing.T) {
	runTest := func(args ...string) {
		if !Test(args...) {
			t.Errorf("Failed on Test%v\n", args)
		}
	}

	runTestPanic := func(args ...string) {
		defer func() {
			if r := recover(); r == nil {
				t.Errorf("Failed on Test%v\n", args)
			}
		}()
		runTest(args...)
	}

	runTest("-a", ".")
	runTest("!", "-b", "sh_test.go")
	runTest("!", "-c", "sh_test.go")
	runTest("!", "-f", "non-exists")
	runTest("-d", ".")
	runTest("-f", "sh_test.go")
	runTest("-s", "sh_test.go")
	runTest("-r", "sh_test.go")
	runTest("-w", "sh_test.go")
	runTest("!", "-x", "sh_test.go")
	runTest("!", "-S", "sh_test.go")
	runTest("!", "-h", "sh_test.go")
	runTest("!", "-p", "sh_test.go")
	runTest("!", "-k", "sh_test.go")
	runTest("!", "-L", "sh_test.go")
	runTest("!", "-g", "sh_test.go")
	runTest("!", "-u", "sh_test.go")
	runTestPanic("-Z", "sh_test.go")
	runTestPanic("!", "b", "-c")
}

func TestSubst(t *testing.T) {
	runTest := func(tar, pat, exp string) {
		act := Subst(tar, pat)
		if exp != act {
			t.Errorf("Failed Subst(\"%v\", \"%v\") => \"%v\" != \"%v\"", tar, pat, act, exp)
		}
	}

	runTestPanic := func(args ...string) {
		defer func() {
			if r := recover(); r == nil {
				t.Errorf("Failed on Test%v\n", args)
			}
		}()
		runTest(args[0], args[1], args[2])
	}

	runTest("", "", "")
	runTest("", "#a", "")
	runTest("", "%a", "")
	runTest("", "##a", "")
	runTest("", "%%a", "")
	runTest("hello", "", "hello")
	runTest("hello", "#", "hello")
	runTest("hello", "##", "hello")
	runTest("hello", "#*", "hello")
	runTest("hello", "##*", "")
	runTest("hello", "#hel", "lo")
	runTest("hello", "#h*l", "lo")
	runTest("hello", "##h*l", "o")

	runTest("hello", "%", "hello")
	runTest("hello", "%%", "hello")
	runTest("hello", "%*", "hello")
	runTest("hello", "%%*", "")
	runTest("hello", "%l*o", "hel")
	runTest("hello", "%%l*o", "he")
	runTest("hello", "%?o", "hel")
	runTest("hello", "%l\\*o", "hello")

	runTest("hello", "^", "Hello")
	runTest("hello", "^^", "HELLO")

	runTest("HELLO", ",", "hELLO")
	runTest("HELLO", ",,", "hello")

	runTest("hello", ":1", "ello")
	runTest("hello", ":10", "")
	runTest("hello", ":-1", "hello")
	runTest("hello", ":0", "hello")

	runTest("hello", ":2:", "")
	runTest("hello", ":2:0", "")
	runTest("hello", ":2:1", "l")
	runTest("hello", ":2:-1", "ll")
	runTestPanic("hello", ":2:-10", "")

	runTest("hello", "/l*l", "heo")
	runTest("hello", "/*l", "o")
	runTest("hello", "/l*", "he")
	runTest("hello", "/l/x", "hexlo")
	runTest("hello", "/ll/LL", "heLLo")
	runTest("hello", "/e", "hllo")
	runTest("hello", "/*", "")
	runTest("hello", "/l*l/xx", "hexxo")
	runTest("hello", "/?/x", "xello")
	runTest("hello", "//l/x", "hexxo")
	runTest("hello", "//l*", "he")
	runTest("hello", "//x", "hello")
	runTest("hello", "///a", "ahaealalao")
	runTest("hello", "///", "hello")
	runTest("hello", "//*", "")

	runTestPanic("hello", ".", "")
}
