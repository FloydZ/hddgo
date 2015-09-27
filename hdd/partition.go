package hdd

import (
	"strings"
	"errors"
	"os/exec"
	linuxproc "github.com/c9s/goprocinfo/linux"
	"unicode"
)
type Partition struct {
	Name			string			//e.g. sda1
	Path 			string			//e.g. /dev/sda1
	Number 			string			//e.g. 1
	Type 			string			//e.g. disk, part, crypt
	Size			string			//e.g. 1000204886016 alles in byte
	Free 			uint64			//%TODO

	MountPoint		string			//e.g. /tmp
	Status 			string 			//e.g. running
	Label 			string
	FSType 			string			//e.g. ext4, LVM2_member, swap
	UUID			string			//e.g. 8a7c2183-f8e4-4c20-8f5a-d7cb1c6196b6
	Owner 			string
	Group 			string
	Mode 			string			//e.g. brw-rw----

	/*Only Detectable in root Mode*/
	StartMB			string			//e.g. 10GB Für parted //%TODO
	EndMB			string

	PhysicalVolume	PhysicalVolume	// nil wenn mountpoint und visa versa

	Diskstat 		linuxproc.DiskStat

}


func (p *Partition) CreateLuksContainer(pwfile string) (LUKS, error){
	l := LUKS{Part:*p, KeyFile:pwfile}
	l.LuksFormat()
	l.LuksOpen()

	return l, nil
}


//Make sure, that the partitiontabel is not dos
func (p *Partition) SetLabel(s string) (error){
	h, err := p.GetHarddrive()
	if err != nil {
		log.Error(err.Error())
		return  err
	}

	cmd := sudo + " parted " + h.Path + " name " + p.Number + " " + s
	log.Info(cmd)

	parts := strings.Fields(cmd)
	head := parts[0]
	parts = parts[1:len(parts)]

	_, err = exec.Command(head, parts...).Output()
	if err != nil {
		log.Error(err.Error())
		return  err
	}


	return nil
}

func (p *Partition) GetAllContainer() (error) {//%TODO
	return nil
}

func (p *Partition) CreateFileSystem(fs string) (error) {
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

	cmd := sudo + " " + cmdfs + " " + p.Path
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

func (p *Partition) CreateNewPhysicalVolume()(error){
	if p.IsPhysicalVolume() == true {return errors.New("Contains already Physical Volume")}

	ph, err := CreateNewPhysicalVolumeFromString(p.Path, "")
	if err != nil {
		return err
	}
	p.PhysicalVolume = ph

	return nil
}

func (p *Partition) CreateNewPhysicalVolumeWithCMD(cmd string)(error){
	if p.IsPhysicalVolume() == true {return errors.New("Contains already Physical Volume")}

	ph, err := CreateNewPhysicalVolumeFromString(p.Path, cmd)
	if err != nil {
		return err
	}
	p.PhysicalVolume = ph

	return nil
}

func (p *Partition) GetHarddrive() (Harddrive, error){
	h := Harddrive{}

	array, err := GetAllHarddrives()
	if err != nil {
		log.Error(err.Error())
		return h, err
	}

	for _, a := range array{
		for _, x := range a.Part{
			if x.Path == p.Path{
				a.init()
				return a, nil
			}
		}
	}
	log.Warning("No Harddrive Found")
	return h, errors.New("No Harddrive Found")
}

func (p *Partition) GetPhysicalVolume() (error){
	a, err :=GetPhysicalVolumeByString(p.Path)
	if err != nil {
		return err
	}

	p.PhysicalVolume = a
	return nil
}

func (p *Partition) IsPhysicalVolume() (bool){
	_, err :=GetPhysicalVolumeByString(p.Path)
	if err != nil {
		return false
	}
	return true
}

func GetPartitionFromString(PartDir string) (Partition, error){
	p := Partition{}

	hArray, err := GetAllHarddrives()
	if err != nil {
		log.Error(err.Error())
		return p, err
	}

	for _, h := range hArray {
		h.GetAllPartitions()
		if (len(h.Part)) > 0 {
			for _, pp := range h.Part {
				if pp.Path == PartDir {
					p = pp
				}
			}
		}
	}

	if p.Name == ""{
		log.Warning("No Partition Found")
		return p, errors.New("No Partitions found")
	}

	return p, nil
}

func GetAllMountedPartitions() ([]Partition, error){
	var Array []Partition

	data, err := GetAllMountedPartitionsAsStrings()
	if err != nil {
		log.Error(err.Error())
		return Array, err
	}

	for _, d := range data {
		p := Partition{}
		p, err = GetPartitionFromString(d)
		if err != nil {
			log.Error(err.Error())
			return Array, err
		}
		Array = append(Array, p)
	}

	return Array, nil
}

func GetAllMountedPartitionsAsStrings() ([]string, error){
	var a []string

	data, err := exec.Command("grep", "^/dev/", "/proc/self/mounts").Output()
	if err != nil {
		log.Error(err.Error())
		return a, err
	}

	lines := strings.Split(string(data), "\n")

	for _, line := range lines {
		if line == ""{continue}

		fields := strings.Fields(line)

		t, _ := isStringValidPartition(fields[0])
		if t == true{
			a = append(a, fields[0])
		}

	}

	return a, nil
}

func GetAllPartitionsByString(HDDDir string) ([]Partition, error){
	var x = 0
	funcf := func(c rune) bool {
		if (c == '"'){
			x = x + 1
			x = x % 2
		}
		if (x == 0){ return unicode.IsSpace(c)}	 // " is nicht offen
		return false							 // ein " ist offen
	}

	Array := []Partition{}
	data, err := exec.Command("lsblk", "-P", "-o", lsblkcmd,HDDDir).Output()
	if err != nil {
		log.Error(err.Error())
		return Array, err
	}

	out := string(data)
	lines := strings.Split(out, "\n")

	for _, line := range lines {
		if line == "" {break;}
		fields := strings.FieldsFunc(line, funcf)

		if !(clean(fields[0],"NAME") == "NAME"){
			if (clean(fields[5],"TYPE") == "part") {//Partition
				e := fields[0][len(fields[0])-1:]

				a := linuxproc.DiskStat{Name:clean(fields[0], "NAME")}
				p := Partition{Number:e, Name:clean(fields[0], "NAME"), Path:"/dev/" + clean(fields[0], "NAME"), Diskstat:a,
					MountPoint:clean(fields[6], "MOUNTPOINT"), UUID:clean(fields[7], "UUID"), Owner:clean(fields[8], "OWNER"), Group:clean(fields[9], "GROUP"),
					Mode:clean(fields[10], "MODE"), Type:clean(fields[5], "TYPE"), Size:clean(fields[3], "SIZE"), Status:clean(fields[11], "STATE"),
					FSType:clean(fields[13], "FSTYPE"), Label:clean(fields[14], "LABEL")}

				Array = append(Array, p)
			}
		}
	}

	//fmt.Println(Array)
	return Array, nil
}
