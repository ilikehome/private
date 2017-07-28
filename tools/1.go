package main

import (
	"fmt"
	"os"
	"flag"
	"strconv"
	"strings"
	"os/exec"
	"path"
	"io/ioutil"
)

const (
	PID_CMD          = "/usr/bin/ps -ef|/usr/bin/grep '%s' |/usr/bin/grep -v grep|/usr/bin/awk '{print $2}'"
	CGROUP_MEM_MOUNT = "/sys/fs/cgroup/memory/nba"
	CGROUP_CPU_MOUNT = "/sys/fs/cgroup/cpu/nba"
)

func main() {
	if isContain(os.Args, "show"){
		show()
		return
	}
	catalog, cpu, mem, greps := getParameter()

	os.Mkdir(CGROUP_MEM_MOUNT, 0755)
	os.Mkdir(CGROUP_CPU_MOUNT, 0755)
	pid := getPid(greps)
	if pid == -1{
		fmt.Printf("Progress is not found.\n")
		return
	}

	memValue := mem
	write(pid, catalog, int(cpu*100000), memValue)

	cpuInfo := strconv.FormatFloat(cpu, 'f', -1, 64)
	if cpu <=0{
		cpuInfo = "Unlimited"
	}
	memInfo := mem
	if mem ==""{
		memInfo = "Unlimited"
	}

	fmt.Printf("Set PID(%d), MaxMem=%s, MaxCpu=%s Success.\n", pid, memInfo, cpuInfo)
}

func show(){
	dir, _ := ioutil.ReadDir(CGROUP_MEM_MOUNT)
	for _, fi := range dir {
		if fi.IsDir() {
			continue
		}
	}
}

func isContain(arr []string, ele string) bool{
	for _,x:= range arr{
		if x==ele{
			return true
		}
	}
	return false
}

func getPid(greps string) (pid int){
	pidGrepCmd := fmt.Sprintf(PID_CMD, greps)
	cmd := exec.Command("/bin/sh", "-c", pidGrepCmd)
	buf, err := cmd.Output()
	if err == nil{
		return readFirstLine(buf)
	}else{
		fmt.Printf("%v", err)
		return -1
	}
}

func readFirstLine(buf []byte) (pid int){
	inx := len(buf)
	for n,x:= range buf{
		if x=='\n'{
			inx = n
			break
		}
	}
	firstLine := string(buf[:inx])
	if pid,err := strconv.Atoi(firstLine); err==nil{
		return pid
	}
	return -1
}

func write(pid int, catalog string, cpu int, mem string){
	if cpu>=0{
		cpuLimit := strconv.Itoa(cpu)
		os.Mkdir(path.Join(CGROUP_CPU_MOUNT, catalog), 0755)
		ioutil.WriteFile(path.Join(CGROUP_CPU_MOUNT, catalog, "tasks") , []byte(strconv.Itoa(pid)), 0644)
		ioutil.WriteFile(path.Join(CGROUP_CPU_MOUNT, catalog, "cpu.cfs_quota_us") , []byte(cpuLimit), 0644)
	}

	if mem != ""{
		os.Mkdir(path.Join(CGROUP_MEM_MOUNT, catalog), 0755)
		ioutil.WriteFile(path.Join(CGROUP_MEM_MOUNT, catalog, "tasks") , []byte(strconv.Itoa(pid)), 0644)
		ioutil.WriteFile(path.Join(CGROUP_MEM_MOUNT, catalog, "memory.limit_in_bytes") , []byte(mem), 0644)
	}
}

func getParameter() (catalog string, cpuValue float64, memValue string, cmd string){
	catalog = *flag.String("catalog", "", "Input your limit type")
	cpu := flag.Float64("cpu", 0, "Input your cpu number")
	mem := flag.String("mem", "", "Input your memory number")
	flag.Parse()
	cpuValue = *cpu
	memValue = *mem

	args := os.Args[1:]
	cmdLines := []string{}
	for _,x := range args{
		if !(strings.HasPrefix(x, "-cpu") || strings.HasPrefix(x, "-mem")){
			cmdLines = append(cmdLines, x)
		}
	}
	cmd = strings.Join(cmdLines, " ")
	return
}