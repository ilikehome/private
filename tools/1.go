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
	PID_CMD          = "/bin/ps -ef|grep '%s' |/bin/grep -v grep|/bin/awk '{print $2}'"
	CGROUP_MEM_MOUNT = "/sys/fs/cgroup/memory/nba"
	CGROUP_CPU_MOUNT = "/sys/fs/cgroup/cpu/nba"
)

func main() {
	if isContain(os.Args, "show"){
		show()
		return
	}
	catalog, cpu, memValue, memUnit, greps := getParameter()
	pid := getPid(greps)
	if pid == -1{
		fmt.Printf("Progress is not found.\n")
		return
	}
	write(pid, catalog, cpu, memValue, memUnit)
	fmt.Println(cpu, memValue, memUnit, pid)
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
	if buf, err := cmd.Output(); err == nil{
		if pid,err := strconv.Atoi(string(buf[:len(buf)-1])); err==nil{
			return pid
		}
	}
	return -1
}

func write(pid int, catalog string, cpu int, memValue int, memUnit string){
	cpuLimit := "Unlimited"
	if cpu!=0{
		os.Mkdir(path.Join(CGROUP_CPU_MOUNT, catalog), 0755)
		ioutil.WriteFile(path.Join(CGROUP_CPU_MOUNT, catalog, "tasks") , []byte(strconv.Itoa(pid)), 0644)
		ioutil.WriteFile(path.Join(CGROUP_CPU_MOUNT, catalog, "memory.limit_in_bytes") , []byte("100m"), 0644)
		cpuLimit = strconv.Itoa(cpu)
	}
	memLimit := "Unlimited"
	if memValue!=0{
		os.Mkdir(path.Join(CGROUP_MEM_MOUNT, catalog), 0755)
		ioutil.WriteFile(path.Join(CGROUP_MEM_MOUNT, catalog, "tasks") , []byte(strconv.Itoa(pid)), 0644)
		ioutil.WriteFile(path.Join(CGROUP_MEM_MOUNT, catalog, "memory.limit_in_bytes") , []byte("100m"), 0644)
		memLimit = strconv.Itoa(memValue)+ memUnit
	}
	fmt.Printf("Set PID(%d), MaxMem=%s, MaxCpu=%d Success.\n", pid, memLimit, cpuLimit)
}

func getParameter() (catalog string, cpuValue int, memValue int, memUint string, cmd string){
	catalog = *flag.String("catalog", "", "Input your limit type")
	cpu := flag.Int("cpu", 0, "Input your cpu number")
	mem := flag.String("mem", "", "Input your memory number")
	flag.Parse()
	cpuValue = *cpu

	memString := []rune(*mem)
	memValue,_ = strconv.Atoi(string(memString[0:len(memString)-2]))
	if strings.HasSuffix(strings.ToUpper(*mem), "MB"){
		memUint = "MB"
	}else if strings.HasSuffix(strings.ToUpper(*mem), "GB"){
		memUint = "GB"
	}

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