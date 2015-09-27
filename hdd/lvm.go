package hdd

import (
	"os/exec"
	"strings"
	"strconv"
	"errors"
)



/*
cryptsetup luksOpen /dev/sda2 lvm
pvcreate /dev/mapper/lvm 								// /dev/sde2						// /dev/sde
vgcreate main /dev/mapper/lvm							// main /dev/sde2					// main /dev/sde
lvcreate -L 10GB -n root main
lvcreate -L 2GB -n swap main
lvcreate -l 100%FREE -n home main

pvmove /dev/sde2 /dev/sde3 //Wichtig LV muss in beiden angemeldet sein
 */



type LVM struct {}

type LogicalVolume struct {
	Path					string	//e.g. /dev/main/root
	Name 					string	//e.g. root
	VGName 					string	//e.g. main
	UUID 					string	//e.g. 5FqkvW-QGqp-2tfo-p0x0-NlgH-kVOP-WoOBZD
	Access 					string	//e.g. read/write
	CreationTime			string	//e.g. archiso, 2015-02-27 17:27:12 +0100
	Status 					string	//e.g. available
	LE 						string	//e.g. 2732
	Segments 				string	//e.g. 1
	Allocation 				string	//e.g. inherit
	BlockDevices			string	//e.g. 254:4

	SizeRead 				string  //e.g. 11,00GiB
	Size					string  //e.g. 123123123123

	CachePool				string
	CacheMeta				string

	SnapshotPath			[]string
}

type VolumeGroup struct {
	Name 					string	//e.g. main
	SystemID 				string	//e.g.
	Format 					string	//e.g. lvm2
	Areas 					int		//e.g. 1
	SequenceNumber			int		//e.g. 2
	Access 					string	//e.g. read/write
	Status 					string	//e.g. resizeable

	MAX_LV 					int		//e.g. 0
	CURRENT_LV 				int		//e.g. 1
	OPEN_LV 				int		//e.g. 1
	MAX_PV					int		//e.g. 0
	CURRENT_PV 				int		//e.g. 1
	ACT_PV 					int		//e.g. 1

	PE_SIZE 				string	//e.g. 4,00 MiB
	PE 						uint64	//e.g. 238466
	ALLOC_PE 				string	//e.g. 238466

	UUID 					string	//e.g. udAlBQ-K2BO-i0yT-mPDG-ZEse-ruGr-Tsg1pS

	SizeRead 				string  //e.g. 11,00GiB
	Size					string  //e.g. 123123123123
	Used 					string	//%TODO

	LogicalVolumeArray					[]LogicalVolume
	//%TODO array von physical volums

}

type PhysicalVolume struct {
	Name 					string	//e.g. root
	Path 					string	//e.g. /dev/mapper/root
	VGName 					string	//e.g. main
	Allocatable 			string	//e.g. yes

	PE_SIZE 				string	//e.g. 4,00MiB
	PE_TOTAL				string	//e.g. 7679
	PE_FREE 				string	//e.g. 0
	PE_ALLOCATED 			string	//e.g. 7679
	UUID 					string	//e.g. w57JSG-4kdT-LZY9-ceJN-M02M-52UP-KVH7fH

	SizeRead 				string  //e.g. 11,00GiB
	Size					string  //e.g. 123123123123

	VolumeGroup				VolumeGroup
}



func GetAllMountedLVM([]LVM, error){}	//erstmal lvm überdenken
func GetLVM([]LVM, error){}		//and so on

func CreateNewPhysicalVolumeFromString(dir string, cmd string) (PhysicalVolume, error){
	p, err := pvcreate(dir, cmd)
	if err != nil {
		log.Error(err.Error())
		return p, err
	}
	return p, nil
}

func CreateNewVolumeGroupFromString(name string, cmd string,p *PhysicalVolume) (VolumeGroup, error){
	v, err := vgcreate(name, cmd, p)
	if err != nil {
		log.Error(err.Error())
		return v, err
	}
	p.VolumeGroup = v

	return v, nil
}

func CreateNewLogicalVolumeFromString(name string, size string, cmd string, v *VolumeGroup) (LogicalVolume, error){
	l, err := lvcreate(name, size, cmd, v)
	if err != nil {
		log.Error(err.Error())
		return l, err
	}

	v.LogicalVolumeArray = append(v.LogicalVolumeArray, l)
	return l, nil
}



func (l *LogicalVolume) GetVolumeGroup() (VolumeGroup, error){
	v := VolumeGroup{}

	a, err :=GetVolumeGroupByString(l.VGName)
	if err != nil {
		return v, err
	}
	return a, nil
}

func (l *LogicalVolume) CreateSnapshot(name string, size string) (error){
	/*
		https://wiki.ubuntuusers.de/Logical_Volume_Manager
		sudo lvcreate --size 100M --snapshot --name <name> /dev/<group>/<volume>
	*/

	cmd := sudo + " lvcreate --snapshot --size " + size + " --name " + name + " " + l.Path
	log.Info(cmd)

	parts := strings.Fields(cmd)
	head := parts[0]
	parts = parts[1:len(parts)]

	_, err := exec.Command(head, parts...).Output()
	if err != nil {
		log.Error(err.Error())
		return err
	}

	l.SnapshotPath = append(l.SnapshotPath, "/dev/" + l.VGName + "/" + name)
	log.Info(l.SnapshotPath[len(l.SnapshotPath) - 1])
	return nil
}

func (l *LogicalVolume) MergeSnapshot(snapshotpath string) (error){
	/*
		https://wiki.ubuntuusers.de/Logical_Volume_Manager
		sudo lvcreate --size 100M --snapshot --name <name> /dev/<group>/<volume>
	*/

	cmd := sudo + "  --merge " + snapshotpath
	err := l.Convert(cmd)
	if err != nil {
		return err
	}

	return nil
}

func (l *LogicalVolume) DeleteSnapshot(snapshotpath string) (error){

	err := lvremove(snapshotpath)
	if err != nil {
		return err
	}

	for i,x := range l.SnapshotPath{
		if x == snapshotpath{
			l.SnapshotPath = append(l.SnapshotPath[:i], l.SnapshotPath[i+1:]...)
		}
	}

	return nil
}

func (l *LogicalVolume) CreateFileSystem(fs string) (error) {
	cmdfs := ""
	switch fs {
	case "swap":
		cmdfs = "mkfs."
	case "ntfs":
		cmdfs = "mkfs.ntfs"
	case "ext4":
		cmdfs = "mkfs.ext4"
	case "zfs":
		cmdfs = "mkfs.zfs"
	default:
		return errors.New("Please provide a valid FileSystem")
	}

	cmd := sudo + " " + cmdfs + " " + l.Path
	log.Info(cmd)

	parts := strings.Fields(cmd)
	head := parts[0]
	parts = parts[1:len(parts)]

	_, err := exec.Command(head, parts...).Output()
	if err != nil {
		log.Error(err.Error())
		return  err
	}


	return nil
}

func (l *LogicalVolume) Delete() (error) {
	err := lvremove(l.Path)
	if err != nil {
		return err
	}

	l = nil
	return nil
}

func (l *LogicalVolume) Change(cmdorg string) (error) {
	//%TODO is aber sack viel

	cmd := sudo + " lvchange " + cmdorg
	log.Info(cmd)

	parts := strings.Fields(cmd)
	head := parts[0]
	parts = parts[1:len(parts)]

	_, err := exec.Command(head, parts...).Output()
	if err != nil {
		log.Error(err.Error())
		return  err
	}
	return nil
}

func (l *LogicalVolume) Convert(cmdorg string) (error) {
	//%TODO is aber sack viel

	cmd := sudo + " lvconvert " + cmdorg
	log.Info(cmd)

	parts := strings.Fields(cmd)
	head := parts[0]
	parts = parts[1:len(parts)]

	_, err := exec.Command(head, parts...).Output()
	if err != nil {
		log.Error(err.Error())
		return  err
	}
	return nil
}

func (l *LogicalVolume) Extend(size string, p *PhysicalVolume) (error) {
/*
Extends the size of the logical volume "vg01/lvol10" by 54MiB on physical volume /dev/sdk3. This is only pos‐
sible if /dev/sdk3 is a member of volume group vg01 and there are enough free physical extents in it:

lvextend -L +54 /dev/vg01/lvol10 /dev/sdk3

Extends  the  size  of logical volume "vg01/lvol01" by the amount of free space on physical volume /dev/sdk3.
This is equivalent to specifying "-l +100%PVS" on the command line:

lvextend /dev/vg01/lvol01 /dev/sdk3

Extends a logical volume "vg01/lvol01" by 16MiB using physical  extents  /dev/sda:8-9  and  /dev/sdb:8-9  for
allocation of extents:

lvextend -L+16M vg01/lvol01 /dev/sda:8-9 /dev/sdb:8-9
*/
	cmd := ""
	if size == "" {
		cmd = sudo + " lvextend " + l.Path + " " + p.Path
	}else{
		cmd = sudo + " lvextend -L " + size + " " + l.Path + " " + p.Path
 	}

	log.Info(cmd)

	parts := strings.Fields(cmd)
	head := parts[0]
	parts = parts[1:len(parts)]

	_, err := exec.Command(head, parts...).Output()
	if err != nil {
		log.Error(err.Error())
		return err
	}


	return nil
}



func (v *VolumeGroup) CreateCachepool(size string, sizeMeta string, mode string, l *LogicalVolume) (error){
	/*
	http://man7.org/linux/man-pages/man7/lvmcache.7.html

	lvcreate -n OriginLV -L LargeSize VG SlowPVs
	lvcreate -n CacheDataLV -L CacheSize VG FastPVs
	lvcreate -n CacheMetaLV -L MetaSize VG FastPVs
	lvconvert --type cache-pool --poolmetadata VG/CacheMetaLV VG/CacheDataLV
	lvconvert --type cache --cachepool VG/CachePoolLV VG/OriginLV
	*/


	// mode : writethrough|writeback

	lCache, err := lvcreate(l.Name + "_cache", size, "", v)
	if err != nil {return err}

	lCacheMeta, err := lvcreate(l.Name + "_cacheMeta", sizeMeta , "", v)
	if err != nil {return err}

	cmd2 := " "
	if mode == "" {
		cmd2 = " "
	}else{
		if mode == "writethough" || mode == "writeback" {
			cmd2 = " --cachemode " + mode + " "
		}
	}

	//Todo -f
	cmd := "-f --type cache-pool --poolmetadata " + v.Name + "/" + lCacheMeta.Name + cmd2 + v.Name + "/" + lCache.Name
	err = lCacheMeta.Convert(cmd)
	if err != nil {return err}

	cmd = "-f --type cache --cachepool " + v.Name + "/" + lCache.Name + " " + v.Name + "/" + l.Name
	err = lCache.Convert(cmd)
	if err != nil {return err}

	l.CachePool = lCache.Name
	l.CacheMeta = lCacheMeta.Name

	return nil
}

func (v *VolumeGroup) RemoveCachepool(l *LogicalVolume) (error){
	/*
	http://man7.org/linux/man-pages/man7/lvmcache.7.html


	lvconvert --splitcache VG/CacheLV
    lvconvert --uncache VG/CacheLV
    lvremove VG/CacheLV
	*/



	err := lvremove(l.CachePool)
	if err != nil {
		return err
	}

	cmd := "--uncache " + l.Name
	l.Convert(cmd)

	return nil
}

func (v *VolumeGroup) Change(size string, sizeMeta string) {
	//TODO

}

func (v *VolumeGroup) Delete() (error) {
	err := vgremove(v.Name)
	if err != nil {
		return err
	}

	v = nil
	return nil
}

func (v *VolumeGroup) Extend(p *PhysicalVolume) (error) {
	if p.Path == "" {
		log.Error("You have to give a PhysicalVolume")
		return errors.New("You have to give a PhysicalVolume")
	}

	cmd := sudo + " vgextend " + v.Name + " " +  p.Path
	log.Info(cmd)

	parts := strings.Fields(cmd)
	head := parts[0]
	parts = parts[1:len(parts)]

	_, err := exec.Command(head, parts...).Output()
	if err != nil {
		log.Error(err.Error())
		return  err
	}


	return nil
}

func (v *VolumeGroup) CreateLogicalVolume(name string, size string) (error) {
	if v.Name == "" {
		log.Error("You have to give a VolumeGroup")
		return errors.New("You have to give a VolumeGroup")
	}

	_, err := CreateNewLogicalVolumeFromString(name, size, "", v)
	if err != nil {
		log.Error(err.Error())
		return  err
	}


	return nil
}

func (v *VolumeGroup) CreateLogicalVolumeWithCMD(name string, size string, cmd string) (error) {
	if v.Name == "" {
		log.Error("You have to give a VolumeGroup")
		return errors.New("You have to give a VolumeGroup")
	}

	_, err := CreateNewLogicalVolumeFromString(name, size, cmd, v)
	if err != nil {
		log.Error(err.Error())
		return  err
	}


	return nil
}

func (v *VolumeGroup) GetPhysicalVolume() (PhysicalVolume, error) {
	parray, err := pvdisplay()
	if err != nil {
		log.Error(err.Error())
		return PhysicalVolume{}, err
	}

	for _, x := range parray{
		if x.VolumeGroup.Name == v.Name{
			return x, nil
		}
	}

	return PhysicalVolume{}, errors.New("No right Physical Volume found")
}




func (p *PhysicalVolume) Check() (error) {
	//%TODO
	return nil
}

func (p *PhysicalVolume) Resize() (error) {
	//%TODO
	return nil
}

func (p *PhysicalVolume) Delete() (error) {
	if p.VGName != "" {
		return errors.New("Please delet first the Volume Group" + p.VGName)
	}
	err := pvremove(p.Path)
	if err != nil {
		return err
	}

	p = nil
	return nil
}

func (p *PhysicalVolume) Change() (error) {
	//%TODO
	return nil
}
//Moves all LV p contains
func (p *PhysicalVolume) Move(pp *PhysicalVolume) (error) {
/*
To  move  all  Physical Extents that are used by simple Logical Volumes on /dev/sdb1 to free Physical Extents
elsewhere in the Volume Group use:

pvmove /dev/sdb1

Additionally, a specific destination device /dev/sdc1 can be specified like this:

pvmove /dev/sdb1 /dev/sdc1

To perform the action only on extents belonging to the single Logical Volume lvol1 do this:

pvmove -n lvol1 /dev/sdb1 /dev/sdc1

Rather than moving the contents of the entire device, it is possible to move a range of  Physical  Extents  -
for example numbers 1000 to 1999 inclusive on /dev/sdb1 - like this:

pvmove /dev/sdb1:1000-1999

A range can also be specified as start+length, so

pvmove /dev/sdb1:1000+1000

also  refers to 1000 Physical Extents starting from Physical Extent number 1000.  (Counting starts from 0, so
this refers to the 1001st to the 2000th inclusive.)

To move a range of Physical Extents to a specific location (which must have sufficient free extents) use  the
form:

pvmove /dev/sdb1:1000-1999 /dev/sdc1

or

pvmove /dev/sdb1:1000-1999 /dev/sdc1:0-999

If  the  source  and  destination  are on the same disk, the anywhere allocation policy would be needed, like
this:

pvmove --alloc anywhere /dev/sdb1:1000-1999 /dev/sdb1:0-999

The part of a specific Logical Volume present within in a range of Physical Extents can also  be  picked  out
and moved, like this:

pvmove -n lvol1 /dev/sdb1:1000-1999 /dev/sdc1
*/

	if pp.Path == "" {
		log.Error("You have to give a PhysicalVolume")
		return errors.New("You have to give a PhysicalVolume")
	}

	if pp.VGName != p.VGName {
		log.Error("Both PhysicalVolumes need the same Volumegroup name")
		return errors.New("Both PhysicalVolumes need the same Volumegroup name")
	}

	cmd := sudo + " pvmove " + p.Path + " " +  pp.Path
	log.Info(cmd)

	parts := strings.Fields(cmd)
	head := parts[0]
	parts = parts[1:len(parts)]

	ret, err := exec.Command(head, parts...).Output()
	if err != nil {
		log.Error(err.Error() + string(ret))
		return  err
	}


	return nil
}

func (p *PhysicalVolume) CreateVolumeGroup(name string) (error) {
	if p.Path == "" {
		log.Error("You have to give a PhysicalVolume")
		return errors.New("You have to give a PhysicalVolume")
	}

	_, err := CreateNewVolumeGroupFromString(name, "", p)
	if err != nil {
		log.Error(err.Error())
		return  err
	}

	return nil
}

func (p *PhysicalVolume) CreateVolumeGroupWithCMD(name string, cmd string) (error) {
	if p.Path == "" {
		log.Error("You have to give a PhysicalVolume")
		return errors.New("You have to give a PhysicalVolume")
	}

	_, err := CreateNewVolumeGroupFromString(name, cmd, p)
	if err != nil {
		log.Error(err.Error())
		return  err
	}

	return nil
}

func GetPhysicalVolumeByString(path string)(PhysicalVolume, error){
	parray, err := pvdisplay()
	if err != nil {
		log.Error(err.Error())
		return PhysicalVolume{}, err
	}

	for _, x := range parray{
		if x.Path == path{
			return x, nil
		}
	}

	return PhysicalVolume{}, errors.New("No right Physical Volume found")
}

func GetVolumeGroupByString(name string)(VolumeGroup, error){
	varray, err := vgdisplay()
	if err != nil {
		log.Error(err.Error())
		return VolumeGroup{}, err
	}

	for _, x := range varray{
		if x.Name == name{
			return x, nil
		}
	}

	return VolumeGroup{}, errors.New("No right Physical Volume found")
}

func GetLogicalVolumeByString(path string)(LogicalVolume, error){
	larray, err := lvdisplay()
	if err != nil {
		log.Error(err.Error())
		return LogicalVolume{}, err
	}

	for _, x := range larray{
		if x.Path == path{
			return x, nil
		}
	}

	return LogicalVolume{}, errors.New("No right Physical Volume found")
}




//======================================================================
func pvremove(path string) (error){//%TODO :Dir = kann part aber auch lukscontainer sein ==> wrapper für beides schreiben
	if path == "" {
		log.Error("You have to give a Name")
		return errors.New("You have to give a Name")
	}
	/*ret, err := isStringValidHDDorPart(Dir)
	if ret == false {
		println("Error: string is not Valid Hardrive or Part")
		return p, err
	}*/

	cmd := sudo + " pvremove " + path
	log.Info(cmd)

	parts := strings.Fields(cmd)
	head := parts[0]
	parts = parts[1:len(parts)]
	_, err := exec.Command(head, parts...).Output()
	if err != nil {
		log.Error(err.Error())
		return err
	}

	return nil
}

func vgremove(name string) (error){
	if name == "" {
		log.Error("You have to give a Name")
		return errors.New("You have to give a Name")
	}

	//%TODO
	/*ret, err := isStringValidHDDorPart(Dir)
	if ret == false {
		println("Error: string is not Valid Hardrive or Part")
		return p, err
	}*/

	cmd := sudo + " vgremove " + name
	log.Info(cmd)

	parts := strings.Fields(cmd)
	head := parts[0]
	parts = parts[1:len(parts)]
	_, err := exec.Command(head, parts...).Output()
	if err != nil {
		log.Error(err.Error())
		return err
	}

	return nil
}

func lvremove(path string) (error){
	if path == "" {
		log.Error("You have to give a Name")
		return errors.New("You have to give a Name")
	}
	/*ret, err := isStringValidHDDorPart(Dir)
	if ret == false {
		println("Error: string is not Valid Hardrive or Part")
		return p, err
	}*/

	cmd := sudo + " lvremove -f " + path
	log.Info(cmd)

	parts := strings.Fields(cmd)
	head := parts[0]
	parts = parts[1:len(parts)]
	_, err := exec.Command(head, parts...).Output()
	if err != nil {
		log.Error(err.Error())
		return err
	}

	return nil
}

func lvcreate(name string, size string, cmdd string,v *VolumeGroup) (LogicalVolume, error){
	l := LogicalVolume{}
	if name == "" {
		log.Error("You have to give a LV Name")
		return l, errors.New("You have to give a LV Name")
	}
	if v.Name == "" {
		log.Error("You have to give a VG")
		return l, errors.New("You have to give a VG")
	}

	//%TODO parse size

	cmd2 := " -L"
	if size[:len(size)-2] == "EE"{//100%FREE
		cmd2 = " -l"
	}

	cmd := sudo + " lvcreate " + cmdd + " " +  cmd2 + " " + size + " -n " + name + " " + v.Name
	log.Info(cmd)

	parts := strings.Fields(cmd)
	head := parts[0]
	parts = parts[1:len(parts)]

	_, err := exec.Command(head, parts...).Output()
	if err != nil {
		log.Error(err.Error())
		return l, err
	}

	array, err := lvdisplay()
	if err != nil {
		log.Error(err.Error())
		return l, err
	}

	for _, a := range array{
		if a.Name == name{
			l = a
		}
	}
	if l.Access == "" {
		log.Error("Something went wrong, Couldnt find new LV")
		return l, errors.New("Something went wrong, Couldnt find new LV")

	}
	return l, nil
}

func vgcreate(name string, cmdd string, p *PhysicalVolume) (VolumeGroup, error){
	v := VolumeGroup{}
	if name == "" {
		log.Error("You have to give a Name")
		return v, errors.New("You have to give a Name")
	}
	if p.Path == "" {
		log.Error("You have to give a PhysicalVolume")
		return v, errors.New("You have to give a PhysicalVolume")
	}

	cmd := sudo + " vgcreate " + cmdd + " "+ name + " " +  p.Path
	log.Info(cmd)

	parts := strings.Fields(cmd)
	head := parts[0]
	parts = parts[1:len(parts)]

	_, err := exec.Command(head, parts...).Output()
	if err != nil {
		log.Error(err.Error())
		return v, err
	}

	array, err := vgdisplay()
	if err != nil {
		log.Error(err.Error())
		return v, err
	}

	for _, a := range array{
		if a.Name == name{
			v = a
		}
	}
	if v.UUID == "" {
		log.Error("Something went wrong, Couldnt find new VG")
		return v, errors.New("Something went wrong, Couldnt find new VG")
	}

	return v, nil
}

func pvcreate(dir string, cmdd string,) (PhysicalVolume, error){
	p := PhysicalVolume{}
	if dir == "" {
		log.Error("You have to give a Dir")
		return p, errors.New("You have to give a Dir")
	}
	/*ret, err := isStringValidHDDorPart(Dir)
	if ret == false {
		println("Error: string is not Valid Hardrive or Part")
		return p, err
	}*/

	cmd := sudo + " pvcreate " + cmdd + " " + dir
	log.Info(cmd)

	parts := strings.Fields(cmd)
	head := parts[0]
	parts = parts[1:len(parts)]

	_, err := exec.Command(head, parts...).Output()
	if err != nil {
		log.Error(err.Error())
		return p, err
	}

	array, err := pvdisplay()
	if err != nil {
		log.Error(err.Error())
		return p, err
	}

	for _, a := range array{
		if a.Path == dir{
			p = a
		}
	}
	if p.UUID == "" {
		log.Error("Something went wrong, Couldnt find new PV")
		return p, errors.New("Something went wrong, Couldnt find new PV")

	}
	return p, nil
}

func vgextend(vv *VolumeGroup, HDDDir string) (error){
	if HDDDir == "" { return errors.New("You have to give a Name")}
	if vv.Name == "" { return errors.New("You have to give a VolumeGroup Name")}

	_, err := exec.Command(sudo, "vgextend", vv.Name, HDDDir).Output()
	if err != nil {
		println("Error: " + err.Error())
		return err
	}

	return  nil
}

func pvdisplay()([]PhysicalVolume, error){
	Array := []PhysicalVolume{}
	log.Info("sudo pvdisplay")

	data, err := exec.Command(sudo, "pvdisplay").Output()
	if err != nil {
		log.Error(err.Error())
		return Array, err
	}

	out := string(data)
	lines := strings.Split(out, "\n")

	l := PhysicalVolume{}

	for i := 0; i < (len(lines) ); i++ {
		if lines[i] == "" {break;}
		if (lines[i][2:3] == "-"){continue;}
		if (lines[i][2:3] == " "){//New Block
			l = PhysicalVolume{}
			continue;
		}

		ind := strings.LastIndex(lines[i + 0][24:], "/")
		l.Name 				= lines[i + 0][24 + ind + 1:]
		l.Path 				= lines[i + 0][24:]
		l.VGName 			= lines[i + 1][24:]
		l.SizeRead 			= lines[i + 2][24:]
		l.Allocatable 		= lines[i + 3][24:]
		l.PE_SIZE 			= lines[i + 4][24:]
		l.PE_TOTAL 			= lines[i + 5][24:]
		l.PE_FREE 			= lines[i + 6][24:]
		l.PE_ALLOCATED 		= lines[i + 7][24:]
		l.UUID 				= lines[i + 8][24:]

		/*c 					:= lines[i +16][24:]
		cc 				   := strings.Index(c, "/")
		l.Allocatable		=lines[i +16][24:cc+24]*/

		i += 10
		Array = append(Array, l)
	}

	//fmt.Println(Array)
	return Array, nil
}//%TODO wrap von "GetAllLogicalVolumes

func vgdisplay()([]VolumeGroup, error){
	Array := []VolumeGroup{}
	log.Info("sudo vgdisplay")

	data, err := exec.Command(sudo, "vgdisplay").Output()
	if err != nil {
		log.Error(err.Error())
		return Array, err
	}

	out := string(data)
	lines := strings.Split(out, "\n")

	l := VolumeGroup{}

	for i := 0; i <= (len(lines)); i++ {
		if lines[i] == "" {break;}
		if (lines[i][2:3] == "-"){continue;}
		if (lines[i][2:3] == " "){//New Block
			l = VolumeGroup{}
			continue;
		}

		l.Name 				= lines[i + 0][24:]
		l.SystemID 			= lines[i + 1][24:]
		l.Format 			= lines[i + 2][24:]
		l.Areas, _ 		   	= strconv.Atoi(lines[i + 3][24:])
		l.SequenceNumber, _ = strconv.Atoi(lines[i + 4][24:])
		l.Access 			= lines[i + 5][24:]
		l.Status 			= lines[i + 6][24:]
		l.MAX_LV, _ 		= strconv.Atoi(lines[i + 7][24:])
		l.CURRENT_LV, _		= strconv.Atoi(lines[i + 8][24:])
		l.OPEN_LV	, _		= strconv.Atoi(lines[i + 9][24:])
		l.MAX_PV, _			= strconv.Atoi(lines[i +10][24:])
		l.CURRENT_PV, _	    = strconv.Atoi(lines[i +11][24:])
		l.ACT_PV, _		 	= strconv.Atoi(lines[i +12][24:])
		l.SizeRead 			= lines[i +13][24:]
		l.PE_SIZE 			= lines[i +14][24:]
		l.PE, _ 			= strconv.ParseUint(lines[i +15][24:], 0, 64)
		l.UUID 				= lines[i +18][24:]

		c 					:= lines[i +16][24:]
		cc 				   := strings.Index(c, "/")
		l.ALLOC_PE		=lines[i +16][24:cc+24]

		i += 18
		Array = append(Array, l)
	}

	//fmt.Println(Array)
	return Array, nil
}

func lvdisplay()([]LogicalVolume, error){
	Array := []LogicalVolume{}
	log.Info("sudo lvdisplay")

	data, err := exec.Command(sudo, "lvdisplay").Output()
	if err != nil {
		log.Error(err.Error())
		return Array, err
	}

	out := string(data)
	lines := strings.Split(out, "\n")

	l := LogicalVolume{}

	for i := 0; i <= (len(lines) - 11); i++ {
		if lines[i] == "" {break;}
		if (lines[i][2:3] == "-"){continue;}
		if (lines[i][2:3] == " "){//New Block
			l = LogicalVolume{}
			continue;
		}

		l.Path 			= lines[i + 0][25:]
		l.Name 			= lines[i + 1][25:]
		l.VGName 		= lines[i + 2][25:]
		l.UUID 			= lines[i + 3][25:]
		l.Access		= lines[i + 4][25:]
		l.CreationTime  = lines[i + 5][25:]
		l.Status 		= lines[i + 6][25:]
		l.SizeRead 		= lines[i + 7][25:]
		l.LE 			= lines[i + 8][25:]
		l.Segments 		= lines[i + 9][25:]
		l.Allocation 	= lines[i +10][25:]
		l.BlockDevices 	= lines[i +13][25:]

		//println(lines[i][25:])
		i += 14
		Array = append(Array, l)
	}

	//fmt.Println(Array)
	return Array, nil
}



