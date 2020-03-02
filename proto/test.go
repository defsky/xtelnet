package proto

import (
	"io"
	"os"
)

func Daemon(cmdstr string) {
	// cmd := exec.Command(cmdstr)
	// cmd.Stdin = nil
	// cmd.Stdout = nil
	// cmd.Stderr = nil
	// cmd.SysProcAttr = &syscall.SysProcAttr{Setsid: true}

	// err := cmd.Start()
	// if err == nil {
	// 	cmd.Process.Release()
	// 	os.Exit(0)
	// }
}

func daemonize(cmd string, args []string, pipe io.WriteCloser) error {
	// pid, _, sysErr := syscall.RawSyscall(syscall.SYS_FORK, 0, 0, 0)
	// if sysErr != 0 {
	// 	return fmt.Errorf("fail to call fork")
	// }
	// if pid > 0 {
	// 	if _, err := syscall.Wait4(int(pid), nil, 0, nil); err != nil {
	// 		return fmt.Errorf("fail to wait for child process: %v", err)
	// 	}
	// 	return nil
	// } else if pid < 0 {
	// 	return fmt.Errorf("child id is incorrect")
	// }

	// ret, err := syscall.Setsid()
	// if err != nil || ret < 0 {
	// 	return fmt.Errorf("fail to call setsid")
	// }

	// signal.Ignore(syscall.SIGHUP)
	// syscall.Umask(0)

	// nullFile, err := os.Open(os.DevNull)
	// if err != nil {
	// 	return fmt.Errorf("fail to open os.DevNull: %v", err)
	// }
	// files := []*os.File{
	// 	nullFile, // (0) stdin
	// 	nullFile, // (1) stdout
	// 	nullFile, // (2) stderr
	// }
	// attr := &os.ProcAttr{
	// 	Dir:   "/",
	// 	Env:   os.Environ(),
	// 	Files: files,
	// }
	// child, err := os.StartProcess(cmd, args, attr)
	// if err != nil {
	// 	return fmt.Errorf("fail to start process: %v", err)
	// }

	// buff := make([]byte, 4)
	// binary.BigEndian.PutUint32(buff[:], uint32(child.Pid))
	// if n, err := pipe.Write(buff); err != nil || n != 4 {
	// 	return fmt.Errorf("fail to write back the pid")
	// }

	os.Exit(0)
	return nil
}
