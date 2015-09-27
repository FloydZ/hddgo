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
	//lvdisplay()
	//vgdisplay()
	//pvdisplay()

	/*
	p, err := hdd.GetPartitionFromString("/dev/sde2")
	if err != nil {println("Error: " + err.Error())}

	l := &hdd.LUKS{PhysicalVolumeName:"Test2", Part:p, KeyFile:"/home/pr0gramming/xen.cfg"}

	l.LuksFormat()
	l.LuksOpen()
	//l.luksErase()
	//l.luksChangeKey("/home/pr0gramming/start.sh")
	//l.luksSuspend()
	//l.luksAddKey("/home/pr0gramming/OpenVPN_RUB.ovpn")
	//l.luksErase()



	//log.Info(cryptsetupIsInstalled())


	pv, err := hdd.pvcreate("/dev/mapper/" + l.PhysicalVolumeName)
	if err != nil {println("Error: " + err.Error())}

	vg, err := hdd.vgcreate("test1", &pv)
	if err != nil {println("Error: " + err.Error())}

	lv, err := hdd.lvcreate("root", "5G", &vg)
	if err != nil {println("Error: " + err.Error())}

	fmt.Println(lv)

	lvremove("/dev/mapper/test1-root")
	vgremove("test1")
	pvremove("/dev/mapper/Test2")


	l.LuksErase()

*/
	/*h, _ := hdd.GetHarddriveFromString("/dev/sde")
	p, _ := hdd.CreateNewPhysicalVolumeFromString("/dev/sde3")
	p2, _ := hdd.CreateNewPhysicalVolumeFromString("/dev/sde4")
	v, _ := hdd.CreateNewVolumeGroupFromString("EineVG", &p)
	hdd.CreateNewLogicalVolumeFromString("main", "10MB", &v)
	v.Extend(&p2)

	p.Move(&p2)
	*/

	/*
	p, _ := hdd.CreateNewPhysicalVolumeFromString("/dev/sde3")
	p.CreateVolumeGroup("TestVG")
	p.VolumeGroup.CreateLogicalVolume("testvg", "1GB")
	p.VolumeGroup.LogicalVolumeArray[0].CreateFileSystem("ext4")
	*/

	//p, _ := hdd.GetPartitionFromString("/dev/sde2")
	//p.CreateFileSystem("ext4")

	//p, _ := hdd.GetPhysicalVolumeByString("/dev/sde3")
	//v, _ := hdd.GetVolumeGroupByString("TestVG")


	//pp, _ := hdd.GetPartitionFromString("/dev/sde2")
	//fmt.Println(pp)
	//pp.SetLabel("test")


	p, _ := hdd.CreateNewPhysicalVolumeFromString("/dev/sde4", "")
	v, _ := hdd.CreateNewVolumeGroupFromString("EineVG", "", &p)
	v.CreateLogicalVolume("testlv", "500MB")
	v.CreateCachepool("500MB", "12MB", "", &v.LogicalVolumeArray[0])
	fmt.Println(v.LogicalVolumeArray[0].CachePool)


}

