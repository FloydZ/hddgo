package main

import (
	"os"
	"github.com/op/go-logging"
	"website2/app/models/hddtest/hdd"
	"fmt"
)

var sudo = "sudo"
var lsblkcmd = "+UUID,OWNER,GROUP,MODE,STATE,HOTPLUG,FSTYPE,LABEL,MODEL,SERIAL"

var log = logging.MustGetLogger("")
var format = logging.MustStringFormatter("%{color}%{time:15:04:05.000} %{module} %{shortfunc} ▶ %{level:.4s} %{id:03x}%{color:reset} %{message}", )


func main(){
	backend1 := logging.NewLogBackend(os.Stderr, "", 0)
	backend1Leveled := logging.AddModuleLevel(backend1)
	backend1Leveled.SetLevel(logging.ERROR, "")
	backend1Formatter := logging.NewBackendFormatter(backend1, format)
	logging.SetBackend(backend1Leveled, backend1Formatter)

	p, _ := hdd.CreateNewPhysicalVolumeFromString("/dev/sde4", "")
	v, _ := hdd.CreateNewVolumeGroupFromString("EineVG", "", &p)
	v.CreateLogicalVolume("testlv", "500MB")
	v.CreateCachepool("500MB", "12MB", "", &v.LogicalVolumeArray[0])
	fmt.Println(v.LogicalVolumeArray[0].CachePool)


}

